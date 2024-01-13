# MiniMux

#### The littlest go HTTP framework.

MiniMux is a HTTP router that attempts to provide a minimal extension to the
otherwise excellent Go standard library package to address the few glaring deficiencies,
namely:
* Path variables
* Contexts
* Per-method routing
* Request logging
* Default routes
* Automatic form parsing
* Error and Panic handling

MiniMux has no external runtime dependencies, comes in at a few hundred lines, including comments,
not tens of thousands.

MiniMux extends the concept of `net/http.ServeMux`, with its own `minimux.Mux` type. Rather than `Handler`s, it is made up of `Route`s. A `Mux` can have a `DefaultRoute`, which is called if none of the other routes are matched. A `Mux` can also optionally include a `PreProcess` function which can inspect or mutate a request, as well as produce an appropriate context, and a `PostProcess` which can inspect not just a request, but the status code and any error returned by the matched `Route`.

If a `Route` panics, `PostProcess` will be called with the status `-1`. If the header has not been writen yet with `WriteHeader()`, a `500` error will be sent to the client. If `PreProcess` panics, a `500` status code will be sent to the client, and `-2` will be provided as the status code to `PostProcess`. In both cases, the error passed to `PostProcess` will be the result the panic. If the panicked value was an error, it will be passed as-is, otherwise, it will be converted using `fmt.Errorf("%#v")`. No attempt will be made to recover from a panicking `PostProcess`, or a deferred function from `PreProcess`.

An empty `Mux` will return `200` for all requests, similar to a `net/http.HandlerFunc` which does nothing.

A `Route` matches not a simple path or host/path, but a regular expression that can be used to extract path variables, multiple hosts, and specific methods. A `Route` can indicate if it intends to use form data, and the `Mux` will call `ParseForm()` for it. A `Route` accepts, in addition to the typical `ResponseWriter` and `Request` parameters, the path variables, in the form of a string map, as well as the error that was produced by `ParseForm()`. `Route`s can also return an error, though `Route`s should not use this in lieu of a typical 5XX status code, and only for situtations where it is already too late to report the error to the client, which the `Mux`'s `PostProcess` function will be able to report.

A `Route` is constructed using the builder pattern, starting with either `LiteralPath()` for exact paths without variables, `PathPattern()` for paths defined by regex without variables, or `PathWithVars()` to provide a regular expression and a set of variable names. It is undefined behavior to provide a pattern with a different number of catpure groups to variable names, but minimux will still attempt to process the request if it can. This builder can then use the methods `WithHosts()` and `WithMethods()` to filter down which requests it can handle, `WithForm()` to request that the form data be parsed for it, and then finally `IsHandledBy()` or `IsHandledByFunc()` to specify the handler logic. Existing `net/http.Handler`s and `net/http.HandlerFunc`s can be used by wrapping them with `Simple` and `SimpleFunc`, respectively.

Once `Route`s are constructed, a `Mux` is constructed as a plain-old-struct, with fields for `PreProcess`, `PostProcess`, and `DefaultHandler`, as well as a slice field `Routes` for the routes to attempt to match. Once constructed, a pointer to a `Mux` can be used anywhere a typical `net/http.Handler` would.


`Mux`s compose, using the `InnerMux` wrapper which implements `minimux.Handler` rather than `net/http.Handler`. Any path variables parsed by the outer `Mux`s routes will be inherited (and overwriten) by the inner one. Because `Route`s are considered sequentially, handling a request is `O(n)`, but using nested `Mux`s with prefixes reduces this significantly, and can be accomplished by giving the outer `Mux` a `Route` with a path pattern such as `/foo/.*`. If this suffix needs to be further inspected by the handler independent of the prefix, it can be used as a path variable, e.g. `minimux.PathWithVars("/foo(/.*)", "path")`.
