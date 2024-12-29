package domain

import (
	"encoding/json"
	"errors"
)

// Role defines a users access level within an Entity's Account Environment.
//
// Possible Values
//    - AccessRoleUnspecified 
//    - AccessRoleAdmin
//    - AccessRoleUser
//    - AccessRoleReadOnly
type Role uint
const (
  AccessRoleUnspecified Role = iota+1
  AccessRoleAdmin
  AccessRoleUser
  AccessRoleReadOnly
)

var roleFromString = map[string]Role {
  "access_role_unspecified" : AccessRoleUnspecified,
  "access_role_admin"       : AccessRoleAdmin,
  "access_role_user"        : AccessRoleUser,
  "access_role_read_only"   : AccessRoleReadOnly,
}

func(r Role) String() string {
  switch r {
  case AccessRoleUnspecified:
    return "access_role_unspecified"
  case AccessRoleAdmin:
    return "access_role_admin"
  case AccessRoleUser:
    return "access_role_user"
  case AccessRoleReadOnly:
    return "access_role_read_only"
  default: 
    return "unknown"
  }
}

func(r Role) MarshalJSON()( []byte,error ){
  return json.Marshal(r.String())
}

func(r *Role) UnmarshalJSON(data []byte) error {
  var s string
  if err := json.Unmarshal(data, &s); err != nil {
    return err
  }
  role, ok := roleFromString[s]
  if !ok {
    return errors.New("role string failed to unmarshal")
  }
  *r = role

  return nil
}
