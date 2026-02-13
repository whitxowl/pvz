package product

import (
	"time"

	"github.com/whitxowl/pvz.git/internal/domain"
)

type AddProductRequest struct {
	Type  domain.Type `json:"type" binding:"required"`
	PvzID string      `json:"pvzId" binding:"required"`
}

type AddProductResponse struct {
	ID          string      `json:"id"`
	DateTime    *time.Time  `json:"dateTime"`
	Type        domain.Type `json:"product_type"`
	ReceptionID string      `json:"receptionId"`
}

type ErrorResponse struct {
	Message string `json:"message"`
}

func ToAddProductResponse(product *domain.Product) AddProductResponse {
	return AddProductResponse{
		ID:          product.ID,
		DateTime:    product.Date,
		Type:        product.Type,
		ReceptionID: product.ReceptionID,
	}
}
