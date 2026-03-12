package pvz

import (
	"errors"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/whitxowl/pvz.git/internal/domain"
	srvErr "github.com/whitxowl/pvz.git/internal/service/errors"
)

func (h *Handler) create(c *gin.Context) {
	var req CreateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Message: "invalid request",
		})
		return
	}

	pvz, err := h.pvzService.CreatePVZ(c, req.ID, req.RegistrationDate, domain.City(req.City))
	if err != nil {
		handleServiceError(c, err)
		return
	}

	response := ToCreateResponse(pvz)

	c.JSON(http.StatusOK, response)
}

func (h *Handler) closeLastReception(c *gin.Context) {
	pvzID := c.Param("pvzId")

	reception, err := h.pvzService.CloseReception(c, pvzID)
	if err != nil {
		handleServiceError(c, err)
		return
	}

	response := ToCloseReceptionResponse(reception)

	c.JSON(http.StatusOK, response)
}

func (h *Handler) deleteLastProduct(c *gin.Context) {
	pvzID := c.Param("pvzId")

	deleted, err := h.pvzService.DeleteLastProduct(c, pvzID)
	if err != nil {
		handleServiceError(c, err)
		return
	}

	if !deleted {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Message: "invalid request",
		})
	}

	c.Status(http.StatusOK)
}

func (h *Handler) get(c *gin.Context) {
	page := 1
	limit := 10

	if p := c.Query("page"); p != "" {
		if v, err := strconv.Atoi(p); err == nil && v >= 1 {
			page = v
		}
	}
	if l := c.Query("limit"); l != "" {
		if v, err := strconv.Atoi(l); err == nil && v >= 1 && v <= 30 {
			limit = v
		}
	}

	var startTime, endTime *time.Time
	if s := c.Query("startDate"); s != "" {
		t, err := time.Parse(time.RFC3339, s)
		if err != nil {
			c.JSON(http.StatusBadRequest, ErrorResponse{Message: "invalid startDate format"})
			return
		}
		startTime = &t
	}
	if e := c.Query("endDate"); e != "" {
		t, err := time.Parse(time.RFC3339, e)
		if err != nil {
			c.JSON(http.StatusBadRequest, ErrorResponse{Message: "invalid endDate format"})
			return
		}
		endTime = &t
	}

	pvzList, err := h.pvzService.GetPVZList(c.Request.Context(), page, limit, startTime, endTime)
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{Message: "internal server error"})
		return
	}

	response := make([]PVZResponse, len(pvzList))
	for i, pvz := range pvzList {
		response[i] = ToPVZResponse(pvz)
	}

	c.JSON(http.StatusOK, response)
}

func handleServiceError(c *gin.Context, err error) {
	switch {
	case errors.Is(err, srvErr.ErrPVZExists):
		c.JSON(http.StatusConflict, ErrorResponse{
			Message: "pvz already exists",
		})
	case errors.Is(err, srvErr.ErrInvalidCity):
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Message: "invalid city",
		})
	case errors.Is(err, srvErr.ErrNoInProgressReception):
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Message: "invalid request",
		})
	default:
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Message: "internal server error",
		})
	}
}
