package pvz

import (
	"context"
	"errors"
	"io"
	"log/slog"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"github.com/whitxowl/pvz.git/internal/domain"
	srvErr "github.com/whitxowl/pvz.git/internal/service/errors"
	"github.com/whitxowl/pvz.git/internal/service/pvz/mocks"
	storageErr "github.com/whitxowl/pvz.git/internal/storage/errors"
)

// ─── helpers ─────────────────────────────────────────────────────────────────

func newTestService(t *testing.T, pvzStorage *mocks.MockPVZStorage, recStorage *mocks.MockReceptionStorage) *Service {
	t.Helper()
	log := slog.New(slog.NewTextHandler(io.Discard, nil))
	return New(log, pvzStorage, recStorage)
}

func newUUID() string {
	return uuid.New().String()
}

// ─── CreatePVZ ───────────────────────────────────────────────────────────────

func TestCreatePVZ_Success(t *testing.T) {
	ctx := context.Background()
	pvzStorage := mocks.NewMockPVZStorage(t)
	recStorage := mocks.NewMockReceptionStorage(t)

	id := newUUID()

	now := time.Now()
	pvzStorage.EXPECT().
		CreatePVZ(ctx, mock.MatchedBy(func(p *domain.PVZ) bool {
			return p != nil &&
				p.ID == id &&
				p.City == domain.Moscow &&
				p.RegistrationDate == &now
		})).
		Return(nil)

	svc := newTestService(t, pvzStorage, recStorage)
	pvz, err := svc.CreatePVZ(ctx, id, &now, domain.Moscow)

	require.NoError(t, err)
	require.NotNil(t, pvz)
	assert.Equal(t, id, pvz.ID)
	assert.Equal(t, domain.Moscow, pvz.City)
	assert.Equal(t, &now, pvz.RegistrationDate)
}

func TestCreatePVZ_InvalidCity(t *testing.T) {
	ctx := context.Background()
	pvzStorage := mocks.NewMockPVZStorage(t)
	recStorage := mocks.NewMockReceptionStorage(t)

	svc := newTestService(t, pvzStorage, recStorage)
	pvz, err := svc.CreatePVZ(ctx, newUUID(), nil, domain.City("Новосибирск"))

	assert.Nil(t, pvz)
	assert.ErrorIs(t, err, srvErr.ErrInvalidCity)
}

func TestCreatePVZ_AlreadyExists(t *testing.T) {
	ctx := context.Background()
	pvzStorage := mocks.NewMockPVZStorage(t)
	recStorage := mocks.NewMockReceptionStorage(t)

	pvzStorage.EXPECT().
		CreatePVZ(ctx, mock.Anything).
		Return(storageErr.ErrPVZExists)

	svc := newTestService(t, pvzStorage, recStorage)
	pvz, err := svc.CreatePVZ(ctx, newUUID(), nil, domain.Moscow)

	assert.Nil(t, pvz)
	assert.ErrorIs(t, err, srvErr.ErrPVZExists)
}

func TestCreatePVZ_StorageError(t *testing.T) {
	ctx := context.Background()
	pvzStorage := mocks.NewMockPVZStorage(t)
	recStorage := mocks.NewMockReceptionStorage(t)
	unexpectedErr := errors.New("connection refused")

	pvzStorage.EXPECT().
		CreatePVZ(ctx, mock.Anything).
		Return(unexpectedErr)

	svc := newTestService(t, pvzStorage, recStorage)
	pvz, err := svc.CreatePVZ(ctx, newUUID(), nil, domain.Moscow)

	assert.Nil(t, pvz)
	assert.ErrorContains(t, err, "connection refused")
}

// ─── CloseReception ──────────────────────────────────────────────────────────

func TestCloseReception_Success(t *testing.T) {
	ctx := context.Background()
	pvzStorage := mocks.NewMockPVZStorage(t)
	recStorage := mocks.NewMockReceptionStorage(t)

	pvzId := newUUID()
	recId := newUUID()

	now := time.Now()
	returned := &domain.Reception{
		ID:     recId,
		PvzID:  pvzId,
		Date:   &now,
		Status: domain.StatusClosed,
	}

	recStorage.EXPECT().
		SetStatusClosed(ctx, pvzId).
		Return(returned, nil)

	svc := newTestService(t, pvzStorage, recStorage)
	reception, err := svc.CloseReception(ctx, pvzId)

	require.NoError(t, err)
	require.NotNil(t, reception)
	assert.Equal(t, recId, reception.ID)

	assert.Equal(t, domain.StatusInProgress, reception.Status)
}

