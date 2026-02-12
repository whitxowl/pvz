package reception

import (
	"time"

	"github.com/whitxowl/pvz.git/internal/domain"
)

type CreateRequest struct {
	PvzID string `json:"pvzId" binding:"required"`
}

type CreateResponse struct {
	ID       string        `json:"id"`
	DateTime *time.Time    `json:"dateTime"`
	PvzID    string        `json:"pvzId"`
	Status   domain.Status `json:"status"`
}

type ErrorResponse struct {
	Message string `json:"message"`
}

func ToCreateResponse(reception *domain.Reception) CreateResponse {
	return CreateResponse{
		ID:       reception.ID,
		DateTime: reception.Date,
		PvzID:    reception.PvzID,
		Status:   reception.Status,
	}
}
