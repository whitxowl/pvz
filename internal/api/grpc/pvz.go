package grpcsrv

import (
	"context"

	"github.com/whitxowl/pvz.git/internal/domain"
	"github.com/whitxowl/pvz.git/pkg/proto/pvz_v1"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type Service interface {
	GetAll(ctx context.Context) ([]*domain.PVZ, error)
}

type pvzHandler struct {
	pvz_v1.UnimplementedPVZServiceServer
	pvzService Service
}

func NewPVZHandler(pvzService Service) *pvzHandler {
	return &pvzHandler{pvzService: pvzService}
}

func (h *pvzHandler) GetPVZList(ctx context.Context, _ *pvz_v1.GetPVZListRequest) (*pvz_v1.GetPVZListResponse, error) {
	const op = "pvzHandler.GetPVZList"

	pvzs, err := h.pvzService.GetAll(ctx)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "%s: %v", op, err)
	}

	result := make([]*pvz_v1.PVZ, 0, len(pvzs))
	for _, p := range pvzs {
		result = append(result, &pvz_v1.PVZ{
			Id:               p.ID,
			RegistrationDate: timestamppb.New(*p.RegistrationDate),
			City:             string(p.City),
		})
	}

	return &pvz_v1.GetPVZListResponse{Pvzs: result}, nil
}
