package pvz

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
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

	pvz, err := h.pvzService.CreatePVZ(c, req.ID, req.RegistrationDate, req.City)
	if err != nil {
		handleServiceError(c, err)
		return
	}

	response := ToCreateResponse(pvz)

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
	default:
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Message: "internal server error",
		})
	}
}
