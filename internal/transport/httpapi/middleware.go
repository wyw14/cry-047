package httpapi

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/wyw14/cry-047/internal/domain"
	"go.uber.org/zap"
)

const actorKey = "authenticated_actor"

func actorMiddleware(logger *zap.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		id := strings.TrimSpace(c.GetHeader("X-Actor-ID"))
		role := domain.Role(strings.TrimSpace(c.GetHeader("X-Actor-Role")))
		name := strings.TrimSpace(c.GetHeader("X-Actor-Name"))
		if id == "" || !validRole(role) {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "缺少有效的操作人身份"})
			return
		}
		if name == "" {
			name = id
		}
		c.Set(actorKey, domain.Actor{ID: domain.ID(id), Name: name, Role: role})
		logger.Debug("request authenticated", zap.String("actor_id", id), zap.String("role", string(role)))
		c.Next()
	}
}

func validRole(role domain.Role) bool {
	switch role {
	case domain.RoleAdmin, domain.RolePlanner, domain.RoleMaintainer, domain.RoleReviewer, domain.RoleViewer:
		return true
	default:
		return false
	}
}

func actorFrom(c *gin.Context) domain.Actor {
	value, _ := c.Get(actorKey)
	actor, _ := value.(domain.Actor)
	return actor
}

func requestLogger(logger *zap.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Next()
		logger.Info("http request", zap.String("method", c.Request.Method), zap.String("path", c.FullPath()),
			zap.Int("status", c.Writer.Status()), zap.Int("size", c.Writer.Size()))
	}
}
