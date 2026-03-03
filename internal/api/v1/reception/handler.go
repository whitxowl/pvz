package reception

import (
	"context"

	"github.com/gin-gonic/gin"
	"github.com/whitxowl/pvz.git/internal/api/middleware"
	"github.com/whitxowl/pvz.git/internal/domain"
)

type AuthService interface {
	ValidateToken(ctx context.Context, token string) (*domain.TokenClaims, error)
}

type ReceptionService interface {
	CreateReception(ctx context.Context, pvzID string) (*domain.Reception, error)
	Add(ctx context.Context, productType domain.Type, pvzID string) (*domain.Product, error)
}

type Handler struct {
	authService      AuthService
	receptionService ReceptionService
}

func New(authService AuthService, receptionService ReceptionService) *Handler {
	return &Handler{
		authService:      authService,
		receptionService: receptionService,
	}
}

func (h *Handler) RegisterRoutes(router *gin.RouterGroup) {
	router.POST("/receptions",
		middleware.AuthMiddleware(h.authService),
		middleware.RequireRoles(domain.RoleEmployee),
		h.create,
	)
	router.POST("/products",
		middleware.AuthMiddleware(h.authService),
		middleware.RequireRoles(domain.RoleEmployee),
		h.add,
	)
}
