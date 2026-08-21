package httpapi

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/wyw14/cry-047/internal/application"
	"go.uber.org/zap"
)

type API struct {
	service *application.Service
	logger  *zap.Logger
}

func New(service *application.Service, logger *zap.Logger) *gin.Engine {
	if logger == nil {
		logger = zap.NewNop()
	}
	api := &API{service: service, logger: logger}
	engine := gin.New()
	engine.Use(gin.Recovery(), requestLogger(logger))
	engine.GET("/healthz", func(c *gin.Context) { c.JSON(http.StatusOK, gin.H{"status": "ok"}) })
	v1 := engine.Group("/api/v1", actorMiddleware(logger))
	v1.POST("/facilities", api.registerFacility)
	v1.GET("/facilities", api.listFacilities)
	v1.POST("/facilities/:id/state", api.changeFacilityState)
	v1.GET("/facilities/:id/timeline", api.timeline)
	v1.GET("/facilities/:id/archive", api.exportArchive)
	v1.POST("/programs", api.publishProgram)
	v1.POST("/work-windows", api.generateWindow)
	v1.POST("/work-windows/:id/reassign", api.reassignWindow)
	v1.POST("/work-windows/:id/skip", api.skipWindow)
	v1.POST("/executions", api.submitExecution)
	v1.POST("/executions/:id/review", api.reviewExecution)
	v1.POST("/incidents", api.openIncident)
	v1.POST("/incidents/:id/advance", api.advanceIncident)
	v1.GET("/risk-board", api.riskBoard)
	v1.GET("/people/:id/tasks", api.personalTasks)
	return engine
}
