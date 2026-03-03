package pvz

import (
	"time"

	"github.com/whitxowl/pvz.git/internal/domain"
)

type CreateRequest struct {
	ID               string     `json:"id" binding:"required"`
	RegistrationDate *time.Time `json:"registrationDate" binding:"required"`
	City             string     `json:"city" binding:"required"`
}

type CreateResponse struct {
	ID               string     `json:"id"`
	RegistrationDate *time.Time `json:"registrationDate"`
	City             string     `json:"city"`
}

type ErrorResponse struct {
	Message string `json:"message"`
}

func ToCreateResponse(pvz *domain.PVZ) CreateResponse {
	return CreateResponse{
		ID:               pvz.ID,
		RegistrationDate: pvz.RegistrationDate,
		City:             string(pvz.City),
	}
}
