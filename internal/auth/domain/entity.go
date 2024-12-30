package domain

import (
  "time"
  "github.com/google/uuid"
)

type EntityID  uuid.UUID

func NilEntity() EntityID{
  return EntityID(uuid.Nil)
}

// Entity defines an Entity. An object that connects a plethora of SubAccounts with metadata associated 
// with 
type Entity struct {
  ID          uuid.UUID   `json:"id"`
  Name        string      `json:"name"`
  Description string      `json:"description"`
  AccountIDs  []uuid.UUID `json:"account_ids"`
  CreatedAt   time.Time   `json:"created_at"`
  UpdatedAt   time.Time   `json:"updated_at"`
}


// EntitySignupReq defines the required data needed during an Entity Creation Event.
type EntitySignupReq struct {
  Name        string `json:"name"`
  Description string `json:"description"`
}
