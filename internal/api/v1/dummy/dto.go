package dummy

import "github.com/whitxowl/pvz.git/internal/domain"

type LoginRequest struct {
	Role domain.Role `json:"role" binding:"required"`
}

type ErrorResponse struct {
	Message string `json:"message"`
}
