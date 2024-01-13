package minimux

import (
	"net/http"
	"regexp"
)

// Route is a handler that accepts only certain requests
type Route struct {
	// Methods is an optional set of HTTP methods that will handle
	Methods StringSet
	// Hosts is an optional set of request hosts that this will handle
	Hosts StringSet
	// Pattern is the regular expression that matches URL routes that this will handle.
	// Each capture group represents a route variable.
	Pattern *regexp.Regexp
	// VarNames is the name of the route variables, in the order their capture groups appear in Pattern
	VarNames []string
	// HasForm indicates that ParseForm should be called for this handler
	HasForm bool
	// Handler is the actual handler logic
	Handler Handler
}

// LiteralPath starts building a handler for an exact route
func LiteralPath(path string) *Route {
	return &Route{Pattern: regexp.MustCompile("^" + regexp.QuoteMeta(path) + "$")}
}

// PathPattern starts building a handler for an route without any variables defined as a regular expression
func PathPattern(path string) *Route {
	return &Route{Pattern: regexp.MustCompile("^" + path + "$")}
}

// Route with vars starts building a handler for a route with variables defined as regular expression
// capture groups
func PathWithVars(pattern string, vars ...string) *Route {
	return &Route{Pattern: regexp.MustCompile("^" + pattern + "$"), VarNames: vars}
}

// WithMethods limits a handler to specific methods
func (r *Route) WithMethods(methods ...string) *Route {
	r.Methods = StringSetOf(methods...)
	return r
}

// WithMethods limits a handler to specific hosts
func (r *Route) WithHosts(hosts ...string) *Route {
	r.Hosts = StringSetOf(hosts...)
	return r
}

// WithForm sets a handler to indicate it needs the form data parsed
func (r *Route) WithForm(hosts ...string) *Route {
	r.HasForm = true
	return r
}

// IsHandledBy finishes building a handler by providing the serving logic
func (r *Route) IsHandledBy(handler Handler) Route {
	r.Handler = handler
	return *r
}

// IsHandledBy finishes building a handler by providing the serving logic
func (r *Route) IsHandledByFunc(handler HandlerFunc) Route {
	r.Handler = handler
	return *r
}

func (r *Route) Matches(req *http.Request) (varValues []string, matches bool, methodNotAllowed bool) {
	if r.Hosts != nil && !r.Hosts.Has(req.Host) {
		return nil, false, false
	}
	groups := r.Pattern.FindStringSubmatch(req.URL.Path)
	if groups == nil {
		return nil, false, false
	}
	if r.Methods != nil && !r.Methods.Has(req.Method) {
		return nil, false, true
	}
	return groups[1:], true, false
}

func (r *Route) VarMap(values []string, varMap map[string]string) {
	for ix, name := range r.VarNames {
		if ix >= len(values) {
			varMap[name] = ""
			continue
		}
		varMap[name] = values[ix]
	}
}

func (r *Route) ParseFormIfNeeded(req *http.Request) error {
	if !r.HasForm {
		return nil
	}
	return req.ParseForm()
}
