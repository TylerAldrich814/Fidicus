package domain

import (
	"fmt"
	"time"

	"github.com/google/uuid"
)

type EntityID uuid.UUID 

// NilEntity - Returns a nil EntityID
func NilEntity() EntityID{
  return EntityID(uuid.Nil)
}

// NewEntityID - Creates and returns a new EntityID,
func NewEntityID() EntityID {
  return EntityID(uuid.New())
}

func(e *EntityID)String()string {
  return uuid.UUID(*e).String()
}

func(id EntityID) MarshalJSON() ([]byte, error) {
  return []byte(fmt.Sprintf("\"%s\"", id.String())), nil
}

func(id *EntityID) UnmarshalJSON(data []byte) error {
  str := string(data)
  if len(str) > 1 && str[0] == '"' && str[len(str)-1] == '"' {
    str = str[1 : len(str)-1]
  }

  parsed, err := uuid.Parse(str)
  if err != nil {
    return err
  }

  *id = EntityID(parsed)
  return nil
}

// Entity defines an Entity. An object that connects a plethora of SubAccounts with metadata associated 
// with 
type Entity struct {
  ID          EntityID    `json:"id"`
  Name        string      `json:"name"`
  Description string      `json:"description"`
  AccountIDs  []AccountID `json:"account_ids"`
  CreatedAt   time.Time   `json:"created_at"`
  UpdatedAt   time.Time   `json:"updated_at"`
}


// EntitySignupReq defines the required data needed during an Entity Creation Event.
type EntitySignupReq struct {
  Name        string `json:"name"`
  Description string `json:"description"`
}
