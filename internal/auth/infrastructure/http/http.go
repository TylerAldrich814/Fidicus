package http

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"time"

  "github.com/gorilla/mux"

	"github.com/TylerAldrich814/Schematix/internal/auth/application"
	"github.com/TylerAldrich814/Schematix/internal/auth/domain"
	repo "github.com/TylerAldrich814/Schematix/internal/auth/infrastructure/repository"
	"github.com/TylerAldrich814/Schematix/internal/shared/utils"
)

type AuthHTTPHandler struct {
  service *application.Service
}

func NewHttpHandler(
  service *application.Service,
) *AuthHTTPHandler{
  return &AuthHTTPHandler{ service }
}

// RegisterRoutes - Creates and Registers all of Schematix's Authentication HTTP Routes
func(a *AuthHTTPHandler) RegisterRoutes(r *mux.Router) error {
  // <TODO> :: Client-side File Serving for Authenticaion purposes(?) 
  // mux.Handle("/", http.FileServer(http.Dir("public")))

  public := r.PathPrefix("/api/auth").Subrouter()
  public.HandleFunc(
    "/signup/{role}",
    a.Signup,
  ).Methods("POST")

  public.HandleFunc(
    "/signin",
    a.Signin,
  ).Methods("POST")

  public.HandleFunc(
    "/refresh",
    a.RefreshToken,
  ).Methods("POST")

  protected := r.PathPrefix("/api/protected").Subrouter()
  protected.Use(AuthMiddleware)

  // protected.HandleFunc(
  //   "/signout".
  //   a.Signout,
  // ).Methods("POST")

  return nil
}

func(a *AuthHTTPHandler) RegisterProtectedRoutes() error {

  return nil
}

// Signup -- A centralized HTTP Handler for Schematix Account Signups. Handles both 
// Entity and SubAccount signups. Which one is determined by the URL Query "role". 
// If role == "entity", then we will expect the provided JSON body to contain the data
// For both an Entity and SubAccount. If now, we return an error.
func(a *AuthHTTPHandler) Signup(w http.ResponseWriter, r *http.Request) {
  role := r.PathValue("role")

  switch role {
  case "entity":
    a.signupEntity(w, r)
  case "subaccount":
    a.signupSubAccount(w, r)
  default:
    http.Error(w, "invalid role", http.StatusBadRequest)
  }
}

// signupEntity - Attempts to create both a brand new Entity account and finally a new Admin-leveled 
// SubAccount.
func(a *AuthHTTPHandler) signupEntity(w http.ResponseWriter, r *http.Request) {
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

  utils.WriteJson(
    w,
    http.StatusOK,
    struct {
      EntityID  domain.EntityID  `json:"entity_id"`
      AccountID domain.AccountID `json:"account_id"`
    }{
      EntityID: eid,
      AccountID: aid,
    },
  )
}

// signupSubAccount - Attempts to create a Subaccount for a specified Entity. 
func(a *AuthHTTPHandler) signupSubAccount(w http.ResponseWriter, r *http.Request) {
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

  if signinReq.EntityName == "" ||
     signinReq.Email      == "" ||
     signinReq.Passw      == "" ||
     signinReq.Role       == domain.AccessRoleUnspecified {
       http.Error(w, "<json error>missing required fieds", http.StatusBadRequest)
       return
     }
  
  access, refresh, err := a.service.AccountSignin(
    r.Context(),
    signinReq,
  )
  if err != nil {
    if errors.Is(err, repo.ErrDBInvalidPassword) {
      http.Error(w, "Invalid Password", http.StatusNotAcceptable)
    } else {
      http.Error(w, err.Error(), http.StatusBadRequest)
    }
    return
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

func(a *AuthHTTPHandler) Signout(w http.ResponseWriter, r *http.Request) {


}

// ValidateRefreshToken - An HTTP URL for validating a refreshtoken.
func(a *AuthHTTPHandler) RefreshToken(w http.ResponseWriter, r *http.Request) {
  var req struct {
    RefreshToken string `json:"refresh_token"`
  }

  if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
    http.Error(w, "invalid request body", http.StatusBadRequest)
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

  // ->> Generate new JWT Tokens.
  newAccessToken, err := domain.GenerateAccessToken(claims.AccountID, claims.EntityID, claims.Role)
  if err != nil {
    http.Error(w, "failed to create access token", http.StatusInternalServerError)
    return
  }

  newRefreshToken, err := domain.GenerateRefreshToken(claims.AccountID, claims.EntityID, claims.Role)
  if err != nil {
    http.Error(w, "failed to create refrehs token", http.StatusInternalServerError)
    return
  }

  // ->> Store new Refresh Token
  if err := a.service.StoreRefreshToken(ctx, claims.AccountID, newRefreshToken); err != nil {
    http.Error(w, "failed to store refresh token", http.StatusInternalServerError)
    return
  }

  // ->> Return new JWT Tokens
  resp := map[string]string {
    "access_token"  : newAccessToken.SignedToken,
    "refresh_token" : newRefreshToken.SignedToken,
  }
  utils.WriteJson(w,
    http.StatusOK,
    resp,
  )
}
