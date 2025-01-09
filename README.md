# Fidicus

## Overview
NOTE: This is currently a personal project for learning and exploring a variety of backend technologies and architectures.

**Fidicus** is a centralized **Schema Registry and Validation Platform** designed to streamline schema management in **distributed microservice architectures**.
Ensuring consistency, reliability, and backward compatibility for APIs and service contracts across teams and environments.

Built for **medium to large-scale microservices**, Fidicus solves the complex challenges of schema versioning, dependency tracking, and validation by providing a **unified repository** and robust **CI/CD integrations**.

---

## Key Features
- **Centralized Schema Registry**  
    Store and manage schemas for multiple services in one secure location.
    With plans to support the following formats:
      - [x] Protobuf
      - [ ] OpenAPI
      - [ ] GraphQL
- ** Automated Validation and Dependency Tracking:
    Fidicus pulls and parses schema files, creating Neo4j cypher queries 
    for keeping track of your API endpoints syntactical structure and the
    inner-relationships between API endpoints.
      - [x] Protobuf
      - [ ] OpenAPI
      - [ ] GraphQL
- ** CI/CD Pipeline Integrations:
     Integrating directly into Github Actions, Gitlab CI and other CI/CD platforms to let users add an aditional 
     level of testing to help find any potential breaking changes. 
- **Version Control & History**  
  Track schema changes with versioning, providing visibility into modifications and ensuring traceability.

---

## Problems Fidicus Solves

### 1. **Breaking Changes in APIs**
Updating APIs in a microservice environment often risks breaking compatibility with dependent services. Fidicus prevents this by detecting incompatible changes and enforcing validation before deployment.

### 2. **Schema Inconsistencies Across Teams**
Lack of centralized schema storage often leads to duplicated or inconsistent definitions. Fidicus enforces canonical data definitions, improving standardization and interoperability.

### 3. **Poor Visibility into Schema Dependencies**
Large systems with interdependent services can suffer from hidden dependencies. Fidicus provides clear visibility into relationships, helping teams assess the impact of schema modifications.

### 4. **Manual Validation Processes**
Manual schema reviews are time-consuming and error-prone. Fidicus automates validation, reducing manual overhead and ensuring correctness.

### 5. **Compliance & Governance**
Microservices need structured governance around data definitions and changes. Fidicus provides audit trails, version control, and compliance-friendly workflows.

---

## Development Quick Start

1. **Clone the Repository**
```bash
git clone https://github.com/TylerAldrich814/fidicus.git
```

2. **Run the Application**
```bash
docker-compose up -d
```
---

## Roadmap
### Authentication:
- [x] Postgres Auth Database for User Creation using RBAC.
- [x] RBAC Architecture: Currently supports the following users:
   - RoleEntity
   - RoleAdmin
   - RoleAccount
   - RoleReadOnly
- [x] JWT Integration: Creation, Validation and refrehing tokens.
- [X] HTTP middleware for JWT Protected Endpoints and for RBAC Protected Endpoints.

### Schema:
- [] Schema Cypher Compiler:
   - [X] Proto File Support
   - [] GraphQL File Support
   - [] OpenAPI file Support
- [] S3 RBAC Validation and Storage
- [] Schema HTTP endpoints: 


