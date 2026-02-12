package reception

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

	reception, err := h.receptionService.CreateReception(c, req.PvzID)
	if err != nil {
		handleServiceError(c, err)
		return
	}

	response := ToCreateResponse(reception)

	c.JSON(http.StatusOK, response)
}

func handleServiceError(c *gin.Context, err error) {
	switch {
	case errors.Is(err, srvErr.ErrInProgressReceptionExists):
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Message: "invalid request",
		})
	default:
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Message: "internal server error",
		})
	}
}
