package main

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"os/signal"
	"time"

	"math/rand/v2"

	"github.com/gorilla/mux"

	log "github.com/sirupsen/logrus"

	AuthService "github.com/TylerAldrich814/Fidicus/internal/auth/application"
	AuthHTTP "github.com/TylerAldrich814/Fidicus/internal/auth/infrastructure/http"
	AuthRepo "github.com/TylerAldrich814/Fidicus/internal/auth/infrastructure/repository"
	"github.com/TylerAldrich814/Fidicus/internal/shared/config"
	"github.com/TylerAldrich814/Fidicus/internal/shared/role"
	"github.com/TylerAldrich814/Fidicus/internal/shared/users"
	"github.com/TylerAldrich814/Fidicus/internal/shared/jwt"
)

type SignupReq struct {
  Entity users.EntitySignupReq   `json:"entity"`
  Account users.AccountSignupReq `json:"account"`
}

var (
  Suc = "[✅] "
  Err = "[❌] "
  Pan = "[💥] "
)

func main(){
  ctx, cancel := signal.NotifyContext(
    context.Background(),
    os.Interrupt,
  )
  defer cancel()
  config.InitLogger()

  dbConfig, err := config.GetPgsqlConfig("-auth")
  if err != nil {
    log.Fatal("Failed to get Database Config: %s", err.Error())
  }
  authURI := dbConfig.GetPostgresURI()

  // Main HTTP Router
  r := mux.NewRouter()

  log.Warn("        ------  Test Setup  ------        ")
  authHTTP, err := StartAuthService(ctx, authURI, r)
  if err != nil {
    log.WithFields(log.Fields{
      "auth_uri": authURI,
      "error": err,
    }).Panic(Pan+"failed to start auth service")
    return
  }
  log.WithFields(log.Fields{
    "auth_uri": authURI,
  }).Info(Suc+"created auth http server")

  log.Warn("        ------  Entity&Admin Tests  ------        ")

  defer func(){
    if authHTTP == nil {
      log.Warn("authHTTP is nil")
    } else {
      authHTTP.Shutdown()
    }
  }()
  pad := rand.IntN(100000)

  entityName := fmt.Sprintf("Entity_%d", pad)

  signupReq := SignupReq{
    Entity: users.EntitySignupReq{
      Name : entityName,
    },
    Account: users.AccountSignupReq{
      Email           : fmt.Sprintf("Admin@Entity_%d.com", pad),
      Passw           : "SomeSuperStrongPassword1",
      FirstName       : "Admin",
      LastName        : "Admin",
      CellphoneNumber : "814-555-6669",
    },
  }

  // ->> Test Entity&Admin Signup:
  entityID, adminID, err := AuthEntityAdminSignup(
    ctx,
    r,
    signupReq,
  )
  if err != nil {
    log.WithFields(log.Fields{
      "error": err,
    }).Panic(Pan+"entity signup failed")
    return
  }

  log.WithFields(log.Fields{
    "entity": entityID,
    "admin": adminID,
  }).Info(Suc+"successfully created entity & admin")

  // ->> Admin account Sign-in:
  signinReq := users.AccountSigninReq{
    EntityName : signupReq.Entity.Name,
    Email      : signupReq.Account.Email,
    Passw      : signupReq.Account.Passw,
    Role       : role.AccessRoleEntity,
  }
  
  access, refresh, err := AccountSignin(
    ctx,
    r,
    signinReq,
  )
  if err != nil {
    log.WithFields(log.Fields{
      "error": err.Error(),
    }).Error("AccountSignin failed")
  }

  log.WithFields(log.Fields{
    "access": access.SignedToken[:10]+"..."+access.SignedToken[len(access.SignedToken)-10:],
    "a_exp": access.Expiration,
    "refresh": refresh.SignedToken[:10]+"..."+refresh.SignedToken[len(refresh.SignedToken)-10:],
    "r_exp": refresh.Expiration,
  }).Info(Suc+"successfully signed in")

  // Needed -- otherwise refreshed tokens are equal.. Maybe I should add a random number to help change final hash..?
  time.Sleep(time.Second)

  log.Warn("        ------  Token Refresh Test  ------        ")
  // ->> Test Refreshing Tokens:
  r_access, r_refresh, err := RefreshTokens(
    ctx,
    r,
    refresh.SignedToken,
  )
  if err != nil {
    log.WithFields(log.Fields{
      "error": err.Error(),
    }).Error(Err+"Failed to Refresh Tokens")
  }

  log.WithFields(log.Fields{
    "r_access": r_access.SignedToken[:10]+"..."+r_access.SignedToken[len(r_access.SignedToken)-10:],
    "a_exp": r_access.Expiration,
    "r_refresh": r_refresh.SignedToken[:10]+"..."+r_refresh.SignedToken[len(r_refresh.SignedToken)-10:],
    "r_exp": r_refresh.Expiration,
  }).Info(Suc+"successfully r_refreshed jwt tokens")


  log.Warn("        ------  Sub Account Tests  ------        ")
  // ->> Create SubAccount
  subAccountReq := users.AccountSignupReq{
    EntityID        : entityID,
    Email           : fmt.Sprintf("SubAccount@Entity_%d.com", pad),
    Passw           : "SomeExtemely6Secretive6Password6",
    Role            : role.AccessRoleAccount,
    FirstName       : "Tyler",
    LastName        : "Aldrich",
    CellphoneNumber : "814-431-0674",
  }

  subAccountID, subAccErr := SubAccountSignup(
    ctx, 
    r, 
    subAccountReq,
    r_access,
  )
  var subAccess jwt.Token

  if subAccErr != nil {
    log.WithFields(log.Fields{
      "error": subAccErr,
    }).Error(Err+"failed to create sub account.")
  } else {
    log.WithFields(log.Fields{
      "account_id": subAccountID,
    }).Info(Suc+"successfully created sub account")

    // ->> SubAccount Signin Request Body:
    subSigninReq := users.AccountSigninReq{
      EntityName : entityName,
      Email      : subAccountReq.Email,
      Passw      : subAccountReq.Passw,
      Role       : role.AccessRoleAccount,
    }

    // ->> SubAccount Signin:
    subAccess, _, err = AccountSignin(
      ctx,
      r,
      subSigninReq,
    )
    if err != nil {
      log.WithFields(log.Fields{
        "error": err.Error(),
      }).Error(Err+"failed to sign in sub account")
    } else {
      log.WithFields(log.Fields{
        "sub_access": subAccess.SignedToken[len(subAccess.SignedToken)-10:],
      }).Info(Suc+"successfully signed in sub account")
    }

    // ->>  Test if AccessRoleAccount can delete Entity:
    shouldFail := DropEntity(
      ctx,
      r,
      subAccess,
    )
    if shouldFail == nil {
      log.Panic(Pan+"a sub account shouldn't be able to remove an entity")
    } else {
      log.Info(Suc+"successully blocked unauthorized entity deletion")
    }

    // -->> Test if Account Creation by AccessRoleAccount is blocked:
    subsubAccountReq := users.AccountSignupReq{
      EntityID        : entityID,
      Email           : fmt.Sprintf("ShouldNotPass@Entity_%d.com", pad),
      Passw           : "Should6Not9Pass",
      Role            : role.AccessRoleAdmin,
      FirstName       : "Tyler",
      LastName        : "A",
      CellphoneNumber : "814-431-0674",
    }
    _, signupShouldFail := SubAccountSignup(
      ctx,
      r,
      subsubAccountReq,
      subAccess,
    )
    if signupShouldFail == nil {
      log.Panic(Pan+"AccessRoleAccount shouldn't be able to create an account")
    } else {
      log.Info(Suc+"successfully blocked AccessRoleAccount from making another account.")
    }

    if err := AccountSignout(ctx, r, subAccess); err != nil {
      log.Panic(Pan+"failed to signout subaccount.")
    } else {
      log.Info(Suc+"successfully signed out subaccount.")
    }
  }

  log.Warn("        ------  Signout Admin Account  ------        ")
  if shouldFail := AccountSignout(ctx, r, access); err != nil {
    log.Panic(Pan+"signed out account with invalid access token: " + shouldFail.Error())
  } else {
    log.Info(Suc+"failed to singout account with invalid access token:")
  }

  if err := AccountSignout(ctx, r, r_access); err != nil {
    log.Panic("failed to signout account with valid token: " + err.Error())
  }

  // Sign back in 
  r_access, r_refresh, err = AccountSignin(ctx, r, signinReq)
  if err != nil {
    log.Panic(Pan+"failed to sign admin account back in.. " + err.Error())
  } else {
    log.Info(Suc+"successfully signed admin account back in")
  }

  // ---------- Test Clean up Functions ---------- 
  log.Warn("        ------  Cleanup Tests  ------        ")
  //       ->> Drop Entity & AdminAccount <<-
  if err := DropEntity(
    ctx,
    r,
    r_access,
  ); err != nil {
    log.WithFields(log.Fields{
      "error": err.Error(),
    }).Error("Failed to Drop Entity")
  } else {
    log.WithFields(log.Fields{
      "name": signupReq.Entity.Name,
    }).Info(Suc+"successfully removed entity")
  }

  // ->> Drop Account(s):
  if err := DropAccount(
    ctx,
    r,
    r_access.SignedToken,
  ); err != nil {
    log.WithFields(log.Fields{
      "error": err.Error(),
    }).Error(Err+"failed to drop account")
  } else {
    log.WithFields(log.Fields{
      "email": signupReq.Account.Email,
    }).Info(Suc+"successfully removed account")
  }

  if subAccErr != nil {
    if err := DropAccount(
      ctx,
      r,
      subAccess.SignedToken,
    ); err != nil {
      log.WithFields(log.Fields{
        "error": err.Error(),
      }).Error(Err+"failed to drop account")
    } else {
    log.WithFields(log.Fields{
      "email": signupReq.Account.Email,
    }).Info(Suc+"successfully removed account")
    }
  }

  log.Info("Done!")
}

