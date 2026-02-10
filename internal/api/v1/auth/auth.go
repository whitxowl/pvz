package auth

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	srvErr "github.com/whitxowl/pvz.git/internal/service/errors"
)

func (h *Handler) register(c *gin.Context) {
	var req RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Message: "invalid request",
		})
		return
	}

	user, err := h.svc.RegisterUser(c.Request.Context(), req.Email, req.Password, req.Role)
	if err != nil {
		handleServiceError(c, err)
		return
	}

	response := ToRegisterResponse(user)

	c.JSON(http.StatusCreated, response)
}

func (h *Handler) login(c *gin.Context) {
	var req LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Message: "invalid request",
		})
		return
	}

	token, err := h.svc.Login(c.Request.Context(), req.Email, req.Password)
	if err != nil {
		handleServiceError(c, err)
		return
	}

	c.JSON(http.StatusOK, token)
}

func handleServiceError(c *gin.Context, err error) {
	switch {
	case errors.Is(err, srvErr.ErrUserExists):
		c.JSON(http.StatusConflict, ErrorResponse{
			Message: "user already exists",
		})
	case errors.Is(err, srvErr.ErrInvalidRole):
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Message: "invalid role",
		})
	case errors.Is(err, srvErr.ErrInvalidCredentials):
		c.JSON(http.StatusUnauthorized, ErrorResponse{
			Message: "invalid credentials",
		})
	default:
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Message: "internal server error",
		})
	}
}
