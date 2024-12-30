package http

import (
	"context"
	"errors"
	"net/http"
	"strings"

	"github.com/TylerAldrich814/Schematix/internal/auth/domain"
)

// AuthMiddleware - Middleware for verifying JWT Token existance and validity.
func AuthMiddleware(next http.Handler) http.Handler {
  return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
    // -> Extract Authorization Header:
    authHeader := r.Header.Get("Authorization")
    if authHeader == "" {
      http.Error(w, 
        "missing authorization header", 
        http.StatusUnauthorized,
      )
      return
    }

    // -> Parse Authorization Header:
    tokenParts := strings.Split(authHeader, " ")
    if len(tokenParts) != 2 || tokenParts[0] != "Bearer" {
      http.Error(w, 
        "invalid authorization format", 
        http.StatusUnauthorized,
      )
      return
    }

    tokenString := tokenParts[1]

    claims, err := domain.VerifyToken(tokenString)
    if err != nil {
      if errors.Is(err, domain.ErrTokenExpired) {
        http.Error(w,
          "expired",
          http.StatusUnauthorized,
        )
      } else {
        http.Error(w,
          err.Error(),
          http.StatusUnauthorized,
        )
      }
      return
    }

    ctx := context.WithValue(r.Context(), "userID", claims.UserID)
    ctx  = context.WithValue(ctx, "entityID", claims.EntityID)
    ctx  = context.WithValue(ctx, "role", claims.Role)

    next.ServeHTTP(w, r.WithContext(ctx))
  })
}