// StartAuthService - Tests the initialization of our Authentication Service + HTTP Endpoints
func StartAuthService(
  ctx context.Context,
  dsn string,
  r   *mux.Router,
)( *AuthHTTP.AuthHTTPHandler, error ){

  authRepo, err := AuthRepo.NewAuthRepo(
    ctx, dsn,
  )
  if err != nil {
    return nil, errors.New("Failed to start auth repo: " + err.Error())
  }

  authService     := AuthService.NewService(authRepo)
  authHTTPHandler := AuthHTTP.NewHttpHandler(authService)
  if err := authHTTPHandler.RegisterRoutes(r); err != nil {
    return nil, errors.New("Failed to initalize Auth HTTP Routes: " + err.Error())
  }

  return authHTTPHandler, nil
}

func AuthEntityAdminSignup(
  ctx       context.Context,
  r         *mux.Router,
  signupReq SignupReq,
)(users.EntityID, users.AccountID, error){
  body, err := json.Marshal(signupReq)
  if err != nil {
    return users.NilEntity(), 
           users.NilAccount(), 
           err
  }

  req, err := http.NewRequest(
    "POST",
    "/auth/signup_entity",
    bytes.NewBuffer(body),
  )
  if err != nil {
    log.Panic("Failed to create entity/account: " + err.Error())
    return users.NilEntity(), users.NilAccount(), err
  }

  rr := httptest.NewRecorder()
  r.ServeHTTP(rr, req)

  if rr.Code != http.StatusOK {
    panic(fmt.Sprintf(
      "Failed to make Signup Request: %d", rr.Code))
  }

  log.Print(rr.Body.String())

  var ids struct{
    EntityID  users.EntityID  `json:"entity_id"`
    AccountID users.AccountID `json:"account_id"`
  }
  if err := json.NewDecoder(rr.Body).Decode(&ids); err != nil {
    log.Panic("Failed to parse response: " + err.Error())
    return users.NilEntity(), users.NilAccount(), err
  }

  return ids.EntityID, ids.AccountID, nil
}

