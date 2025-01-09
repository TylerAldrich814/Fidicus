package jwt

import "errors"

var (
  ErrInternal           = errors.New("an unknown internal error")
  ErrTokenGenFailed     = errors.New("failed to generate new JWT Token")

  ErrTokenGeneration    = errors.New("failed to create auth token")
  ErrTokenMalformed     = errors.New("provided token is malformed")
  ErrTokenInvalid       = errors.New("failed to verify jwt token")
  ErrTokenInvalidAlg    = errors.New("invalid jwt verification algorithm")
  ErrTokenExpired       = errors.New("jwt token is expired")
  ErrTokenInvalidClaims = errors.New("invalid jwt token claims")
  ErrTokenInvalidSig    = errors.New("invalid jwt token signature")
)
