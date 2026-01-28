package openapi

import (
	"io"
	"net/http"
	"strings"
)

func HasRequiredBodyParam(spec *Spec, swaggerPath, method string) bool {
	if spec == nil || spec.Paths == nil {
		return false
	}

	m := spec.Paths[swaggerPath]
	if m == nil {
		return false
	}

	opAny, ok := m[strings.ToLower(method)]
	if !ok || opAny == nil {
		return false
	}

	op, ok := opAny.(map[string]any)
	if !ok {
		return false
	}

	paramsAny, ok := op["parameters"]
	if !ok || paramsAny == nil {
		return false
	}

	params, ok := paramsAny.([]any)
	if !ok {
		return false
	}

	for _, pAny := range params {
		p, ok := pAny.(map[string]any)
		if !ok {
			continue
		}
		inVal, _ := p["in"].(string)
		if inVal != "body" {
			continue
		}
		req, _ := p["required"].(bool)
		if req {
			return true
		}
	}

	return false
}

func IsEmptyBody(r *http.Request) (bool, error) {
	if r.Body == nil {
		return true, nil
	}
	b, err := io.ReadAll(r.Body)
	if err != nil {
		return false, err
	}
	r.Body = io.NopCloser(strings.NewReader(string(b)))
	return len(strings.TrimSpace(string(b))) == 0, nil
}
