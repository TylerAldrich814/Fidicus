package repository

import (
	"context"
	"testing"

	"github.com/google/uuid"
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
        Email           : "some_user@gmail.com",
        Passw           : "some_other_password",
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

  // Tests for Creating an Entity and recreating the same entity, which shouold fail.
  for _, tt := range tests {
    t.Run(tt.name, func(t *testing.T){
      eid, _, err := db.CreateEntity(
        ctx,
        tt.newEntity,
        tt.newAccount,
      )
      if eid != domain.EntityID(uuid.Nil) {
      }
      assert.Equal(t, tt.wantErr, err, tt.name)
    })
  }

  id, err := db.GetEntityIDByName(ctx, tests[0].newEntity.Name)
  if err != nil {
    panic(err)
  }
  
  // Delete test Entity
  if err := db.RemoveEntityByID(ctx, id); err != nil {
    panic(err)
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
  eid := uuid.UUID(testEID)

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
