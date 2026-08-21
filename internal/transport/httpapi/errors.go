package httpapi

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/wyw14/cry-047/internal/domain"
)

func writeError(c *gin.Context, err error) {
	if errors.Is(err, domain.ErrVersionChanged) || errors.Is(err, domain.ErrDuplicate) {
		response := classifyConflict(err).withRequest(c.GetString("request_id"))
		c.JSON(response.httpStatus(), response.payload())
		return
	}
	status := http.StatusInternalServerError
	code := "internal_error"
	switch {
	case errors.Is(err, domain.ErrNotFound):
		status, code = http.StatusNotFound, "not_found"
	case errors.Is(err, domain.ErrConflict):
		status, code = http.StatusConflict, "conflict"
	case errors.Is(err, domain.ErrForbidden):
		status, code = http.StatusForbidden, "forbidden"
	case errors.Is(err, domain.ErrInvalidState):
		status, code = http.StatusUnprocessableEntity, "invalid_state"
	case errors.Is(err, domain.ErrEvidence):
		status, code = http.StatusUnprocessableEntity, "invalid_evidence"
	default:
		var validation *domain.ValidationError
		if errors.As(err, &validation) {
			status, code = http.StatusBadRequest, "validation_error"
		}
	}
	c.JSON(status, gin.H{"error": code, "message": err.Error()})
}

func bindJSON(c *gin.Context, target any) bool {
	if err := c.ShouldBindJSON(target); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid_json", "message": err.Error()})
		return false
	}
	return true
}
