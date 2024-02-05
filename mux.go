package minimux

import (
	"context"
	"fmt"
	"net/http"
)

// StringSet is a set of strings
type StringSet map[string]struct{}

// Has returns true if an element is in the set
func (s StringSet) Has(elem string) bool {
	_, ok := s[elem]
	return ok
}

// StringSetOf returns a set with the given elements
func StringSetOf(elems ...string) StringSet {
	s := StringSet{}
	for _, elem := range elems {
		s[elem] = struct{}{}
	}
	return s
}

type snoopingResponseWriter struct {
	inner      http.ResponseWriter
	statusCode *int
}

func (s snoopingResponseWriter) Header() http.Header {
	return s.inner.Header()
}

func (s snoopingResponseWriter) Write(b []byte) (int, error) {
	return s.inner.Write(b)
}

func (s snoopingResponseWriter) WriteHeader(statusCode int) {
	*s.statusCode = statusCode
	s.inner.WriteHeader(statusCode)
}

// Mux routes http requests to handlers
type Mux struct {
	// Routes is the set of potential handlers to consider, in the order to check them
	Routes []Route
	// DefaultHander is an optional handler to use if no routes match a request
	DefaultHandler Handler
	// PreProcess is an optional function to call before attempting to match any routes, and to
	// generate the context for the request, along with a function to defer to the end of the request.
	// PreProcess is intended for logging and other "transparent" operations.
	// If PreProcess is not specified, context.Background() is used
	PreProcess PreProcessor
	// PostProcess is an optional function to call with the result
	// PostProcess is intended for logging and other "transparent" operations.
	// PostProcess is only called if one of Routes or DefaultHandler is called.
	// If a handler panics, statusCode will be -1, and err will be either the panic'ed error,
	// or an error containing a string representation of the panic'ed value.
	PostProcess PostProcessor
}

// InnerMux wraps a Mux so that it implements minimux.Handler instead of net/http.Handler .
// The request path will not be modified, so the inner Route's must match the entire path
func InnerMux(m *Mux) Handler {
	return innerMux{Mux: m}
}

// InnerMuxWithPrefix wraps a Mux so that it implements minimux.Handler instead of net/http.Handler .
// The request path will be overwritten to the value of the specified path variable and that variable
// discarded, so the inner Route's must match only the suffix.
func InnerMuxWithPrefix(suffixVar string, m *Mux) Handler {
	return innerMux{Mux: m, suffixVar: suffixVar}
}

type innerMux struct {
	*Mux
	suffixVar string
}

// ServeHTTP implements Handler
func (m innerMux) ServeHTTP(ctx context.Context, w http.ResponseWriter, req *http.Request, pathVars map[string]string, formErr error) (err error) {
	if m.suffixVar != "" {
		req.URL.Path = pathVars[m.suffixVar]
		delete(pathVars, m.suffixVar)
	}
	var statusCode int

	// Set up a handler in case pre-processor panics
	preProcessorDone := false
	if m.PostProcess != nil {
		defer func() {
			if preProcessorDone {
				return
			}
			r := recover()
			if r != nil {
				w.WriteHeader(http.StatusInternalServerError)
				statusCode = StatusPreProcessPanic
				var ok bool
				err, ok = r.(error)
				if !ok {
					err = fmt.Errorf("%v", r)
				}
				m.PostProcess(ctx, req, statusCode, err)
			}
		}()
	}
	// Call the pre-processor, and defer the function it returns, if any
	if m.PreProcess != nil {
		var toDefer func()
		ctx, toDefer = m.PreProcess(ctx, req)
		if toDefer != nil {
			defer toDefer()
		}
	}
	// Disable the preprocess panic handler now that it has completed
	preProcessorDone = true

	// Set up the method not allowed handler, default handler, and post-processor
	snoopW := snoopingResponseWriter{inner: w, statusCode: &statusCode}
	found := false
	methodNotAllowed := false
	defer func() {
		r := recover()
		if r != nil {
			if statusCode == 0 {
				w.WriteHeader(http.StatusInternalServerError)
			}
			statusCode = StatusPanic
			var ok bool
			err, ok = r.(error)
			if !ok {
				err = fmt.Errorf("%v", r)
			}
			// The panicked part of the stack trace is only available within this block,
			// which means if the use wants to potentially handle the panic by displaying
			// the trace, e.g. logr.Logger.Error, this has to be called here, and we must
			// duplicate the call
			m.PostProcess(ctx, req, statusCode, err)
		} else {
			if methodNotAllowed {
				statusCode = http.StatusMethodNotAllowed
				w.WriteHeader(statusCode)
			} else if !found {
				if m.DefaultHandler == nil {
					return
				}
				err = m.DefaultHandler.ServeHTTP(ctx, snoopW, req, nil, nil)
			}
			if statusCode == 0 {
				statusCode = http.StatusOK
			}
			if m.PostProcess != nil {
				m.PostProcess(ctx, req, statusCode, err)
			}
		}
	}()

	// Find the first matching route and call it
	for _, r := range m.Routes {
		var values []string
		values, found, methodNotAllowed = r.Matches(req)
		if !found {
			continue
		}
		r.VarMap(values, pathVars)
		formErr := r.ParseFormIfNeeded(req)
		err = r.Handler.ServeHTTP(ctx, snoopW, req, pathVars, formErr)
		break
	}
	return
}

// ServeHTTP implements net/http.Handler
func (m *Mux) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	ctx := context.Background()
	innerMux{Mux: m}.ServeHTTP(ctx, w, req, map[string]string{}, nil)
}
