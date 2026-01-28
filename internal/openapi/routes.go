package openapi

import (
	"fmt"
	"regexp"
	"strings"
)

type Route struct {
	Method     string
	Swagger    string
	Regex      *regexp.Regexp
	SampleFile string
}

// BuildRoutes creates a list of routes from the spec.
func BuildRoutes(spec *Spec) []Route {
	var out []Route
	for swaggerPath, methods := range spec.Paths {
		for m := range methods {
			method := strings.ToUpper(m)
			out = append(out, Route{
				Method:     method,
				Swagger:    swaggerPath,
				Regex:      swaggerPathToRegex(swaggerPath),
				SampleFile: swaggerPathToSampleName(method, swaggerPath),
			})
		}
	}
	return out
}

// FindRoute finds the first route matching method + path.
func FindRoute(routes []Route, method, path string) *Route {
	method = strings.ToUpper(method)
	for i := range routes {
		r := &routes[i]
		if r.Method == method && r.Regex.MatchString(path) {
			return r
		}
	}
	return nil
}

func swaggerPathToSampleName(method, swaggerPath string) string {
	s := strings.TrimPrefix(swaggerPath, "/")
	s = strings.ReplaceAll(s, "/", "_")
	return fmt.Sprintf("%s__%s.json", strings.ToUpper(method), s)
}

func swaggerPathToRegex(swaggerPath string) *regexp.Regexp {
	parts := strings.Split(swaggerPath, "/")
	var out []string

	for _, p := range parts {
		if p == "" {
			continue
		}
		if strings.HasPrefix(p, "{") && strings.HasSuffix(p, "}") {
			out = append(out, `([^/]+)`)
		} else {
			out = append(out, regexp.QuoteMeta(p))
		}
	}

	pat := "^/" + strings.Join(out, "/") + "/?$"
	return regexp.MustCompile(pat)
}
