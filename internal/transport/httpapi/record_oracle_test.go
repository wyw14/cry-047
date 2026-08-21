package httpapi

import (
	"encoding/json"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/wyw14/cry-047/internal/domain"
	"net/http/httptest"
	"testing"
)

func TestVersionConflictIncludesCorrelationAndRetryContract(t *testing.T) {
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Set("request_id", "req-version-47")
	writeError(c, fmt.Errorf("stale epoch: %w", domain.ErrVersionChanged))
	var body map[string]any
	if e := json.Unmarshal(w.Body.Bytes(), &body); e != nil {
		t.Fatal(e)
	}
	if w.Code != 409 || body["error"] != "version_changed" || body["request_id"] != "req-version-47" || body["retryable"] != true {
		t.Fatalf("status=%d body=%s", w.Code, w.Body.String())
	}
}
