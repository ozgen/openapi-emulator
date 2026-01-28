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

func BuildRoutes(spec *Spec) []Route {
	if spec == nil || spec.Doc3 == nil || spec.Doc3.Paths == nil {
		return nil
	}

	var out []Route
	for swaggerPath, item := range spec.Doc3.Paths.Map() {
		if item == nil {
			continue
		}

		for method := range item.Operations() {
			m := strings.ToUpper(method)
			out = append(out, Route{
				Method:     m,
				Swagger:    swaggerPath,
				Regex:      swaggerPathToRegex(swaggerPath),
				SampleFile: swaggerPathToSampleName(m, swaggerPath),
			})
		}
	}
	return out
}

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
