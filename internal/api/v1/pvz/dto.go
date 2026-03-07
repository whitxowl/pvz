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

type CloseReceptionResponse struct {
	ID       string     `json:"id"`
	DateTime *time.Time `json:"dateTime"`
	PvzID    string     `json:"pvzId"`
	Status   string     `json:"status"`
}

type ProductResponse struct {
	ID          string     `json:"id"`
	DateTime    *time.Time `json:"dateTime"`
	ProductType string     `json:"type"`
	ReceptionID string     `json:"receptionId"`
}

type PVZInfo struct {
	ID               string     `json:"id"`
	RegistrationDate *time.Time `json:"registrationDate"`
	City             string     `json:"city"`
}

type ReceptionInfo struct {
	ID       string     `json:"id"`
	DateTime *time.Time `json:"dateTime"`
	PvzID    string     `json:"pvzId"`
	Status   string     `json:"status"`
}

type ReceptionWithProducts struct {
	Reception ReceptionInfo     `json:"reception"`
	Products  []ProductResponse `json:"products"`
}

type PVZResponse struct {
	PVZ        PVZInfo                 `json:"pvz"`
	Receptions []ReceptionWithProducts `json:"receptions"`
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

func ToCloseReceptionResponse(reception *domain.Reception) CloseReceptionResponse {
	return CloseReceptionResponse{
		ID:       reception.ID,
		DateTime: reception.Date,
		PvzID:    reception.PvzID,
		Status:   string(reception.Status),
	}
}

func ToPVZResponse(pvz *domain.PVZ) PVZResponse {
	receptions := make([]ReceptionWithProducts, len(pvz.Receptions))
	for i, r := range pvz.Receptions {
		products := make([]ProductResponse, len(r.Products))
		for j, p := range r.Products {
			products[j] = ProductResponse{
				ID:          p.ID,
				DateTime:    p.Date,
				ProductType: string(p.Type),
				ReceptionID: p.ReceptionID,
			}
		}
		receptions[i] = ReceptionWithProducts{
			Reception: ReceptionInfo{
				ID:       r.ID,
				DateTime: r.Date,
				PvzID:    r.PvzID,
				Status:   string(r.Status),
			},
			Products: products,
		}
	}
	return PVZResponse{
		PVZ: PVZInfo{
			ID:               pvz.ID,
			RegistrationDate: pvz.RegistrationDate,
			City:             string(pvz.City),
		},
		Receptions: receptions,
	}
}
