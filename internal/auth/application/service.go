package application

import (
	"context"

	"github.com/TylerAldrich814/Schematix/internal/auth/domain"
	"github.com/google/uuid"
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
)( *domain.AuthToken, error ) {
  // Call repo, attemp account signin. Returns AuthToken 
  auth, err := s.repo.AccountSignin(ctx, signInReq)
  if err != nil {
    return nil, err
  }

  return auth, nil
}


func(s *Service) ValidateRefreshToken(
  ctx    context.Context,
  acc_id uuid.UUID,
  token  string,
) error {
  if err := s.repo.ValidateRefreshToken(
    ctx,
    acc_id,
    token,
  ); err != nil {
    // <TODO> Any Serive Middleware actions..?
    return err
  }

  return nil
}
