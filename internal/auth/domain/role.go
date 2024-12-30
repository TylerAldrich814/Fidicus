package domain

import (
	"encoding/json"
)

// Role defines a accounts access level within an Entity's Account Environment.
//
// Possible Values
//    - AccessRoleUnspecified 
//    - AccessRoleAdmin
//    - AccessRoleAccount
//    - AccessRoleReadOnly
type Role uint
const (
  AccessRoleUnspecified Role = iota+1
  AccessRoleEntity
  AccessRoleAdmin
  AccessRoleAccount
  AccessRoleReadOnly
)

var roleFromString = map[string]Role {
  "access_role_unspecified" : AccessRoleUnspecified,
  "access_role_entity"      : AccessRoleEntity,
  "access_role_admin"       : AccessRoleAdmin,
  "access_role_account"     : AccessRoleAccount,
  "access_role_read_only"   : AccessRoleReadOnly,
}

func(r Role) String() string {
  switch r {
  case AccessRoleUnspecified:
    return "access_role_unspecified"
  case AccessRoleAdmin:
    return "access_role_admin"
  case AccessRoleEntity:
    return "access_role_entity"
  case AccessRoleAccount:
    return "access_role_account"
  case AccessRoleReadOnly:
    return "access_role_read_only"
  default: 
    return "access_role_unspecified"
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
    *r = AccessRoleUnspecified
  }
  *r = role

  return nil
}
