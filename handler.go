package minimux

import (
	"context"
	"net/http"
)

// A Handler handles requests
type Handler interface {
	// ServeHTTP serves an http request, along with parsed route variables and the error
	// from parsing form data, if any.
	ServeHTTP(ctx context.Context, w http.ResponseWriter, req *http.Request, pathVars map[string]string, formErr error) error
}

// HandlerFunc wraps a function into a Handler
type HandlerFunc func(ctx context.Context, w http.ResponseWriter, req *http.Request, pathVars map[string]string, formErr error) error

// ServeHTTP implements Handler
func (f HandlerFunc) ServeHTTP(ctx context.Context, w http.ResponseWriter, req *http.Request, pathVars map[string]string, formErr error) error {
	return f(ctx, w, req, pathVars, formErr)
}

// Simple wraps a net/http.Handler to implement Handler by discarding the context, path variables
// and form error, and returning a nil error
func Simple(handler http.Handler) Handler {
	return simple{Handler: handler}
}

type simple struct {
	http.Handler
}

func (s simple) ServeHTTP(ctx context.Context, w http.ResponseWriter, req *http.Request, pathVars map[string]string, formErr error) error {
	s.Handler.ServeHTTP(w, req)
	return nil
}

// SimpleFunc wraps a net/http.HandlerFunc to implement Handler by discarding the context, path variables
// and form error, and returning a nil error
func SimpleFunc(handlerFunc http.HandlerFunc) Handler {
	return simpleFunc{HandlerFunc: handlerFunc}
}

type simpleFunc struct {
	http.HandlerFunc
}

func (s simpleFunc) ServeHTTP(ctx context.Context, w http.ResponseWriter, req *http.Request, pathVars map[string]string, formErr error) error {
	s.HandlerFunc(w, req)
	return nil
}

// NotFound is a handler that returns a 404 status and does nothing else
var NotFound Handler = HandlerFunc(func(ctx context.Context, w http.ResponseWriter, req *http.Request, pathVars map[string]string, formErr error) error {
	w.WriteHeader(http.StatusNotFound)
	return nil
})

// RedirectingTo returns a handler which will redirect to a URL with a specific status code
func RedirectingTo(url string, statusCode int) Handler {
	return HandlerFunc(func(ctx context.Context, w http.ResponseWriter, req *http.Request, pathVars map[string]string, formErr error) error {
		http.Redirect(w, req, url, statusCode)
		return nil
	})
}

// StaticBytes is static data to return
type StaticBytes struct {
	Data        []byte
	ContentType string
}

// ServeHTTP implements Handler
func (s StaticBytes) ServeHTTP(ctx context.Context, w http.ResponseWriter, req *http.Request, pathVars map[string]string, formErr error) error {
	w.Header().Add("Content-Type", s.ContentType)
	w.WriteHeader(http.StatusOK)
	_, err := w.Write(s.Data)
	return err
}

// StaticString is static data to return
type StaticString struct {
	Data        string
	ContentType string
}

// ServeHTTP implements Handler
func (s StaticString) ServeHTTP(ctx context.Context, w http.ResponseWriter, req *http.Request, pathVars map[string]string, formErr error) error {
	return StaticBytes{Data: []byte(s.Data), ContentType: s.ContentType}.ServeHTTP(ctx, w, req, pathVars, formErr)
}

// StaticData is a set of static strings and bytes which answers requests with the matching data.
// If there is no match, and Default is non-nil, it will be called, otherwise, the response will be untouched.
// If PathVar is non-empty, that path variable will be used as the map key instead of the entire URL path.
// If that variable is not present, it will act as if the path was not matched.
type StaticData struct {
	StaticBytes    map[string]StaticBytes
	DefaultHandler Handler
	PathVar        string
}

// ServeHTTP implements Handler
func (s StaticData) ServeHTTP(ctx context.Context, w http.ResponseWriter, req *http.Request, pathVars map[string]string, formErr error) error {
	ok := true
	key := req.URL.Path
	var byteData StaticBytes
	if s.PathVar != "" {
		key, ok = pathVars[s.PathVar]
	}
	if ok {
		byteData, ok = s.StaticBytes[key]
	}
	if ok {
		return byteData.ServeHTTP(ctx, w, req, pathVars, formErr)
	}
	if s.DefaultHandler != nil {
		return s.DefaultHandler.ServeHTTP(ctx, w, req, pathVars, formErr)
	}
	return nil
}
