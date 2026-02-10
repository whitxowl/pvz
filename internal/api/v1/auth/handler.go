package auth

import (
	"context"

	"github.com/gin-gonic/gin"
	"github.com/whitxowl/pvz.git/internal/domain"
)

type AuthService interface {
	RegisterUser(ctx context.Context, email, password string, role domain.Role) (*domain.User, error)
}

type Handler struct {
	svc AuthService
}

func NewHandler(svc AuthService) *Handler {
	return &Handler{
		svc: svc,
	}
}

func (h *Handler) RegisterRoutes(router *gin.RouterGroup) {
	router.POST("/register", h.register)
}
