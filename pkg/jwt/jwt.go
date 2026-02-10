package jwt

import (
	"errors"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/whitxowl/pvz.git/internal/domain"
)

var (
	ErrInvalidToken = errors.New("invalid token")
	ErrTokenExpired = errors.New("token is expired")
)

type Claims struct {
	Role domain.Role
	jwt.RegisteredClaims
}

type TokenManager struct {
	secretKey           []byte
	accessTokenDuration time.Duration
}

func NewTokenManager(secretKey string, accessTokenDuration time.Duration) *TokenManager {
	return &TokenManager{
		secretKey:           []byte(secretKey),
		accessTokenDuration: accessTokenDuration,
	}
}

func (tm *TokenManager) GenerateToken(claims domain.TokenClaims) (string, error) {
	return tm.generateToken(claims, tm.accessTokenDuration)
}

func (tm *TokenManager) ValidateToken(tokenString string) (*domain.TokenClaims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return tm.secretKey, nil
	})

	if err != nil {
		if errors.Is(err, ErrTokenExpired) {
			return nil, ErrTokenExpired
		}
		return nil, ErrInvalidToken
	}

	claims, ok := token.Claims.(*Claims)
	if !ok || !token.Valid {
		return nil, ErrInvalidToken
	}

	return &domain.TokenClaims{Role: claims.Role}, nil
}

func (tm *TokenManager) generateToken(claims domain.TokenClaims, duration time.Duration) (string, error) {
	now := time.Now()
	jwtClaims := Claims{
		Role: claims.Role,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(now.Add(duration)),
			IssuedAt:  jwt.NewNumericDate(now),
			NotBefore: jwt.NewNumericDate(now),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwtClaims)
	return token.SignedString(tm.secretKey)
}
