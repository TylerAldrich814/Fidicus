package domain

import (
	"context"
	role "github.com/TylerAldrich814/Fidicus/internal/shared/domain"
)

// AuthRepository defines our Authentication's storage logic.
type AuthRepository interface {
  // CreateEntity - First trys to create a new Entity, then tries to create the root Account for said Entity.
  CreateEntity(context.Context, EntitySignupReq, AccountSignupReq)( EntityID, AccountID, error)
  // CreateAccount - Creates a new Sub Account. 
  CreateAccount(context.Context, AccountSignupReq)( AccountID, error)
  // RemoveEntityByID - <TEMP> Removes an entity via it's ID
  RemoveEntityByID(context.Context, EntityID) error
  // RemoveEntity - <TEMP> Removes an entity by passing the name
  RemoveEntityByName(ctx context.Context, entityName string) error
  // RemoveAccountByID    - <TEMP> Removes an Account via it's ID
  RemoveAccountByID(ctx context.Context, id AccountID) error
  // RemoveAccountByEmail - <TEMP> Removes an Account via it's Email
  RemoveAccountByEmail(ctx context.Context, email string) error
  // GetEntityIDByName - Query and returns an Entitys ID via it's name.
  GetEntityIDByName(context.Context, string)( EntityID, error )
  // GetAccountIDByName - Query and returns an Accounts ID via it's Email.
  GetAccountIDByEmail(context.Context, string)( AccountID, error )
  // AccountSignin - Attempts a Account Sign in event: returns AccessToken, RefreshToken, err
  AccountSignin(context.Context, AccountSigninReq)( Token, Token, error)
  // AccountSignout - Signs the user out of Shematix; removing their Access Token from the DB.
  AccountSignout(context.Context, AccountID) error
  // RefreshToken - Refreshes an account's JWT Tokens.
  StoreRefreshToken(ctx context.Context, acc_id AccountID, token Token) error
  // ValidateRefreshToken - Validates an Refresh Token by querying 'tokens' and detecting if expired.
  ValidateRefreshToken(ctx context.Context, acc_id AccountID, token string) error
  // CreateRefreshToken - For creating a new Access Token, requires an accountID to verify account validity.
  // RefreshToken is wrapped with this AuthRepository function in order to verify that the caller is 
  // a valid account holder.
  CreateRefreshToken(context.Context, EntityID, AccountID, role.Role)( Token, Token, error )

  // Shutdown - Allows for graceful shutdown operations.
  Shutdown() error
}
