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

// Shutdown - Allows for graceful shutdown operations.
func(s *Service) Shutdown(){
  if s.repo == nil {
    return 
  }
  s.repo.Shutdown()
}

// CreateEntity - First attempts to create An Entity Account If Successful,
// We then call create an Account with AccessRoleEntity privileges. When
// both an Entity and Account are created, 
// we return the EntityID and AccountID
func(s *Service) CreateEntity(
  ctx     context.Context,
  entity  domain.EntitySignupReq,
  account domain.AccountSignupReq,
)( domain.EntityID, domain.AccountID, error ){
  return s.repo.CreateEntity(ctx, entity, account)
}

// RemoveEntity: <TODO> For the time being, this calls repo.RemoveEntityByID and completely
// wipes Entity from our DB. In the future, this will only disable  entity and
// all of Entity's SubAccunts from accessing Schematix.
func(s *Service) RemoveEntity(
  ctx      context.Context,
  entityID domain.EntityID,
) error {
  return s.repo.RemoveEntityByID(ctx, entityID)
}

// CreateSubAccount attempts to create a Sub Account under a specified Entity.
func(s *Service) CreateSubAccount(
  ctx      context.Context,
  account  domain.AccountSignupReq,
)( domain.AccountID, error) {
  return s.repo.CreateAccount(ctx, account)
}

// RemoveSubAccount - <TODO> For the time beign, this calls repo.RemoveAccountByID and completely 
// wipes Account from our DB.In the future, this will only disable account without removing all data.
func(s *Service) RemoveSubAccount(
  ctx       context.Context,
  accountID domain.AccountID,
) error {
  return s.repo.RemoveAccountByID(ctx, accountID)
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

// AccountSignout - Communicates to our Repository to perform an AccountSignout event.
// Which removes the user's Refresh Token from our Database.
func(s *Service) AccountSignout(
  ctx       context.Context,
  accountID domain.AccountID,
) error {
  if err := s.repo.AccountSignout(ctx, accountID); err != nil {
    return err
  }
  
  return nil
}

func(s *Service) CreateRefreshToken(
  ctx       context.Context,
  entityID  domain.EntityID,
  accountID domain.AccountID,
  role      domain.Role,
)( domain.Token, domain.Token, error ){
  return s.repo.CreateRefreshToken(
    ctx,
    entityID,
    accountID,
    role,
  )
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
