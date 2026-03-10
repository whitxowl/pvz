package reception

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
	"github.com/whitxowl/pvz.git/internal/service/reception/mocks"
	storageErr "github.com/whitxowl/pvz.git/internal/storage/errors"
)

// ─── helpers ─────────────────────────────────────────────────────────────────

func newTestService(t *testing.T, recStorage *mocks.MockReceptionStorage, txManager *mocks.MockTxManager) *Service {
	t.Helper()
	log := slog.New(slog.NewTextHandler(io.Discard, nil))
	return New(log, recStorage, txManager)
}

func newUUID() string {
	return uuid.New().String()
}

func passThroughTx(txManager *mocks.MockTxManager) {
	txManager.EXPECT().
		WithTx(mock.Anything, mock.Anything).
		RunAndReturn(func(ctx context.Context, fn func(context.Context) error) error {
			return fn(ctx)
		})
}

// ─── CreateReception ─────────────────────────────────────────────────────────

func TestCreateReception_Success(t *testing.T) {
	ctx := context.Background()
	recStorage := mocks.NewMockReceptionStorage(t)
	txManager := mocks.NewMockTxManager(t)

	pvzId, recId := newUUID(), newUUID()

	now := time.Now()
	expected := &domain.Reception{
		ID:     recId,
		PvzID:  pvzId,
		Date:   &now,
		Status: domain.StatusInProgress,
	}

	recStorage.EXPECT().
		CreateReception(ctx, pvzId, domain.StatusInProgress).
		Return(expected, nil)

	svc := newTestService(t, recStorage, txManager)
	reception, err := svc.CreateReception(ctx, pvzId)

	require.NoError(t, err)
	require.NotNil(t, reception)
	assert.Equal(t, recId, reception.ID)
	assert.Equal(t, pvzId, reception.PvzID)
	assert.Equal(t, domain.StatusInProgress, reception.Status)
}

func TestCreateReception_AlreadyExists(t *testing.T) {
	ctx := context.Background()
	recStorage := mocks.NewMockReceptionStorage(t)
	txManager := mocks.NewMockTxManager(t)

	id := newUUID()

	recStorage.EXPECT().
		CreateReception(ctx, id, domain.StatusInProgress).
		Return(nil, storageErr.ErrInProgressReceptionExists)

	svc := newTestService(t, recStorage, txManager)
	reception, err := svc.CreateReception(ctx, id)

	assert.Nil(t, reception)
	assert.ErrorIs(t, err, srvErr.ErrInProgressReceptionExists)
}

func TestCreateReception_StorageError(t *testing.T) {
	ctx := context.Background()
	recStorage := mocks.NewMockReceptionStorage(t)
	txManager := mocks.NewMockTxManager(t)
	unexpectedErr := errors.New("error")

	id := newUUID()

	recStorage.EXPECT().
		CreateReception(ctx, id, domain.StatusInProgress).
		Return(nil, unexpectedErr)

	svc := newTestService(t, recStorage, txManager)
	reception, err := svc.CreateReception(ctx, id)

	assert.Nil(t, reception)
	assert.ErrorContains(t, err, "error")
}

// ─── Add ─────────────────────────────────────────────────────────────────────

func TestAdd_Success(t *testing.T) {
	ctx := context.Background()
	recStorage := mocks.NewMockReceptionStorage(t)
	txManager := mocks.NewMockTxManager(t)

	prodId, recId, pvzId := newUUID(), newUUID(), newUUID()

	now := time.Now()
	expectedProduct := &domain.Product{
		ID:          prodId,
		Date:        &now,
		Type:        domain.TypeElectronics,
		ReceptionID: recId,
	}

	passThroughTx(txManager)

	recStorage.EXPECT().
		GetReceptionInProgressID(mock.Anything, pvzId).
		Return(recId, nil)

	recStorage.EXPECT().
		CreateProduct(mock.Anything, domain.TypeElectronics, recId).
		Return(expectedProduct, nil)

	svc := newTestService(t, recStorage, txManager)
	product, err := svc.Add(ctx, domain.TypeElectronics, pvzId)

	require.NoError(t, err)
	require.NotNil(t, product)
	assert.Equal(t, prodId, product.ID)
	assert.Equal(t, domain.TypeElectronics, product.Type)
	assert.Equal(t, recId, product.ReceptionID)
}

func TestAdd_NoInProgressReception(t *testing.T) {
	ctx := context.Background()
	recStorage := mocks.NewMockReceptionStorage(t)
	txManager := mocks.NewMockTxManager(t)

	id := newUUID()

	passThroughTx(txManager)

	recStorage.EXPECT().
		GetReceptionInProgressID(mock.Anything, id).
		Return("", storageErr.ErrNoInProgressReception)

	svc := newTestService(t, recStorage, txManager)
	product, err := svc.Add(ctx, domain.TypeClothes, id)

	assert.Nil(t, product)
	assert.ErrorIs(t, err, srvErr.ErrNoInProgressReception)
}

func TestAdd_GetReceptionIDStorageError(t *testing.T) {
	ctx := context.Background()
	recStorage := mocks.NewMockReceptionStorage(t)
	txManager := mocks.NewMockTxManager(t)
	unexpectedErr := errors.New("error")

	id := newUUID()

	passThroughTx(txManager)

	recStorage.EXPECT().
		GetReceptionInProgressID(mock.Anything, id).
		Return("", unexpectedErr)

	svc := newTestService(t, recStorage, txManager)
	product, err := svc.Add(ctx, domain.TypeShoes, id)

	assert.Nil(t, product)
	assert.ErrorContains(t, err, "error")
}

func TestAdd_CreateProductStorageError(t *testing.T) {
	ctx := context.Background()
	recStorage := mocks.NewMockReceptionStorage(t)
	txManager := mocks.NewMockTxManager(t)
	unexpectedErr := errors.New("error")

	pvzId, recId := newUUID(), newUUID()

	passThroughTx(txManager)

	recStorage.EXPECT().
		GetReceptionInProgressID(mock.Anything, pvzId).
		Return(recId, nil)

	recStorage.EXPECT().
		CreateProduct(mock.Anything, domain.TypeElectronics, recId).
		Return(nil, unexpectedErr)

	svc := newTestService(t, recStorage, txManager)
	product, err := svc.Add(ctx, domain.TypeElectronics, pvzId)

	assert.Nil(t, product)
	assert.ErrorContains(t, err, "error")
}

func TestAdd_TxManagerError(t *testing.T) {
	ctx := context.Background()
	recStorage := mocks.NewMockReceptionStorage(t)
	txManager := mocks.NewMockTxManager(t)
	txErr := errors.New("error")

	txManager.EXPECT().
		WithTx(mock.Anything, mock.Anything).
		Return(txErr)

	svc := newTestService(t, recStorage, txManager)
	product, err := svc.Add(ctx, domain.TypeElectronics, newUUID())

	assert.Nil(t, product)
	assert.ErrorIs(t, err, txErr)
}
