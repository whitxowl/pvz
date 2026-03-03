package reception

import (
	"time"

	"github.com/whitxowl/pvz.git/internal/domain"
)

type CreateRequest struct {
	PvzID string `json:"pvzId" binding:"required"`
}

type CreateResponse struct {
	ID       string     `json:"id"`
	DateTime *time.Time `json:"dateTime"`
	PvzID    string     `json:"pvzId"`
	Status   string     `json:"status"`
}

type AddProductRequest struct {
	Type  string `json:"type" binding:"required"`
	PvzID string `json:"pvzId" binding:"required"`
}

type AddProductResponse struct {
	ID          string     `json:"id"`
	DateTime    *time.Time `json:"dateTime"`
	Type        string     `json:"product_type"`
	ReceptionID string     `json:"receptionId"`
}

type ErrorResponse struct {
	Message string `json:"message"`
}

func ToCreateResponse(reception *domain.Reception) CreateResponse {
	return CreateResponse{
		ID:       reception.ID,
		DateTime: reception.Date,
		PvzID:    reception.PvzID,
		Status:   string(reception.Status),
	}
}

func ToAddProductResponse(product *domain.Product) AddProductResponse {
	return AddProductResponse{
		ID:          product.ID,
		DateTime:    product.Date,
		Type:        string(product.Type),
		ReceptionID: product.ReceptionID,
	}
}
