package minimux

import (
	"context"
	"fmt"
	"io"
	"net/http"
)

const (
	// StatusPanic is the statusCode passed to a PostProcessor when recovering from a panicked Route
	StatusPanic = -1
	// StatusPreProcessPanic is the statusCode passed to a PostProcessor when recovering from a panicked PreProcessor
	StatusPreProcessPanic = -2
)

// A PreProcessor is a function that can be called before a request to mutate a context
// and provide an optional function to be defered until the end of the request

// A PostProcessor is a function which can handle the result of a request
type PostProcessor func(ctx context.Context, req *http.Request, statusCode int, err error)

// LogCompletedRequest returns a PostProcessor that logs the method, url, agent, status code,
// and fatal error of a request
func LogCompletedRequest(w io.Writer) PostProcessor {
	return func(ctx context.Context, req *http.Request, statusCode int, err error) {
		fmt.Fprintf(w, "%s %s %s %d %v\n", req.Method, req.URL, req.UserAgent(), statusCode, err)
	}
}
