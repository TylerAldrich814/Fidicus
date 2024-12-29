package domain

import (
  "time"

  "golang.org/x/crypto/bcrypt"
  "github.com/google/uuid"
)

type EntityID string
type AccountID string

// Account defines an Entity SubAccount
type Account struct {
  ID              uuid.UUID `json:"id"`
  EntityID        uuid.UUID `json:"entity_id"`
  Email           string    `json:"email"`
  PasswHash       string    `json:"password_hash"`
  Role            Role      `json:"role"`
  FirstName       string    `json:"first_name"`
  LastName        string    `json:"last_name"`
  CellphoneNumber string    `json:"cellphone_number"`
  CreatesAt       time.Time `json:"created_at"`
  UpdatedAt       time.Time `json:"updated_at"`
}

func NewAccount(
  email  string,
  passw  string,
  role   Role,
  firstName string,
  lastName  string,
  number    string,
)( *Account,error ){

  passHash, err := HashPassword(passw)
  if err != nil {
    return nil, err
  }

  return &Account {
    Email           : email,
    PasswHash       : passHash,
    FirstName       : firstName,
    LastName        : lastName,
    CellphoneNumber : number,
  }, nil
}


// HashPassword - Takes in an Account password and hashes it.
func HashPassword(passw string)( string, error ){
  hash, err := bcrypt.GenerateFromPassword(
    []byte(passw),
    bcrypt.DefaultCost,
  )
  if err != nil {
    return "", err
  }

  return string(hash), nil
}

// ValidatePassword - Validates a password against a stored hash of the raw password.
func ValidatePassword(password, hash string) bool {
  err := bcrypt.CompareHashAndPassword(
    []byte(hash),
    []byte(password),
  )
  return err == nil
}