func SubAccountSignup(
  ctx       context.Context,
  r         *mux.Router,
  signupReq users.AccountSignupReq,
  atoken    jwt.Token,
)( users.AccountID, error ){
  var throwError = func(f string, args ...any)(users.AccountID, error){
    return users.NilAccount(), fmt.Errorf("SubAccountSignup: " + f, args...)
  }

  body, err := json.Marshal(signupReq)
  if err != nil {
    return users.NilAccount(), err
  }
  req, err := http.NewRequest(
    "POST",
    "/pauth/signup_account",
    bytes.NewBuffer(body),
  )
  req.Header.Set("Authorization", "Bearer "+atoken.SignedToken)

  rr := httptest.NewRecorder()
  r.ServeHTTP(rr, req)

  if rr.Code != http.StatusOK {
    return throwError("http status code is non-ok: %d", rr.Code)
  }

  var accountID struct{
    AccountID users.AccountID `json:"account_id"`
  }

  if err := json.NewDecoder(rr.Body).Decode(&accountID); err != nil {
    return throwError("failed to decode response body: %s", err.Error())
  }

  return users.NilAccount(), nil
}

func AccountSignin(
  ctx       context.Context,
  r         *mux.Router,
  signinReq users.AccountSigninReq,
)(jwt.Token, jwt.Token, error) {
  var throwError = func(f string, args ...any)(jwt.Token, jwt.Token, error){
    return jwt.Token{}, 
           jwt.Token{},
           fmt.Errorf("AccountSignin: " + f, args...)
  }

  body, err := json.Marshal(signinReq)
  if err != nil {
    return throwError("invalid sign in request: %s", err.Error())
  }

  req, err := http.NewRequest(
    "POST",
    "/auth/signin",
    bytes.NewBuffer(body),
  )
  if err != nil {
    return throwError("failed to create account signin req: " + err.Error())
  }

  rr := httptest.NewRecorder()
  r.ServeHTTP(rr, req)

  if rr.Code != http.StatusAccepted {
    return throwError("http status code is non-ok: %d - %s", rr.Code, rr.Body)
  }

  var tokens struct {
    AccessToken  jwt.Token `json:"access_token"`
    RefreshToken jwt.Token `json:"refresh_token"`
  }
  if err := json.NewDecoder(rr.Body).Decode(&tokens); err != nil {
    return throwError("failed to unmarshal resposne: %s", err.Error())
  }

  return tokens.AccessToken, tokens.RefreshToken, nil
}

