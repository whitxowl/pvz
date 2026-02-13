package product

import (
	"context"

	"github.com/gin-gonic/gin"
	"github.com/whitxowl/pvz.git/internal/api/middleware"
	"github.com/whitxowl/pvz.git/internal/domain"
)

type AuthService interface {
	ValidateToken(ctx context.Context, token string) (*domain.TokenClaims, error)
}

type ProductService interface {
	Add(ctx context.Context, productType domain.Type, pvzID string) (*domain.Product, error)
}

type Handler struct {
	authService    AuthService
	productService ProductService
}

func New(authService AuthService, productService ProductService) *Handler {
	return &Handler{
		authService:    authService,
		productService: productService,
	}
}

func (h *Handler) RegisterRoutes(router *gin.RouterGroup) {
	router.POST("/products",
		middleware.AuthMiddleware(h.authService),
		middleware.RequireRoles(domain.RoleEmployee),
		h.add,
	)
}
