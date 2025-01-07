CREATE (:Message {
  name: "User",
  package: "user",
  fields: ["id:TYPE_STRING", "name:TYPE_STRING", "age:TYPE_INT32"]
})CREATE (:Message {
  name: "GetUserRequest",
  package: "user",
  fields: ["id:TYPE_STRING"]
})CREATE (:Message {
  name: "GetUserResponse",
  package: "user",
  fields: ["user:TYPE_MESSAGE"]
})CREATE (:Service {
  name: "UserService",
  package: "user"
})CREATE (:Method {
  name: "GetUser",
  input: "GetUserRequest",
  output: "GetUserResponse"
})
MATCH (svc:Service {name: "UserService"})
MATCH (m:Method {name: "GetUser"})
MERGE (svc)-[:EXPOSES]->(m)

