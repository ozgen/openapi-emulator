package openapi

import (
	"io"
	"net/http"
	"strings"
)

func HasRequiredBodyParam(spec *Spec, swaggerPath, method string) bool {
	op := findOperation(spec, swaggerPath, method)
	if op == nil || op.RequestBody == nil || op.RequestBody.Value == nil {
		return false
	}
	return op.RequestBody.Value.Required
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
