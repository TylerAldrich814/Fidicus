package repository

import (
	"context"
	log "github.com/sirupsen/logrus"
	"testing"

	"github.com/TylerAldrich814/Schematix/internal/auth/domain"
	"github.com/TylerAldrich814/Schematix/internal/shared/config"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

func setupTestDB(ctx context.Context) *PGRepo {
  config.InitLogger()

  db, err := New(ctx)
  if err != nil {
    panic(err)
  }

  return db
}

func TestEntitySignup(t *testing.T) {
  ctx := context.Background()
  db := setupTestDB(ctx)

  tests := []struct{
    name      string
    newAccount   domain.Account
    newEntity domain.Entity
    wantErr   error
  }{
    {
      name      : "Creates new entity and new user",
      newAccount   : domain.Account{
        Email           : "some_user@gmail.com",
        PasswHash       : "some_password",
        Role            : domain.AccessRoleAdmin,
        FirstName       : "Timmy",
        LastName        : "D.",
        CellphoneNumber : "814-555-0666",
      },
      newEntity : domain.Entity{
        Name            : "SomeEntity Inc",
        Description     : "Some Software Company",
        AccountIDs         : []uuid.UUID{},
      },
      wantErr   : nil,
    },
    {
      name      : "entity should already exist",
      newAccount   : domain.Account{
        Email           : "some_user@gmail.com",
        PasswHash       : "some_password",
        Role            : domain.AccessRoleAdmin,
        FirstName       : "Timmy",
        LastName        : "D.",
        CellphoneNumber : "814-555-0666",
      },
      newEntity : domain.Entity{
        Name            : "SomeEntity Inc",
        Description     : "Some Software Company",
        AccountIDs         : []uuid.UUID{},
      },
      wantErr   : ErrDBEntityAlreadyExists,
    },
  }

  for _, tt := range tests {
    t.Run(tt.name, func(t *testing.T){
      eid, uid, err := db.CreateEntity(
        ctx,
        tt.newEntity,
        tt.newAccount,
      )
      if eid != "" && uid != "" {
        log.Info("EID: " + eid)
        log.Info("UID: " + eid)
      }
      assert.Equal(t, tt.wantErr, err, tt.name)
    })
  }
}


func TestCreateAccount(t *testing.T) {
  ctx := context.Background()
  db := setupTestDB(ctx)

  testEID, _, err := db.CreateEntity(
    ctx,
    domain.Entity{
      Name: "SomeEntity",
    },
    domain.Account {
      Email     : "someAdminEmail@entity.com",
      FirstName : "Admin",
      LastName  : "Admin",
    },
  )
  if err != nil {
    panic(err)
  }

  tests := []struct{
    name      string
    newAccount   domain.Account
    wantErr   error
  }{
    {
      name      : "Creates new entity and new user",
      newAccount   : domain.Account{
        Email           : "some_user@gmail.com",
        PasswHash       : "some_password",
        Role            : domain.AccessRoleAdmin,
        FirstName       : "Timmy",
        LastName        : "D.",
        CellphoneNumber : "814-555-0666",
      },
      wantErr   : nil,
    },
    {
      name      : "entity should already exist",
      newAccount   : domain.Account{
        Email           : "some_user@gmail.com",
        PasswHash       : "some_other_password",
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
      uid, err := db.CreateAccount(ctx, testEID, tt.newAccount)
      if uid != "" {
        log.Info("UID: " + uid)
      }

      assert.Equal(t, tt.wantErr, err, tt.name)
    })
  }
}
