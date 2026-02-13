package product

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	srvErr "github.com/whitxowl/pvz.git/internal/service/errors"
)

func (h *Handler) add(c *gin.Context) {
	var req AddProductRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Message: "invalid request",
		})
		return
	}

	product, err := h.productService.Add(c, req.Type, req.PvzID)
	if err != nil {
		handleServiceError(c, err)
		return
	}

	response := ToAddProductResponse(product)

	c.JSON(http.StatusOK, response)
}

func handleServiceError(c *gin.Context, err error) {
	switch {
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
