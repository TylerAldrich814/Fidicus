package httnp

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/gorilla/mux"

	"github.com/TylerAldrich814/Fidicus/internal/auth/application"
	"github.com/TylerAldrich814/Fidicus/internal/auth/domain"
	"github.com/TylerAldrich814/Fidicus/internal/auth/infrastructure/http/middleware"
	repo "github.com/TylerAldrich814/Fidicus/internal/auth/infrastructure/repository"
	"github.com/TylerAldrich814/Fidicus/internal/shared/utils"
)

type AuthHTTPHandler struct {
  service *application.Service
}

func NewHttpHandler(
  service *application.Service,
) *AuthHTTPHandler{
  return &AuthHTTPHandler{ service }
}

// RegisterRoutes - Creates and Registers all of Fidicus's Authentication HTTP Routes
  // <TODO> :: Client-side File Serving for Authenticaion purposes(?) 
  // mux.Handle("/", http.FileServer(http.Dir("public")))

func(a *AuthHTTPHandler) RegisterRoutes(r *mux.Router) error {
  public := r.PathPrefix("/api/auth").Subrouter()
  public.HandleFunc(
    "/signup_entity",
    a.SignupEntity,
  ).Methods("POST")

  public.HandleFunc(
    "/signin",
    a.Signin,
  ).Methods("POST")

  public.HandleFunc(
    "/refresh",
    a.UpdateRefreshToken,
  ).Methods("POST")

  protected := r.PathPrefix("/api/pauth").Subrouter()
  protected.Use(middleware.AuthMiddleware)

  protected.Handle(
    "/signup_account",
    middleware.RoleAuthMiddleware(
      http.HandlerFunc(a.SignupSubAccount),
      domain.AccessRoleAdmin,
    ),
  ).Methods("POST")

  protected.Handle(
    "/remove_entity",
    middleware.RoleAuthMiddleware(
      http.HandlerFunc(a.RemoveEntity),
      domain.AccessRoleEntity,
    ),
  ).Methods("POST")

  protected.Handle(
    "/remove_account",
    middleware.RoleAuthMiddleware(
      http.HandlerFunc(a.RemoveSubAccount),
      domain.AccessRoleAdmin,
    ),
  ).Methods("POST")
  
  protected.HandleFunc(
    "/signout",
    a.Signout,
  ).Methods("POST")

  return nil
}

// SignupEntity: |PUBLIC| Allows for the creation of a new Entity + AccessRoleEntity Account.
// 
func(a *AuthHTTPHandler) SignupEntity(w http.ResponseWriter, r *http.Request) {
  var req struct {
    Entity  domain.EntitySignupReq
    Account domain.AccountSignupReq
  }

  // ->> Extract Entity and Account JSON Data:
  if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
    http.Error(w, "<json error>missing required fields", http.StatusBadRequest)
    return
  }
  // ->> Verify all required data was included in JSON Body:
  if req.Entity.Name   == "" || 
     req.Account.Email == "" || 
     req.Account.Passw == "" ||
     req.Account.Role  == domain.AccessRoleUnspecified {
       http.Error(w, "<json error>missing required fields", http.StatusBadRequest)
       return
     }

  eid, aid, err := a.service.CreateEntity(
    r.Context(),
    req.Entity,
    req.Account,
  )
  if err != nil {
    if errors.Is(err, repo.ErrDBEntityAlreadyExists) || 
       errors.Is(err, repo.ErrDBAccountAlreadyExists){
      http.Error(w, err.Error(), http.StatusNotAcceptable)
      return
    }
    http.Error(w, err.Error(), http.StatusInternalServerError)
    return
  }

  ids := struct {
    EntityID  domain.EntityID  `json:"entity_id"`
    AccountID domain.AccountID  `json:"account_id"`
  }{
    EntityID  : eid,
    AccountID : aid,
  }

  utils.WriteJson(
    w,
    http.StatusOK,
    ids,
  )
}


// signupSubAccount - Attempts to create a Subaccount for a specified Entity. 
func(a *AuthHTTPHandler) SignupSubAccount(w http.ResponseWriter, r *http.Request) {
  var account domain.AccountSignupReq

  if err := json.NewDecoder(r.Body).Decode(&account); err != nil {
    http.Error(w, "<json error>missing required fields", http.StatusBadRequest)
    return
  }

  if account.EntityID == domain.NilEntity() ||
     account.Email    == ""                 ||
     account.Passw    == "" {
       http.Error(w, "<json error>missing required fields", http.StatusBadRequest)
       return
     }

  aid, err := a.service.CreateSubAccount(
    r.Context(),
    account,
  )
  if err != nil {
    if errors.Is(err, repo.ErrDBAccountAlreadyExists) || 
       errors.Is(err, repo.ErrDBEntityNotFound) {
         http.Error(w, err.Error(), http.StatusNotAcceptable)
         return
       }
       http.Error(w, err.Error(), http.StatusInternalServerError)
       return
  }

  utils.WriteJson(
    w,
    http.StatusOK,
    struct {
      AccountID domain.AccountID `json:"account_id"`
    }{
      AccountID: aid,
    },
  )
}

