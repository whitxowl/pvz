package pvz

import (
	"context"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/whitxowl/pvz.git/internal/api/middleware"
	"github.com/whitxowl/pvz.git/internal/domain"
)

type AuthService interface {
	ValidateToken(ctx context.Context, token string) (*domain.TokenClaims, error)
}

type PVZService interface {
	CreatePVZ(ctx context.Context, id string, registrationDate *time.Time, city domain.City) (*domain.PVZ, error)
	CloseReception(ctx context.Context, pvzID string) (*domain.Reception, error)
}

type Handler struct {
	authService AuthService
	pvzService  PVZService
}

func New(authService AuthService, pvzService PVZService) *Handler {
	return &Handler{
		authService: authService,
		pvzService:  pvzService,
	}
}

func (h *Handler) RegisterRoutes(router *gin.RouterGroup) {
	router.POST("/pvz",
		middleware.AuthMiddleware(h.authService),
		middleware.RequireRoles(domain.RoleModerator),
		h.create,
	)
	router.POST("/pvz/:pvzId/close_last_reception",
		middleware.AuthMiddleware(h.authService),
		middleware.RequireRoles(domain.RoleEmployee),
		h.closeLastReception,
	)
}
