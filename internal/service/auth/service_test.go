package auth

import (
	"context"
	"errors"
	"io"
	"log/slog"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"github.com/whitxowl/pvz.git/internal/domain"
	"github.com/whitxowl/pvz.git/internal/service/auth/mocks"
	srvErr "github.com/whitxowl/pvz.git/internal/service/errors"
	storageErr "github.com/whitxowl/pvz.git/internal/storage/errors"
	"github.com/whitxowl/pvz.git/pkg/hash"
	"github.com/whitxowl/pvz.git/pkg/jwt"
)

func newTestService(t *testing.T, storage *mocks.MockAuthStorage) *Service {
	t.Helper()
	log := slog.New(slog.NewTextHandler(io.Discard, nil))
	tokenManager := jwt.NewTokenManager("test-secret-key", time.Hour)
	passHasher := hash.NewPasswordHasher()
	return New(log, storage, tokenManager, passHasher)
}

// ─── RegisterUser ────────────────────────────────────────────────────────────

func TestRegisterUser_Success(t *testing.T) {
	ctx := context.Background()
	storage := mocks.NewMockAuthStorage(t)

	storage.EXPECT().
		RegisterUser(ctx, mock_user()).
		RunAndReturn(func(_ context.Context, u *domain.User) (string, error) {
			assert.Equal(t, "test@example.com", u.Email)
			assert.Equal(t, domain.RoleEmployee, u.Role)
			assert.NotEmpty(t, u.PassHash)
			return "generated-uuid", nil
		})

	svc := newTestService(t, storage)
	user, err := svc.RegisterUser(ctx, "test@example.com", "password123", domain.RoleEmployee)

	require.NoError(t, err)
	require.NotNil(t, user)
	assert.Equal(t, "generated-uuid", user.ID)
	assert.Equal(t, "test@example.com", user.Email)
	assert.Equal(t, domain.RoleEmployee, user.Role)
}

func TestRegisterUser_InvalidRole(t *testing.T) {
	ctx := context.Background()
	storage := mocks.NewMockAuthStorage(t)

	svc := newTestService(t, storage)

	user, err := svc.RegisterUser(ctx, "test@example.com", "password123", domain.Role("invalid"))

	assert.Nil(t, user)
	assert.ErrorIs(t, err, srvErr.ErrInvalidRole)
}

func TestRegisterUser_UserAlreadyExists(t *testing.T) {
	ctx := context.Background()
	storage := mocks.NewMockAuthStorage(t)

	storage.EXPECT().
		RegisterUser(ctx, mock_user()).
		Return("", storageErr.ErrUserExists)

	svc := newTestService(t, storage)
	user, err := svc.RegisterUser(ctx, "test@example.com", "password123", domain.RoleEmployee)

	assert.Nil(t, user)
	assert.ErrorIs(t, err, srvErr.ErrUserExists)
}

func TestRegisterUser_StorageError(t *testing.T) {
	ctx := context.Background()
	storage := mocks.NewMockAuthStorage(t)
	unexpectedErr := errors.New("db connection lost")

	storage.EXPECT().
		RegisterUser(ctx, mock_user()).
		Return("", unexpectedErr)

	svc := newTestService(t, storage)
	user, err := svc.RegisterUser(ctx, "test@example.com", "password123", domain.RoleEmployee)

	assert.Nil(t, user)
	assert.ErrorContains(t, err, "db connection lost")
}

// ─── Login ───────────────────────────────────────────────────────────────────

func TestLogin_Success(t *testing.T) {
	ctx := context.Background()
	storage := mocks.NewMockAuthStorage(t)
	passHasher := hash.NewPasswordHasher()

	passHash, err := passHasher.Hash("secret")
	require.NoError(t, err)

	storage.EXPECT().
		User(ctx, "user@example.com").
		Return(&domain.User{
			ID:       "some-uuid",
			Email:    "user@example.com",
			PassHash: passHash,
			Role:     domain.RoleModerator,
		}, nil)

	svc := newTestService(t, storage)
	token, err := svc.Login(ctx, "user@example.com", "secret")

	require.NoError(t, err)
	assert.NotEmpty(t, token)
}

func TestLogin_UserNotFound(t *testing.T) {
	ctx := context.Background()
	storage := mocks.NewMockAuthStorage(t)

	storage.EXPECT().
		User(ctx, "ghost@example.com").
		Return(nil, storageErr.ErrUserNotFound)

	svc := newTestService(t, storage)
	token, err := svc.Login(ctx, "ghost@example.com", "password")

	assert.Empty(t, token)
	assert.ErrorIs(t, err, srvErr.ErrInvalidCredentials)
}

func TestLogin_WrongPassword(t *testing.T) {
	ctx := context.Background()
	storage := mocks.NewMockAuthStorage(t)
	passHasher := hash.NewPasswordHasher()

	passHash, err := passHasher.Hash("correct-password")
	require.NoError(t, err)

	storage.EXPECT().
		User(ctx, "user@example.com").
		Return(&domain.User{
			ID:       "some-uuid",
			Email:    "user@example.com",
			PassHash: passHash,
			Role:     domain.RoleEmployee,
		}, nil)

	svc := newTestService(t, storage)
	token, err := svc.Login(ctx, "user@example.com", "wrong-password")

	assert.Empty(t, token)
	assert.ErrorIs(t, err, srvErr.ErrInvalidCredentials)
}

func TestLogin_StorageError(t *testing.T) {
	ctx := context.Background()
	storage := mocks.NewMockAuthStorage(t)
	unexpectedErr := errors.New("timeout")

	storage.EXPECT().
		User(ctx, "user@example.com").
		Return(nil, unexpectedErr)

	svc := newTestService(t, storage)
	token, err := svc.Login(ctx, "user@example.com", "password")

	assert.Empty(t, token)
	assert.ErrorContains(t, err, "timeout")
}

// ─── ValidateToken ────────────────────────────────────────────────────────────

func TestValidateToken_Success(t *testing.T) {
	ctx := context.Background()
	storage := mocks.NewMockAuthStorage(t)

	svc := newTestService(t, storage)

	passHasher := hash.NewPasswordHasher()
	passHash, _ := passHasher.Hash("pass")

	storage.EXPECT().
		User(ctx, "u@example.com").
		Return(&domain.User{PassHash: passHash, Role: domain.RoleModerator}, nil)

	token, err := svc.Login(ctx, "u@example.com", "pass")
	require.NoError(t, err)

	claims, err := svc.ValidateToken(ctx, token)
	require.NoError(t, err)
	assert.Equal(t, domain.RoleModerator, claims.Role)
}

func TestValidateToken_InvalidToken(t *testing.T) {
	ctx := context.Background()
	storage := mocks.NewMockAuthStorage(t)
	svc := newTestService(t, storage)

	claims, err := svc.ValidateToken(ctx, "not.a.valid.token")

	assert.Nil(t, claims)
	assert.Error(t, err)
}

// ─── helpers ──────────────────────────────────────────────────────────────────

// mock_user returns a mock.MatchedBy matcher that checks only the fields
// the service controls (email, role) while ignoring the hashed password,
// which will differ on every call because bcrypt uses a random salt.
func mock_user() interface{} {
	return mock.MatchedBy(func(u *domain.User) bool {
		return u != nil &&
			u.Email == "test@example.com" &&
			u.Role == domain.RoleEmployee &&
			u.PassHash != ""
	})
}