// Signin - Handles User Signin Request.
func(a *AuthHTTPHandler) Signin(w http.ResponseWriter, r *http.Request) {
  var signinReq domain.AccountSigninReq

  if err := json.NewDecoder(r.Body).Decode(&signinReq); err != nil {
    http.Error(w, "failed to decode json body", http.StatusBadRequest)
  }

  if signinReq.EntityName == "" ||
     signinReq.Email      == "" ||
     signinReq.Passw      == "" ||
     signinReq.Role       == domain.AccessRoleUnspecified {
       http.Error(
         w, 
         fmt.Sprintf("<json error>missing required fields"), 
         http.StatusBadRequest)
       return
     }
  
  access, refresh, err := a.service.AccountSignin(
    r.Context(),
    signinReq,
  )
  if err != nil {
    if errors.Is(err, repo.ErrDBInvalidPassword) {
      http.Error(w, "Invalid Password", http.StatusNotAcceptable)
      return
    } 
    http.Error(w, err.Error(), http.StatusInternalServerError)
    return
  }

  if access.SignedToken == "" || refresh.SignedToken == "" {
    panic("Signin: FAILED TO CREATE JWT TOKENS")
  }

  var tokens = struct {
    AccessToken  domain.Token `json:"access_token"`
    RefreshToken domain.Token `json:"refresh_token"`
  }{
    AccessToken  : access,
    RefreshToken : refresh,
  }

  utils.WriteJson(w,
    http.StatusAccepted,
    tokens,
  )
}


// ValidateRefreshToken - Communicates to our Auth service to perfrom a RefreshToken event.
// Returns new JWT Tokens to the user and stores an updated RefreshToken in our Database.
func(a *AuthHTTPHandler) UpdateRefreshToken(w http.ResponseWriter, r *http.Request) {
  var req struct {
    RefreshToken string `json:"refresh_token"`
  }

  if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
    http.Error(w, "invalid request body", http.StatusBadRequest)
    return
  }

  if req.RefreshToken == "" {
    http.Error(w, "missing refresh token", http.StatusBadRequest)
    return
  }

  // ->> Validate and extract Claims from user provided Refresh Token.
  claims, err := domain.VerifyToken(req.RefreshToken)
  if err != nil {
    http.Error(w, "invalid or expired refresh token", http.StatusUnauthorized)
    return
  }

  ctx, cancel := context.WithTimeout(r.Context(), 10*time.Second)
  defer cancel()

  if err := a.service.ValidateRefreshToken(
    ctx,
    claims.AccountID,
    req.RefreshToken,
  ); err != nil {
    http.Error(w, "refresh token invalid or revoked", http.StatusUnauthorized)
    return
  }

  newAccessToken, newRefreshToken, err := a.service.CreateRefreshToken(
    ctx,
    claims.EntityID,
    claims.AccountID,
    claims.Role,
  )
  if err != nil {
    http.Error(w, "internal error", http.StatusInternalServerError)
    return
  }

  // ->> Return new JWT Tokens
  resp := domain.TokenResponse {
    AccessToken  : newAccessToken,
    RefreshToken : newRefreshToken,
  }

  utils.WriteJson(w,
    http.StatusOK,
    resp,
  )
}

// RemoveEntity - [PROTECTED] Communicates to our Auth service to perfrom a RemoveEntity event.
// Effectively blocking all Subaccounts from accessing this Entities data.
func(a *AuthHTTPHandler) RemoveEntity(w http.ResponseWriter, r *http.Request){
  claims, ok := r.Context().Value(middleware.ClaimsKey).(*domain.AuthClaims)
  if !ok {
    http.Error(w, "missing claims in context", http.StatusInternalServerError)
    return
  }

  // Remove Entity via ID
  if err := a.service.RemoveEntity(r.Context(), claims.EntityID); err != nil {
    http.Error(w, "failed to remove entity", http.StatusInternalServerError)
  }

  w.WriteHeader(http.StatusOK)
}

// RemoveSubAccount - [PROTECTED] Communicates to our Auth service to perfrom a RemoveSubAccount event.
// Effectively blocking access to this account.
func(a *AuthHTTPHandler) RemoveSubAccount(w http.ResponseWriter, r *http.Request){
  // Extract and Unmarshal EntityID and AccountID provded via AuthMiddleware:
  claims, ok := r.Context().Value(middleware.ClaimsKey).(*domain.AuthClaims)
  if !ok {
    http.Error(w, "missing claims in context", http.StatusUnauthorized)
    return
  }

  // Remove Account via ID
  if err := a.service.RemoveSubAccount(r.Context(), claims.AccountID); err != nil {
    http.Error(w, "failed to remove account", http.StatusInternalServerError)
  }

  w.WriteHeader(http.StatusOK)
}

// Signout - [PROTECTED] Communicates to our Auth service to perfrom an AccountSignout event.
// Effectively removing the user's Refresh Token from our Database.
func(a *AuthHTTPHandler) Signout(w http.ResponseWriter, r *http.Request) {
  claims, ok := r.Context().Value(middleware.ClaimsKey).(*domain.AuthClaims)
  if !ok {
    http.Error(w, "missing claims in context", http.StatusUnauthorized)
    return
  }

  if err := a.service.AccountSignout(r.Context(), claims.AccountID); err != nil {
    http.Error(w, "failed to signout", http.StatusInternalServerError)
  }

  w.WriteHeader(http.StatusOK)
}

// Shutdown - Allows for graceful shutdown 
func(a *AuthHTTPHandler) Shutdown(){
  a.service.Shutdown()
}
