# Fidicus

## Overview
**Fidicus** is a centralized **Schema Registry and Validation Platform** designed to streamline schema management in **distributed microservice architectures**. It ensures consistency, reliability, and backward compatibility for APIs and service contracts across teams and environments.

Built for **medium to large-scale microservices**, Fidicus solves the complex challenges of schema versioning, dependency tracking, and validation by providing a **unified repository** and robust **CI/CD integrations**.

---

## Key Features

- **Centralized Schema Registry**  
  Store and manage schemas for multiple services in one secure location. Supports OpenAPI, GraphQL, and Protobuf formats.

- **Version Control & History**  
  Track schema changes with versioning, providing visibility into modifications and ensuring traceability.

- **Automated Validation & Compatibility Checks**  
  Identify breaking changes in schema updates, preventing runtime errors and deployment failures.

- **Dependency Mapping & Consumer Tracking**  
  Map dependencies between services to evaluate the impact of schema updates on consumers.

- **CI/CD Pipeline Integration**  
  Integrate directly into GitHub Actions, GitLab CI, and other CI/CD pipelines to enforce schema validation during build and deployment stages.

- **OAuth2 Authentication & RBAC**  
  Secure access control with Role-Based Access Control (RBAC) and OAuth2 support for GitHub/GitLab integrations.

- **Graph-Based Relationships**  
  Leverage graph databases to map and analyze schema dependencies and relationships between API components.

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

## Target Audience
Fidicus is built for:
- **DevOps Engineers** automating schema validation within CI/CD pipelines.
- **API Developers** managing and versioning schemas across services.
- **Product Teams** ensuring API stability for internal and external consumers.
- **Enterprise Architects** maintaining compliance and governance in distributed systems.

---

## Development Quick Start

1. **Clone the Repository**
```bash
git clone https://github.com/your-org/fidicus.git
```

2. **Run the Application**
```bash
docker-compose up
```

3. **Access the Dashboard**  
Navigate to: `http://localhost:3000`

4. **API Integration Example**
```bash
curl -X POST http://localhost:8080/api/v1/schemas \
-H "Authorization: Bearer <TOKEN>" \
-H "Content-Type: application/json" \
-d @schema.json
```

---

## Roadmap
- **Multi-Format Support**: Add Avro and JSON Schema compatibility.
- **Consumer Contracts**: Enable tracking of downstream consumers and their schema dependencies.
- **Custom Rules & Validations**: Allow configurable validation rules for specific use cases.
- **Plugin System**: Introduce custom extensions and plugins for specialized workflows.
- **Advanced Analytics**: Provide insights into schema usage and API performance.

---

## Contributing
We welcome contributions! Please check our [CONTRIBUTING.md](CONTRIBUTING.md) for guidelines.

---

## License
This project is licensed under the [EULA License](LICENSE.md).

---

## Contact
For inquiries, please contact: **support@fidicus.io**


