package httpapi

import (
	"errors"
	"github.com/wyw14/cry-047/internal/domain"
	"net/http"
)

type conflictResponse struct {
	status                   int
	code, message, requestID string
	retryable                bool
	details                  map[string]string
}

func classifyConflict(err error) conflictResponse {
	r := conflictResponse{status: http.StatusInternalServerError, code: "internal_error", message: err.Error(), details: map[string]string{}}
	switch {
	case errors.Is(err, domain.ErrVersionChanged):
		r.status = http.StatusConflict
		r.code = "version_changed"
		r.retryable = true
	case errors.Is(err, domain.ErrConflict):
		r.status = http.StatusConflict
		r.code = "conflict"
	case errors.Is(err, domain.ErrDuplicate):
		r.status = http.StatusConflict
		r.code = "duplicate_command"
		r.retryable = true
	}
	return r
}
func (r conflictResponse) withRequest(id string) conflictResponse { r.requestID = id; return r }
func (r conflictResponse) payload() map[string]any {
	return map[string]any{"error": r.code, "message": r.message}
}
func (r conflictResponse) isConflict() bool { return r.status == http.StatusConflict }
func (r conflictResponse) canRetry() bool   { return r.retryable }
func (r conflictResponse) correlated() bool { return r.requestID != "" }
func (r conflictResponse) detail(key, value string) conflictResponse {
	r.details[key] = value
	return r
}
func (r conflictResponse) valid() bool     { return r.status >= 400 && r.code != "" }
func (r conflictResponse) httpStatus() int { return r.status }
