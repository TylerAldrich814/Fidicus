package repository

import (
	"context"
	"errors"
	"testing"

	log "github.com/sirupsen/logrus"

	"github.com/TylerAldrich814/Schematix/internal/auth/domain"
	"github.com/TylerAldrich814/Schematix/internal/shared/config"
	"github.com/stretchr/testify/assert"
)


func setupTestDB(ctx context.Context) domain.AuthRepository {
  // ->> App Config
  config.LoadEnv()
  config.InitLogger()

  // ->> Auth Repository Initialization:
  log.Warn()
  dbConfig := config.GetDBConfig()
  dsn := dbConfig.GetPostgresURI()

  db, err := NewAuthRepo(ctx, dsn)
  if err != nil {
    panic(err)
  }

  return db
}

func TestEntitySignupAndRemoval(t *testing.T) {
  ctx := context.Background()
  db := setupTestDB(ctx)

  tests := []struct{
    name       string
    newAccount domain.AccountSignupReq
    newEntity  domain.EntitySignupReq
    wantErr    error
  }{
    {
      name       : "Creates new entity and new user",
      newAccount : domain.AccountSignupReq{
        Email           : "some_user@gmail.com",
        Passw           : "some_password",
        Role            : domain.AccessRoleAdmin,
        FirstName       : "Timmy",
        LastName        : "D.",
        CellphoneNumber : "814-555-0666",
      },
      newEntity : domain.EntitySignupReq{
        Name        : "SomeEntity Inc",
        Description : "Some Software Company",
      },
      wantErr   : nil,
    },
    {
      name      : "entity should already exist",
      newAccount   : domain.AccountSignupReq{
        Email           : "some_other_user@gmail.com",
        Passw           : "some_other_other_password",
        Role            : domain.AccessRoleAdmin,
        FirstName       : "Timmy",
        LastName        : "D.",
        CellphoneNumber : "814-555-0666",
      },
      newEntity : domain.EntitySignupReq{
        Name            : "SomeEntity Inc",
        Description     : "Some Software Company",
      },
      wantErr   : ErrDBEntityAlreadyExists,
    },
  }
  // Cleanup:
  cleanup := func(){
    id, err := db.GetEntityIDByName(ctx, tests[0].newEntity.Name)
    if err != nil {
      if !errors.Is(err, ErrDBEntityNotFound){
        panic(err)
      }
      return
    }
    if err := db.RemoveEntityByID(ctx, id); err != nil {
      panic(err)
    }
  }
  defer func(){
    if r := recover(); r != nil {
      cleanup()
      t.Fail()
    } else {
      cleanup()
    }
  }()

  // Tests for Creating an Entity and recreating the same entity, which shouold fail.
  for _, tt := range tests {
    t.Run(tt.name, func(t *testing.T){
      _, _, err := db.CreateEntity(
        ctx,
        tt.newEntity,
        tt.newAccount,
      )
      assert.Equal(t, tt.wantErr, err, tt.name)
    })
  }
}

func TestCreateAccountAndRemoval(t *testing.T) {
  ctx := context.Background()
  db := setupTestDB(ctx)

  testEID, _, err := db.CreateEntity(
    ctx,
    domain.EntitySignupReq{
      Name        : "SomeEntity",
      Description : "Some Software Company",
    },
    domain.AccountSignupReq {
      Email     : "someAdminEmail@entity.com",
      Passw     : "some_super_secure_password",
      FirstName : "Admin",
      LastName  : "Admin",
    },
  )
  if err != nil {
    panic(err)
  }
  eid := testEID

  tests := []struct{
    name       string
    newAccount domain.AccountSignupReq
    wantErr    error
  }{
    {
      name       : "Creates new entity and new user",
      newAccount : domain.AccountSignupReq{
        EntityID        : eid,
        Email           : "some_user@gmail.com",
        Passw           : "some_password",
        Role            : domain.AccessRoleAdmin,
        FirstName       : "Timmy",
        LastName        : "D.",
        CellphoneNumber : "814-555-0666",
      },
      wantErr   : nil,
    },
    {
      name       : "entity should already exist",
      newAccount : domain.AccountSignupReq{
        EntityID        : eid,
        Email           : "some_user@gmail.com",
        Passw           : "some_other_password",
        Role            : domain.AccessRoleAdmin,
        FirstName       : "T.",
        LastName        : "D.",
        CellphoneNumber : "814-666-0666",
      },
      wantErr   : ErrDBAccountAlreadyExists,
    },
  }

  for _, tt := range tests {
    t.Run(tt.name, func(t *testing.T) {
      _, err := db.CreateAccount(ctx, tt.newAccount)

      assert.Equal(t, tt.wantErr, err, tt.name)
    })
  }
   id, err := db.GetAccountIDByEmail(ctx, tests[0].newAccount.Email)
   if err != nil {
     panic(err)
   }

  // Remove Test User
  if err := db.RemoveAccountByID(ctx, id); err != nil {
    panic(err)
  }

  if err := db.RemoveEntityByID(ctx, testEID); err != nil {
    panic(err)
  }
}

// TestAccessTokenValidation - Tests the following 
// 
//  - Account Signup
//  - Account Signin -- Retreive JWT Token
//  - Validate Refresh Token
func TestAccessTokenValidation(t *testing.T){
  ctx := context.Background()
  db  := setupTestDB(ctx)

  test := struct{
    name       string
    newAccount domain.AccountSignupReq
    newEntity  domain.EntitySignupReq
    wantErr    error
  }{
    name       : "Creates new entity and new user",
    newAccount : domain.AccountSignupReq{
      Email           : "testing_jwt@gmail.com",
      Passw           : "some_password",
      Role            : domain.AccessRoleAdmin,
      FirstName       : "Timmy",
      LastName        : "D.",
      CellphoneNumber : "814-555-0666",
    },
    newEntity : domain.EntitySignupReq{
      Name        : "JWT Token Inc",
      Description : "Some Software Company",
    },
    wantErr   : nil,
  }

  log.Print(" -->> CREATED ENTITY && ACCOUNT")
  eid, aid, err := db.CreateEntity(ctx, test.newEntity, test.newAccount)
  if err != nil {
    log.WithFields(log.Fields{
      "error": err.Error(),
    }).Error("Failed to create entity")
    t.Fail()
  }

  log.Print(" -->> ACCOUNT SIGN IN")
  access, refresh, err := db.AccountSignin(ctx, domain.AccountSigninReq{
    EntityName : test.newEntity.Name,
    Email      : test.newAccount.Email,
    Passw      : test.newAccount.Passw,
  })
  if err != nil {
    log.WithFields(log.Fields{
      "error": err.Error(),
    }).Error("Failed to sign account in")
    t.Fail()
  }
  if access.SignedToken == "" || refresh.SignedToken == "" {
    log.Error("Failed to create signed tokens")
    t.Fail()
  }

  log.Print(" -->> ACCOUNT SIGNED IN")

  if err := db.ValidateRefreshToken(
    ctx,
    aid,
    refresh.SignedToken,
  ); err != nil {
    log.Error("Failed to validate refresh token")
    t.Fail()
  }

  if err := db.RemoveEntityByID(ctx, eid); err != nil {
    log.WithFields(log.Fields{
      "error": err.Error(),
    }).Error("Failed to remove Entity")
    t.Fail()
  }
  if err := db.RemoveAccountByID(ctx, aid); err != nil {
    log.WithFields(log.Fields{
      "error": err.Error(),
    }).Error("Failed to remove Account")

    t.Fail()
  }
}
