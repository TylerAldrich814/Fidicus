package application

import (
	"context"

	"github.com/TylerAldrich814/Schematix/internal/auth/domain"
)

type Service struct {
  repo domain.AuthRepository
}

func NewService(
  repo domain.AuthRepository,
) *Service {
  return &Service{ repo }
}

// CreateEntity - First attempts to create the Root account(i.e., the Entity)
// If Successful, we then call 'CreateSubAccount', creating the Entity's first Subaccount
// if a Admin level Role. And Finally, we return the EntityID and AccountID
func(s *Service) CreateEntity(
  ctx     context.Context,
  entity  domain.EntitySignupReq,
  account domain.AccountSignupReq,
)( domain.EntityID, domain.AccountID, error ){
  return s.repo.CreateEntity(ctx, entity, account)
}

// CreateSubAccount attempts to create a Sub Account under a specified Entity.
func(s *Service) CreateSubAccount(
  ctx      context.Context,
  account  domain.AccountSignupReq,
)( domain.AccountID, error) {
  return s.repo.CreateAccount(ctx, account)
}

func(s *Service) AccountSignin(
  ctx       context.Context,
  signInReq domain.AccountSigninReq,
)( domain.Token, domain.Token, error ) {
  // Call repo, attemp account signin. Returns AuthToken 
  access, refresh, err := s.repo.AccountSignin(ctx, signInReq)
  if err != nil {
    return domain.Token{}, domain.Token{}, err
  }

  return access, refresh, nil
}

// StoreRefreshToken - A Communication channel between AuthHTTPHandler and AuthRepository.
//
// After successfully creating a new RefreshToken. We call this method to upsert our newly
// genreated Refresh Token.
func(s *Service) StoreRefreshToken(
  ctx       context.Context,
  accountID domain.AccountID,
  token     domain.Token,
) error {
  return s.repo.StoreRefreshToken(
    ctx,
    accountID,
    token,
  )
}

// ValidateRefreshToken - Tests whether or not a provided JWT Refresh Token is both
// valid and not expired. Returning nil if the Refresh Token passes.
//
// Potential Errors:
//   - ErrDBFailedToQuery
//   - domain.ErrTokenExpired
func(s *Service) ValidateRefreshToken(
  ctx       context.Context,
  accountID domain.AccountID,
  token     string,
) error {
  if err := s.repo.ValidateRefreshToken(
    ctx,
    accountID,
    token,
  ); err != nil {
    // <TODO> Any Serive Middleware actions..?
    return err
  }

  return nil
}
