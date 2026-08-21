package httpapi

import (
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/wyw14/cry-047/internal/domain"
)

func (a *API) registerFacility(c *gin.Context) {
	var command domain.RegisterFacility
	if !bindJSON(c, &command) {
		return
	}
	result, err := a.service.RegisterFacility(c.Request.Context(), actorFrom(c), command)
	if err != nil {
		writeError(c, err)
		return
	}
	c.JSON(http.StatusCreated, result)
}

func (a *API) listFacilities(c *gin.Context) {
	offset, _ := strconv.Atoi(c.Query("offset"))
	limit, _ := strconv.Atoi(c.Query("limit"))
	result, err := a.service.Facilities(c.Request.Context(), actorFrom(c), c.Query("query"), domain.Page{Offset: offset, Limit: limit, Sort: c.Query("sort")})
	if err != nil {
		writeError(c, err)
		return
	}
	c.JSON(http.StatusOK, result)
}

func (a *API) changeFacilityState(c *gin.Context) {
	var request struct {
		State         domain.FacilityState `json:"state"`
		ExpectedEpoch int64                `json:"expected_epoch"`
	}
	if !bindJSON(c, &request) {
		return
	}
	result, err := a.service.ChangeFacilityState(c.Request.Context(), actorFrom(c), domain.ID(c.Param("id")), request.State, request.ExpectedEpoch)
	if err != nil {
		writeError(c, err)
		return
	}
	c.JSON(http.StatusOK, result)
}

func (a *API) timeline(c *gin.Context) {
	result, err := a.service.Timeline(c.Request.Context(), actorFrom(c), domain.ID(c.Param("id")))
	if err != nil {
		writeError(c, err)
		return
	}
	c.JSON(http.StatusOK, result)
}

func (a *API) exportArchive(c *gin.Context) {
	location, bundle, err := a.service.ExportFacilityArchive(c.Request.Context(), actorFrom(c), domain.ID(c.Param("id")))
	if err != nil {
		writeError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"location": location, "checksum": bundle.Checksum, "exported_at": bundle.ExportedAt})
}

func (a *API) riskBoard(c *gin.Context) {
	result, err := a.service.RiskBoard(c.Request.Context(), actorFrom(c), time.Now().UTC())
	if err != nil {
		writeError(c, err)
		return
	}
	c.JSON(http.StatusOK, result)
}

func (a *API) personalTasks(c *gin.Context) {
	result, err := a.service.PersonalTasks(c.Request.Context(), actorFrom(c), domain.ID(c.Param("id")))
	if err != nil {
		writeError(c, err)
		return
	}
	c.JSON(http.StatusOK, result)
}
