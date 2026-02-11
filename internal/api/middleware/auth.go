package middleware

import (
	"context"
	"errors"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/whitxowl/pvz.git/internal/domain"
	"github.com/whitxowl/pvz.git/pkg/jwt"
)

const (
	authHeader     = "Authorization"
	userContextKey = "user"
)

type AuthService interface {
	ValidateToken(ctx context.Context, token string) (*domain.TokenClaims, error)
}

// AuthMiddleware checks presence and validity of JWT token
func AuthMiddleware(auth AuthService) gin.HandlerFunc {
	return func(c *gin.Context) {
		token, err := extractToken(c)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, ErrorResponse{
				Message: "unauthorized",
			})
			return
		}

		claims, err := auth.ValidateToken(c, token)
		if err != nil {
			message := "invalid token"

			if errors.Is(err, jwt.ErrTokenExpired) {
				message = "token has expired"
			}

			c.AbortWithStatusJSON(http.StatusUnauthorized, ErrorResponse{
				Message: message,
			})
			return
		}

		c.Set(userContextKey, claims)
		c.Next()
	}
}

// RequireRoles checks whether user has any role from roles
// Requires to use AuthMiddleware before, otherwise won't get claims
func RequireRoles(roles ...domain.Role) gin.HandlerFunc {
	return func(c *gin.Context) {
		claims, _ := c.Get(userContextKey)

		userClaims, ok := claims.(*domain.TokenClaims)
		if !ok {
			c.AbortWithStatusJSON(http.StatusInternalServerError, ErrorResponse{
				Message: "internal server error",
			})
			return
		}

		hasRole := false
		for _, role := range roles {
			if userClaims.Role == role {
				hasRole = true
				break
			}
		}

		if !hasRole {
			c.AbortWithStatusJSON(http.StatusForbidden, ErrorResponse{
				Message: "insufficient permissions",
			})
			return
		}

		c.Next()
	}
}

// extractToken gets JWT token from Authorization header
func extractToken(c *gin.Context) (string, error) {
	header := c.GetHeader(authHeader)
	if header == "" {
		return "", errors.New("authorization header is missing")
	}

	parts := strings.SplitN(header, " ", 2)
	if len(parts) != 2 || parts[0] != "Bearer" {
		return "", errors.New("invalid authorization header format")
	}

	return parts[1], nil
}
