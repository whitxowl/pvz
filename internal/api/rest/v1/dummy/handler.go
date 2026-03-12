package dummy

import (
	"context"

	"github.com/gin-gonic/gin"
	"github.com/whitxowl/pvz.git/internal/domain"
)

type DummyLoginService interface {
	Login(ctx context.Context, role domain.Role) (string, error)
}

type Handler struct {
	svc DummyLoginService
}

func New(svc DummyLoginService) *Handler {
	return &Handler{
		svc: svc,
	}
}

func (h *Handler) RegisterRoutes(router *gin.RouterGroup) {
	router.POST("dummyLogin", h.dummyLogin)
}
