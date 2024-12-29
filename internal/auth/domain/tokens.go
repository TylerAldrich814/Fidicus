package domain

import (
	"errors"
	"fmt"
	"time"

	"github.com/TylerAldrich814/Schematix/internal/shared/config"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	log "github.com/sirupsen/logrus"
)

var (
  jwtSigningSecret = config.GetEnv("JWT_SECRET_KEY", "TODO")
)

// Token - For Backend operations. Contains the final Signed JWT Token and it's Expiration
type Token struct {
  SignedToken string    `json:"signed_token"`
  Expiration  time.Time `json:"expiration"`
}

// AuthToken - Defines a structure for passing both Access and Refresh Tokens. 
type AuthToken struct {
  AccessToken  Token `json:"access_token"`
  RefreshToken Token `json:"refresh_token"`
  AccessRole   Role  `json:"role"`
}

// AuthClaims - Defines our custom JWT Token Claims to be added into each Token.
type AuthClaims struct {
  UserID    string  `json:"user_id"`
  EntityID  string  `json:"entity_id"`
  Role      Role    `json:"role"`
  jwt.RegisteredClaims
}

// GenerateToken creates a single JWT Token with a custom expiration time.
func GenerateToken(
  userID   uuid.UUID,
  entityID uuid.UUID,
  role     Role,
  exp      time.Duration,
)( Token, error ){
  claims := AuthClaims {
    UserID   : userID.String(),
    EntityID : entityID.String(),
    Role     : role,
    RegisteredClaims : jwt.RegisteredClaims{
      ExpiresAt : jwt.NewNumericDate(time.Now().Add(exp)),
      IssuedAt  : jwt.NewNumericDate(time.Now()),
    },
  }

  accessToken := jwt.NewWithClaims(
    &jwt.SigningMethodHMAC{},
    claims,
  )

  tokenString, err := accessToken.SignedString([]byte(jwtSigningSecret))
  if err != nil {
    return Token{}, err
  }

  return Token{
    SignedToken : tokenString,
    Expiration  : claims.RegisteredClaims.ExpiresAt.Time,
  }, nil
}


// GenerateJWTTokens creates both Access and Regresh JWT Tokens for a user.
//   - Access Token will have an exiration of 1 hour
//   - Refresh Token will have an expiration of 7 days.
func GenerateJWTTokens(
  userID   uuid.UUID,
  entityID uuid.UUID,
  role     Role,
)( *AuthToken, error ){
  accessToken, err := GenerateToken(
    userID,
    entityID,
    role,
    time.Duration(1 * time.Hour),
  )
  if err != nil {
    log.Printf("Failed to create AcceessToken: %v", err)
    return nil, ErrTokenGeneration
  }

  refreshToken, err := GenerateToken(
    userID,
    entityID,
    role,
    time.Duration(7 * 24 * time.Hour),
  )
  if err != nil {
    log.Printf("Failed to create RefreshToken: %v", err)
    return nil, ErrTokenGeneration
  }

  return &AuthToken{
    AccessToken  : accessToken,
    RefreshToken : refreshToken,
    AccessRole   : role,
  },nil
}

// VerifyToken - Attempts to validate a given JWT token. Returns specified error if validation fails for any reason.
func VerifyToken(
  rtoken string,
)( *AuthClaims, error ){
  var logError = func(f string, args ...any) {
    log.Error(fmt.Sprintf("VerifyToken: " + f, args...))
  }
  // Parse Signed JWT Token String:
  token, err := jwt.ParseWithClaims(
    rtoken,
    &AuthClaims{},
    func(token *jwt.Token)( interface{}, error ){
      if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
        log.Error(fmt.Sprintf("VerifyToken: unexpected signing method: %s", token.Header["alg"]))
        return nil, ErrTokenInvalidAlg
      }
      return []byte(jwtSigningSecret), nil
    },
  )
  if err != nil {
    if errors.Is(err, jwt.ErrTokenMalformed){
      logError("Token is Malformed: %v", err)
      return nil, ErrTokenMalformed
    } else if errors.Is(err, jwt.ErrTokenSignatureInvalid){
      logError("Invalid Token Signature: %v", err)
      return nil, ErrTokenInvalidSig
    } else if errors.Is(err, ErrTokenExpired) || 
              errors.Is(err, jwt.ErrTokenNotValidYet){
      logError("Token is Expired: %v", err)
      return nil, ErrTokenExpired
    }
    logError("An unknown error occurred: %v", err)
    return nil, ErrInternal
  }

  claims, ok := token.Claims.(*AuthClaims)
  if !ok {
    logError("invalid Auth Claims")
    return nil, ErrTokenInvalidClaims
  }

  if claims.ExpiresAt.Time.Before(time.Now()) {
    return nil, ErrTokenExpired
  }

  return claims, nil
}
