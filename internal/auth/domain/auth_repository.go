package domain

import (
	"context"
	"github.com/google/uuid"
)

// AuthRepository defines our Authentication's storage logic.
type AuthRepository interface {
  // CreateEntity - First trys to create a new Entity, then tries to create the root Account for said Entity.
  CreateEntity(context.Context, Entity, Account)( EntityID, AccountID, error)
  // CreateAccount - Creates a new Sub Account. 
  CreateAccount(context.Context, Account)( AccountID, error)
  // RemoveEntityByID - <TEMP> Removes an entity via it's ID
  RemoveEntityByID(ctx context.Context, id string) error
  // RemoveEntity - <TEMP> Removes an entity by passing the name
  RemoveEntityByName(ctx context.Context, entityName string) error
  // RemoveAccountByID    - <TEMP> Removes an Account via it's ID
  RemoveAccountByID(ctx context.Context, id string) error
  // RemoveAccountByEmail - <TEMP> Removes an Account via it's Email
  RemoveAccountByEmail(ctx context.Context, email string) error
  // GetEntityIdByName - Query and returns an Entitys ID via it's name.
  GetEntityIdByName(context.Context, string)( EntityID, error )
  // GetAccountIdByName - Query and returns an Accounts ID via it's Email.
  GetAccountIdByEmail(context.Context, string)( AccountID, error )
  // AccountSignin - Attempts a Account Sign in event.
  AccountSignin(context.Context, Credentials)( *AuthToken, error)
  // RefreshToken - Refreshes an account's JWT Tokens.
  StoreRefreshToken(ctx context.Context, acc_id uuid.UUID, token Token) error
  // ValidateRefreshToken - Validates an Refresh Token by querying 'tokens' and detecting if expired.
  ValidateRefreshToken(ctx context.Context, acc_id uuid.UUID, token string) error
}
