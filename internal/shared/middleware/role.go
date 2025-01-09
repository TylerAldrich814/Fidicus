package middleware

import (
	"net/http"
	"github.com/TylerAldrich814/Fidicus/internal/shared/role"
	"github.com/TylerAldrich814/Fidicus/internal/shared/jwt"
)

// RoleAuthMiddleware -- 2nd degree middleware under AuthMiddleware:
//
//  Compares the scores of Context.claims.Role vs the provided domain.Role
//  If Context.Claims.Role is smaller than the provided Role, we return an
//  http Error of StatusUnauthorized. Else if Context.Claims.Role is equal
//  or greater than the provided Role, we then continue with the Request.
func RoleAuthMiddleware(next http.Handler, role role.Role) http.Handler {
  return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
    if next == nil {
      http.Error(w, "internal server error: no handler", http.StatusInternalServerError)
      return
    }
    claims, ok := r.Context().Value(ClaimsKey).(*jwt.AuthClaims)
    if !ok {
      http.Error(w, "missing auth claims", http.StatusUnauthorized)
      return
    }

    if claims.Role.Score() < role.Score() {
      http.Error(w, "user role not authorized", http.StatusUnauthorized)
      return
    }

    next.ServeHTTP(w, r)
  }) 
}
