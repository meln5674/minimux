package minimux

import (
	"context"
	"fmt"
	"io"
	"net/http"
)

type PreProcessor func(ctx context.Context, req *http.Request) (context.Context, func())

// CancelWhenDone is a PreProcessor that cancels the context when finished
var CancelWhenDone PreProcessor = func(ctx context.Context, req *http.Request) (context.Context, func()) {
	return context.WithCancel(ctx)
}

// LogPendingRequest returns a PreProcessor that logs the method, url, and agent of a request to
// the given writer
func LogPendingRequest(w io.Writer) PreProcessor {
	return func(ctx context.Context, req *http.Request) (context.Context, func()) {
		fmt.Fprintf(w, "%s %s %s\n", req.Method, req.URL, req.UserAgent())
		return ctx, func() {}
	}
}

// PreProcessorChain takes a sequence of PreProcessor and returns one which calls them in order,
// and returns a defered function which calls their defered functions in reverse order
func PreProcessorChain(chain ...PreProcessor) PreProcessor {
	return func(ctx context.Context, req *http.Request) (context.Context, func()) {
		fs := [](func()){}
		var f func()
		for _, next := range chain {
			ctx, f = next(ctx, req)
			if f != nil {
				fs = append(fs, f)
			}
		}
		return ctx, func() {
			for _, f := range fs {
				defer f()
			}
		}
	}
}
