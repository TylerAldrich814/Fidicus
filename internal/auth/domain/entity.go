package domain

import (
  "time"
  "github.com/google/uuid"
)

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

func NewEntity(
  name string,
  description string,
)( *Entity, error ){

  return nil, nil
}
