package pvz

import (
	"errors"
	"net/http"

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
