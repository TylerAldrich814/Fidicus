package application

import (
	"context"

	"github.com/TylerAldrich814/Schematix/internal/auth/domain"
	"github.com/TylerAldrich814/Schematix/pkg/jwt"
)

type Service struct {
  repo domain.AuthRepository
  jwt  *jwt.JWTHandler
}

func NewService(
  repo domain.AuthRepository,
  jwt  *jwt.JWTHandler,
) *Service {
  return &Service{ repo, jwt }
}

// CreateRootAccount Will first attempt to create the Root account(i.e., the Entity)
// If Successful, we then call 'CreateSubAccount', creating the Entity's first Subaccount
// if a Admin level Role. And Finally, we return the EntityID and UserID
func(s *Service) CreateRootAccount(
  ctx context.Context,
  entity domain.Entity,
  user domain.User,
)( domain.EntityID, domain.UserID, error ){
  return s.repo.CreateEntity(ctx, entity, user)
}

// CreateSubAccount attempts to create a Sub Account under a specified Entity.
func(s *Service) CreateSubAccount(
  ctx  context.Context,
  user domain.User,
)( domain.UserID, error) {
  return s.repo.CreateAccount(ctx, user)
}

func(s *Service) UserSignin(
  ctx   context.Context,
  creds domain.Credentials,
)( *domain.AuthToken, error ) {
  // Call repo, attemp user signin. Returns AuthToken 
  auth, err := s.repo.UserSignin(ctx, creds)
  if err != nil {
    return nil, err
  }

  return auth, nil
}

func(s *Service) RefreshTokens(
  ctx context.Context,
  refreshToken string,
)( *domain.AuthToken, error ){
  tokens, err := s.repo.RefreshToken(ctx, refreshToken)
  if err != nil {
    return nil, err
  }

  return tokens, nil
}

// func(s *Service) VerifyAccessToken(
//   ctx context.Context, 
//   accessToken string,
// )( *domain.AuthClaims, error ){
//
//   return nil, nil
// }
