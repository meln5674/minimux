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
