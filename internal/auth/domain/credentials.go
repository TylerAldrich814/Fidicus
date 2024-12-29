package domain


type Credentials struct {
  EntityName  string `json:"entity_name"`
  Email       string `json:"email"`
  Password    string `json:"password"`
}
