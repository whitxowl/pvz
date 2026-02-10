package auth

import "github.com/whitxowl/pvz.git/internal/domain"

type RegisterRequest struct {
	Email    string      `json:"email" binding:"required,email"`
	Password string      `json:"password" binding:"required,min=6"`
	Role     domain.Role `json:"role" binding:"required"`
}

type RegisterResponse struct {
	ID    string      `json:"id"`
	Email string      `json:"email"`
	Role  domain.Role `json:"role"`
}

type ErrorResponse struct {
	Message string `json:"message"`
}

func ToRegisterResponse(user *domain.User) RegisterResponse {
	return RegisterResponse{
		ID:    user.ID,
		Email: user.Email,
		Role:  user.Role,
	}
}