func TestCloseReception_NoInProgressReception(t *testing.T) {
	ctx := context.Background()
	pvzStorage := mocks.NewMockPVZStorage(t)
	recStorage := mocks.NewMockReceptionStorage(t)

	id := newUUID()

	recStorage.EXPECT().
		SetStatusClosed(ctx, id).
		Return(nil, storageErr.ErrNoInProgressReception)

	svc := newTestService(t, pvzStorage, recStorage)
	reception, err := svc.CloseReception(ctx, id)

	assert.Nil(t, reception)
	assert.ErrorIs(t, err, srvErr.ErrNoInProgressReception)
}

func TestCloseReception_StorageError(t *testing.T) {
	ctx := context.Background()
	pvzStorage := mocks.NewMockPVZStorage(t)
	recStorage := mocks.NewMockReceptionStorage(t)
	unexpectedErr := errors.New("error")

	id := newUUID()

	recStorage.EXPECT().
		SetStatusClosed(ctx, id).
		Return(nil, unexpectedErr)

	svc := newTestService(t, pvzStorage, recStorage)
	reception, err := svc.CloseReception(ctx, id)

	assert.Nil(t, reception)
	assert.ErrorContains(t, err, "error")
}

// ─── DeleteLastProduct ───────────────────────────────────────────────────────

func TestDeleteLastProduct_Deleted(t *testing.T) {
	ctx := context.Background()
	pvzStorage := mocks.NewMockPVZStorage(t)
	recStorage := mocks.NewMockReceptionStorage(t)

	id := newUUID()

	recStorage.EXPECT().
		DeleteLastAddedProduct(ctx, id).
		Return(true, nil)

	svc := newTestService(t, pvzStorage, recStorage)
	deleted, err := svc.DeleteLastProduct(ctx, id)

	require.NoError(t, err)
	assert.True(t, deleted)
}

func TestDeleteLastProduct_NothingToDelete(t *testing.T) {
	ctx := context.Background()
	pvzStorage := mocks.NewMockPVZStorage(t)
	recStorage := mocks.NewMockReceptionStorage(t)

	id := newUUID()

	recStorage.EXPECT().
		DeleteLastAddedProduct(ctx, id).
		Return(false, nil)

	svc := newTestService(t, pvzStorage, recStorage)
	deleted, err := svc.DeleteLastProduct(ctx, id)

	require.NoError(t, err)
	assert.False(t, deleted)
}

func TestDeleteLastProduct_StorageError(t *testing.T) {
	ctx := context.Background()
	pvzStorage := mocks.NewMockPVZStorage(t)
	recStorage := mocks.NewMockReceptionStorage(t)
	unexpectedErr := errors.New("error")

	id := newUUID()

	recStorage.EXPECT().
		DeleteLastAddedProduct(ctx, id).
		Return(false, unexpectedErr)

	svc := newTestService(t, pvzStorage, recStorage)
	deleted, err := svc.DeleteLastProduct(ctx, id)

	assert.False(t, deleted)
	assert.ErrorContains(t, err, "error")
}

// ─── GetPVZList ──────────────────────────────────────────────────────────────

