package httpapi

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/wyw14/cry-047/internal/application"
	"github.com/wyw14/cry-047/internal/domain"
)

func (a *API) publishProgram(c *gin.Context) {
	var command domain.PublishProgram
	if !bindJSON(c, &command) {
		return
	}
	result, err := a.service.PublishProgram(c.Request.Context(), actorFrom(c), command)
	if err != nil {
		writeError(c, err)
		return
	}
	c.JSON(http.StatusCreated, result)
}

func (a *API) generateWindow(c *gin.Context) {
	var command domain.GenerateWindow
	if !bindJSON(c, &command) {
		return
	}
	result, err := a.service.GenerateWorkWindow(c.Request.Context(), actorFrom(c), command)
	if err != nil {
		writeError(c, err)
		return
	}
	c.JSON(http.StatusCreated, result)
}

func (a *API) reassignWindow(c *gin.Context) {
	var command domain.ReassignWindow
	if !bindJSON(c, &command) {
		return
	}
	command.WindowID = domain.ID(c.Param("id"))
	result, err := a.service.ReassignWorkWindow(c.Request.Context(), actorFrom(c), command)
	if err != nil {
		writeError(c, err)
		return
	}
	c.JSON(http.StatusOK, result)
}

func (a *API) skipWindow(c *gin.Context) {
	var command domain.SkipWindow
	if !bindJSON(c, &command) {
		return
	}
	command.WindowID = domain.ID(c.Param("id"))
	skipped, compensation, err := a.service.SkipWorkWindow(c.Request.Context(), actorFrom(c), command)
	if err != nil {
		writeError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"skipped": skipped, "compensation": compensation})
}

func (a *API) submitExecution(c *gin.Context) {
	var command domain.SubmitExecution
	if !bindJSON(c, &command) {
		return
	}
	result, err := a.service.SubmitExecution(c.Request.Context(), actorFrom(c), command)
	if err != nil {
		writeError(c, err)
		return
	}
	c.JSON(http.StatusCreated, result)
}

func (a *API) reviewExecution(c *gin.Context) {
	var command application.ReviewExecution
	if !bindJSON(c, &command) {
		return
	}
	command.ExecutionID = domain.ID(c.Param("id"))
	result, err := a.service.ReviewExecution(c.Request.Context(), actorFrom(c), command)
	if err != nil {
		writeError(c, err)
		return
	}
	c.JSON(http.StatusOK, result)
}

func (a *API) openIncident(c *gin.Context) {
	var command domain.OpenIncident
	if !bindJSON(c, &command) {
		return
	}
	result, err := a.service.OpenIncident(c.Request.Context(), actorFrom(c), command)
	if err != nil {
		writeError(c, err)
		return
	}
	c.JSON(http.StatusCreated, result)
}

func (a *API) advanceIncident(c *gin.Context) {
	var command application.AdvanceIncident
	if !bindJSON(c, &command) {
		return
	}
	command.IncidentID = domain.ID(c.Param("id"))
	result, err := a.service.AdvanceIncident(c.Request.Context(), actorFrom(c), command)
	if err != nil {
		writeError(c, err)
		return
	}
	c.JSON(http.StatusOK, result)
}
