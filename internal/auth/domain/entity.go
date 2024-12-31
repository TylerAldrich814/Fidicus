package domain

import (
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