func DropEntity(
  ctx     context.Context,
  r       *mux.Router,
  atoken  jwt.Token,
) error {
  var throwError = func(f string, args ...any) error {
    return fmt.Errorf("DropEntity: " + f, args...)
  }

  req, err := http.NewRequest(
    "POST", 
    "/pauth/remove_entity",
    nil,
  )
  if err != nil {
    return throwError("Failed to create Request: " + err.Error())
  }

  req.Header.Set("Authorization", "Bearer "+atoken.SignedToken)
  rr := httptest.NewRecorder()
  r.ServeHTTP(rr, req)

  if rr.Code != http.StatusOK {
    return throwError("htto status non-ok: %d", rr.Code)
  }
  
  return nil
}

func DropAccount(
  ctx    context.Context,
  r      *mux.Router,
  atoken string,
) error {
  var throwError = func(f string, args ...any) error {
    return fmt.Errorf("DropAccount: " + f, args...)
  }

  req, err := http.NewRequest(
    "POST",
    "/pauth/remove_account",
    nil,
  )
  if err != nil {
    return throwError("failed to Request: " + err.Error())
  }
  req.Header.Set("Authorization", "Bearer " + atoken)

  rr := httptest.NewRecorder()
  r.ServeHTTP(rr, req)

  if rr.Code != http.StatusOK {
    return throwError("htto status non-ok: %d", rr.Code)
  }

  return nil
}

func RefreshTokens(
  ctx     context.Context,
  r       *mux.Router,
  refresh string,
)( jwt.Token, jwt.Token, error ){
  var throwError = func(f string, args ...any)(jwt.Token, jwt.Token, error){
    return jwt.Token{}, 
           jwt.Token{}, 
           fmt.Errorf("RefreshTokens: "+f, args...)
  }

  body, err := json.Marshal(struct {
    RefreshToken string `json:"refresh_token"`
  }{ RefreshToken: refresh })
  if err != nil {
    return throwError("failed to marshal refresh token: %s", err.Error())
  }

  req, err := http.NewRequest(
    "POST",
    "/auth/refresh",
    bytes.NewBuffer(body),
  )
  if err != nil {
    return throwError("failed to create new Request: %s", err.Error())
  }

  rr := httptest.NewRecorder()
  r.ServeHTTP(rr, req)


  if rr.Code != http.StatusOK {
    return throwError("http status code is non-ok: %d", rr.Code)
  }

  var newTokens jwt.TokenResponse 

  if err := json.NewDecoder(rr.Body).Decode(&newTokens); err != nil {
    return throwError("failed to marshal response body: %s", err.Error())
  }

  return newTokens.AccessToken, newTokens.RefreshToken, nil
}

func AccountSignout(
  ctx     context.Context,
  r       *mux.Router,
  access  jwt.Token,
) error {
  var throwError = func(f string, args ...any) error {
    return fmt.Errorf("AccountSignout: " + f, args...)
  }

  req, err := http.NewRequest(
    "POST",
    "/pauth/signout",
    nil,
  )
  if err != nil {
    return throwError("failed to create request: %s", err.Error())
  }

  req.Header.Set("Authorization", "Bearer "+access.SignedToken)
  rr := httptest.NewRecorder()
  r.ServeHTTP(rr, req)

  if rr.Code != http.StatusOK {
    return throwError("http status non-ok: %s", rr.Code)
  }

  return nil
}
