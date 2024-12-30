package http

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/TylerAldrich814/Schematix/internal/auth/application"
	"github.com/TylerAldrich814/Schematix/internal/auth/domain"
	repo "github.com/TylerAldrich814/Schematix/internal/auth/infrastructure/repository"
	"github.com/TylerAldrich814/Schematix/internal/shared/utils"
	"github.com/google/uuid"
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
func(a *AuthHTTPHandler) RegisterRoutes( mux *http.ServeMux,
) error {
  // <TODO> :: Client-side File Serving for Authenticaion purposes(?) 
  // mux.Handle("/", http.FileServer(http.Dir("public")))

  mux.HandleFunc(
    "POST /api/auth/signup/{role}",
    a.Signup,
  )
  mux.HandleFunc(
    "POST /api/auth/signin",
    a.Signin,
  )
  mux.HandleFunc(
    "POST /api/auth/validate_refresh_token/",
    a.ValidateRefreshToken,
  )

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
  if req.Entity.Name == ""   || 
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

  if account.EntityID == uuid.Nil ||
     account.Email    == ""       ||
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
  
  tokens, err := a.service.AccountSignin(
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
  utils.WriteJson(w,
    http.StatusAccepted,
    tokens,
  )
}

// ValidateRefreshToken - An HTTP URL for validating a refreshtoken.
func(a *AuthHTTPHandler) ValidateRefreshToken(w http.ResponseWriter, r *http.Request) {
  var req struct {
    AccountID    domain.AccountID `json:"account_id"`,
    RefreshToken string           `json:"refresh_token"`,
  }

  err := a.service.ValidateRefreshToken(r.Context(),)
}


