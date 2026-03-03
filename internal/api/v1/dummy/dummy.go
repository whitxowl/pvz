package dummy

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/whitxowl/pvz.git/internal/domain"
	srvErr "github.com/whitxowl/pvz.git/internal/service/errors"
)

func (h *Handler) dummyLogin(c *gin.Context) {
	var req LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Message: "invalid request",
		})
		return
	}

	token, err := h.svc.Login(c, domain.Role(req.Role))
	if errors.Is(err, srvErr.ErrInvalidRole) {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Message: "invalid role",
		})
		return
	}
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Message: "internal server error",
		})
	}

	c.JSON(http.StatusOK, token)
}
