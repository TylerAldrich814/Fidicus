package domain

// Role defines a accounts access level within an Entity's Account Environment.
//
// Possible Values
//    - AccessRoleUnspecified 
//    - AccessRoleEntity
//    - AccessRoleAdmin
//    - AccessRoleAccount
//    - AccessRoleReadOnly
type Role string
const (
  AccessRoleUnspecified Role = "access_role_unspecified"
  AccessRoleEntity      Role = "access_role_entity"
  AccessRoleAdmin       Role = "access_role_admin"
  AccessRoleAccount     Role = "access_role_account"
  AccessRoleReadOnly    Role = "access_role_read_only"
)
var roleFromString = map[string]Role{
  "access_role_unspecified" : AccessRoleUnspecified,
  "access_role_entity"      : AccessRoleEntity,
  "access_role_admin"       : AccessRoleAdmin,
  "access_role_account"     : AccessRoleAccount,
  "access_role_read_only"   : AccessRoleReadOnly,
}