func TestGetPVZList_Success_WithReceptions(t *testing.T) {
	ctx := context.Background()
	pvzStorage := mocks.NewMockPVZStorage(t)
	recStorage := mocks.NewMockReceptionStorage(t)

	pvzId1, pvzId2 := newUUID(), newUUID()
	recId1, recId2 := newUUID(), newUUID()

	now := time.Now()
	pvzList := []*domain.PVZ{
		{ID: pvzId1, City: domain.Moscow, RegistrationDate: &now},
		{ID: pvzId2, City: domain.Kazan, RegistrationDate: &now},
	}

	receptionsByPVZ := map[string][]*domain.Reception{
		pvzId1: {
			{ID: recId1, PvzID: pvzId1, Status: domain.StatusClosed},
			{ID: recId2, PvzID: pvzId1, Status: domain.StatusInProgress},
		},
	}

	pvzStorage.EXPECT().
		GetPVZList(ctx, 1, 10).
		Return(pvzList, nil)

	recStorage.EXPECT().
		GetReceptionsByPVZIDs(ctx, []string{pvzId1, pvzId2}, (*time.Time)(nil), (*time.Time)(nil)).
		Return(receptionsByPVZ, nil)

	svc := newTestService(t, pvzStorage, recStorage)
	result, err := svc.GetPVZList(ctx, 1, 10, nil, nil)

	require.NoError(t, err)
	require.Len(t, result, 2)

	assert.Len(t, result[0].Receptions, 2)
	assert.Equal(t, recId1, result[0].Receptions[0].ID)
	assert.Equal(t, recId2, result[0].Receptions[1].ID)

	assert.Empty(t, result[1].Receptions)
}

func TestGetPVZList_Success_WithTimeFilter(t *testing.T) {
	ctx := context.Background()
	pvzStorage := mocks.NewMockPVZStorage(t)
	recStorage := mocks.NewMockReceptionStorage(t)

	start := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
	end := time.Date(2024, 12, 31, 0, 0, 0, 0, time.UTC)

	id := newUUID()

	pvzList := []*domain.PVZ{
		{ID: id, City: domain.SaintPetersburg},
	}

	pvzStorage.EXPECT().
		GetPVZList(ctx, 2, 5).
		Return(pvzList, nil)

	recStorage.EXPECT().
		GetReceptionsByPVZIDs(ctx, []string{id}, &start, &end).
		Return(map[string][]*domain.Reception{}, nil)

	svc := newTestService(t, pvzStorage, recStorage)
	result, err := svc.GetPVZList(ctx, 2, 5, &start, &end)

	require.NoError(t, err)
	require.Len(t, result, 1)
	assert.Empty(t, result[0].Receptions)
}

func TestGetPVZList_EmptyList(t *testing.T) {
	ctx := context.Background()
	pvzStorage := mocks.NewMockPVZStorage(t)
	recStorage := mocks.NewMockReceptionStorage(t)

	pvzStorage.EXPECT().
		GetPVZList(ctx, 1, 10).
		Return([]*domain.PVZ{}, nil)

	recStorage.EXPECT().
		GetReceptionsByPVZIDs(ctx, []string{}, (*time.Time)(nil), (*time.Time)(nil)).
		Return(map[string][]*domain.Reception{}, nil)

	svc := newTestService(t, pvzStorage, recStorage)
	result, err := svc.GetPVZList(ctx, 1, 10, nil, nil)

	require.NoError(t, err)
	assert.Empty(t, result)
}

func TestGetPVZList_PVZStorageError(t *testing.T) {
	ctx := context.Background()
	pvzStorage := mocks.NewMockPVZStorage(t)
	recStorage := mocks.NewMockReceptionStorage(t)
	unexpectedErr := errors.New("error")

	pvzStorage.EXPECT().
		GetPVZList(ctx, 1, 10).
		Return(nil, unexpectedErr)

	svc := newTestService(t, pvzStorage, recStorage)
	result, err := svc.GetPVZList(ctx, 1, 10, nil, nil)

	assert.Nil(t, result)
	assert.ErrorContains(t, err, "error")
}

func TestGetPVZList_ReceptionStorageError(t *testing.T) {
	ctx := context.Background()
	pvzStorage := mocks.NewMockPVZStorage(t)
	recStorage := mocks.NewMockReceptionStorage(t)
	unexpectedErr := errors.New("error")

	id := newUUID()

	pvzList := []*domain.PVZ{
		{ID: id, City: domain.Moscow},
	}

	pvzStorage.EXPECT().
		GetPVZList(ctx, 1, 10).
		Return(pvzList, nil)

	recStorage.EXPECT().
		GetReceptionsByPVZIDs(ctx, []string{id}, (*time.Time)(nil), (*time.Time)(nil)).
		Return(nil, unexpectedErr)

	svc := newTestService(t, pvzStorage, recStorage)
	result, err := svc.GetPVZList(ctx, 1, 10, nil, nil)

	assert.Nil(t, result)
	assert.ErrorContains(t, err, "error")
}
