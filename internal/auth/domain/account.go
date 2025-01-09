package domain

import (
	"fmt"
	"time"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
	role "github.com/TylerAldrich814/Fidicus/internal/shared/domain"
)

type AccountID uuid.UUID

// Returns a Nil AccountID
func NilAccount() AccountID{
  return AccountID(uuid.Nil)
}

// NewAccountID - creates and returns a new AccountID
func NewAccountID() AccountID{
  return AccountID(uuid.New())
}

func(a *AccountID)String()string {
  return uuid.UUID(*a).String()
}

func(id AccountID) MarshalJSON() ([]byte, error) {
  return []byte(fmt.Sprintf("\"%s\"", id.String())), nil
}

func(id *AccountID) UnmarshalJSON(data []byte) error {
  str := string(data)
  if len(str) > 1 && str[0] == '"' && str[len(str)-1] == '"' {
    str = str[1 : len(str)-1]
  }

  parsed, err := uuid.Parse(str)
  if err != nil {
    return err
  }

  *id = AccountID(parsed)
  return nil
}

// Account defines the default Entity SubAccount with all of it's paramters.
type Account struct {
  ID              AccountID  `json:"id"`
  EntityID        EntityID   `json:"entity_id"`
  Email           string     `json:"email"`
  PasswHash       string     `json:"password_hash"`
  Role            role.Role  `json:"role"`
  FirstName       string     `json:"first_name"`
  LastName        string     `json:"last_name"`
  CellphoneNumber string     `json:"cellphone_number,omitempty"`
  CreatesAt       time.Time  `json:"created_at"`
  UpdatedAt       time.Time  `json:"updated_at"`
}

// AccountSignupReq - Defines the expected data structure for when new requesting Account owner makes a Signup Request.
type AccountSignupReq struct {
  EntityID        EntityID  `json:"entity_id,omitempty"`
  EntityName      string    `json:"entity_name,omitempty"`
  Email           string    `json:"email"`
  Passw           string    `json:"password"`
  Role            role.Role `json:"role"`
  FirstName       string    `json:"first_name"`
  LastName        string    `json:"last_name"`
  CellphoneNumber string    `json:"cellphone_number,omitempty"`
}

// AccountSigninReq - Defines the expected data structure for when an Account Owner makes a Signin Request. 
type AccountSigninReq struct {
  EntityName string    `json:"entity_name"`
  Email      string    `json:"email"`
  Passw      string    `json:"password"`
  Role       role.Role `json:"role"`
}

func NewAccount(
  email     string,
  passw     string,
  role      role.Role,
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
