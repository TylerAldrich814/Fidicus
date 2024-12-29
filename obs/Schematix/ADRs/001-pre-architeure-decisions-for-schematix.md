# 001-pre-architecture-descisions-for-schematix

## **Status**:
Accepted

## **Context**:

## A B2B SAAS Project for creating a Developer tool for helping Micro Service developers keep a central schematic repository that will solve the following problems:
  - Service Schema Validation and breaking changes deterrence. Letting developers creates a central repository; a Central of *'truth'* for their schemas.
  - Micro Service Schema compatibility layering. i.e., Creating a service that will attempt to convert one schematic variant into another( from gRPC -> OpenAPI)
  - *Schema Parameter Consistency*; Keeping a relational database that stores every schematic parameter's
	  - semantic meaning
	  - data type
	  - evolutionary history
	  - parental ownership(which schemas own which variant)  
	- Example: in cases where two separate services, at two different historical times, define a schema parameter which has a different name(key) but they both semantically the same.i.e.,
		- Zipcode, PostalCode, zip, postCode, etc..
		- firstName, givenName, fName, etc..
		- lastName, surname, familyName, paternalName, etc..
		- dob, birthDate, dateOfBirth, etc..
		- phoneNumber, mobile, telephone, cellNumber, etc..
	- These examples showcase that even though their names(keys) differ, but how they are also semantically the same. Which also means that they will contain the same data primitives. 
  - *Backward Compatibility Testing*: 
	  - Automatically detect if a newly proposed schema is incompatible with existing consumers.
	  - For example; If you remove a field that's being used by a known consumer, the system flags the change.
  - **CI/CD** integration:
	  - Hooks into Github Actions, Gitlab CI, Jenkins, or other CI systems. Create an abstraction layer over CI/CD in order to make 3rd party integration setup easier. 
  - **Documentation & Developer Portal**:
	  - Auto-generate docs or each schema or endpoint, helping internal teams or external partners quickly understand how to consume the service.
	  - Provide a sandbox or "try it out" feature for REST or GraphQL, etc..

## Tech Stack:
### Web Application( Frontend+Backend ):
 - **Frontend**: A dashboard for browsing schemas, generating Diffs, and seeing validation results. Possibly React or HTMX+Go. 
 - **Backend**: Micro Service Architecture with the following Services:
    - Static/generates front end pages. 
	 - Schema Handler: Upload, Download, Validators, Generators, Result Reporting, Diffs, etc..
### **Database Layer**:
 - Utilize either Relational or Non-Relational Databases for client-side operations.
	 - Either PostgreSQL or MongoDB:
	 - Schema Storage: Upload, Download, diffs, etc..
	 - Client Data storage

### **Validation/Comparison Engine**:
 - Contains the logic for comparing new and old versions. 
 - Utilizing a GDB( Graph Database )for storing the semantical information for Service Schemas. This could be used for both Validating schemas and for helping visualize to the frontend what a user's entire system looks like. 
   - **Nodes**: 
	   - **Service Nodes**: Each Microservice(i.e., "AuthService", "OrderService", etc..)
	   - **Schema Nodes**: Each version of a schema(i.e., "AuthService:v1.0", "OrderService:v0.5", etc..)
	   - **Field/Entity Nodes**: Individual fields or data entities. For tracking each schematic primitive at a granular level.
   - **Edges**: 
	   - **Creators & Consumers**: These Edges will determine which other *service* consumes or depends on. If a service defined by *Service B*'s schema; You'd have an edge from *Service A -> Service B: schema XX*.
	   - **Evolves Into**: If a schema changes from v1.0 to v1.1, you could store a versioning edge: **V1.0 -> V1.1**.
	   - **Contains**: Tracking field-level nodes, we would then link *"AuthService:v1.1"* to a node for "countryCode" or "postalCode" to show these fields live in that schema.
   - **Properties**: 
	   - Properties can include metadata about how strongly the dependency is enforced(e.g, optional usage vs. must-have usage).
	   - Node properties might describe the schema's format(OpenAPI, GraphQL, gRPC, etc..) or the last updated timestamp, etc..
 - **Utilizing Graph Queries**:
	 - **Impact Analysis**: 
		 - Seeing which consumers will break when you remove or rename a field, you can create a query from *"schema node X"* that follows edges to all *downstream* services.
	 - **Cypher Example**(*Neo4j style*):
```sql
MATCH (s:Schema {name: "AuthService v1.1})<-[:DEPENDS_ON]-(Consumer)
RETURN customer
```
 - **Schema Compatibility Checks**:
	 - Suppose you store a **"new version"** node. A Check or script compares *"v1.1"* to *"v1.2"*  for backward compatibility. If it finds a removed field used by any consumer, you can query the graph for which edges are impacted.
	 - This is beneficial if you have complex or indirect dependencies(e.g, Service A depends on Service B, which in turn depends on Service C and so forth).
 - **Visualization**:
	 - As mentioned above, utilizing a GDB results in the added benefit of UI Visualization tools that will allow users to visually inspect their Schematic relationships by seeing the Node/Edge relationships in a Graph layout.
	    - Neo4j
	    - ArangoDB
	    - Graph
	    - etc..
 - **Potential Architecture**:
	 - **Schema Storage**:
	   - We'll keep the Raw Schema files(e.g., OpenAPI, JSON, YAML, Proto, etc) in a blob store or document DB.
	   - The GDB will then reference them, linking the schema version node to the physical location of *ID* for easy retrieval.
	 - **Validation Logic**:
		- When a new schema version is proposed. A *Validation* Diffing service will trigger and run pre/post GDB Queries and determine if any breaking changes have occurred.
		- If the Diff detects a potential breaking change, it will compile the Diff results of Pre Query vs Post Query and compile the report for the user.
		- Depending on the Query result. It either blocks the update or flags it for manual approval.
	 - **CI/CD Integration**:
		- The developer commits a new/update schema file.
		- A Pipeline step uploads the new version to the GDB( creating a new Node or linking it the old version).
		- The pipeline calls the Validation Service to do a  graph-based dependency check.
		- If any red flags are found, the pipeline fails and notifies the relevant stakeholders.
 - **Pros | Cons**:
	 - **Pros**:
		 - Rich Relationship Modeling: Graph Databases excel at many-to-many relationships. Which in the case of Schema primitive relationships between many services, would make the most sense.
		 - Complex Dependency Analysis: You can quickly traverse the graph to find all impacted service or fields.
		 - Visualization & Query Flexibility: Tools like Neo4j Bloom of ArangoDB's visual explores are intuitive for seeing how everything connects.
	 - **Cons**:
	    - Setup & Maintenance Overhead: Running a GDB cluster can be more complex than a single relational Database.
		 - Learning Curve: Teams need to be more comfortable with graph queries.
		 - Overfill for Simple Use Cases: If a client only has a handful of services and straight forward dependencies. A relational or document-based approach would suffice.
 - **Example Queries**:
	 - **Scenario**: A customer wants to publish a new version of **OrderService:V1.2**. They removed the field *province*. The system must determine if that field was used by **CheckoutService** or **AnalyticsService**.
	   - **Graph Update**:
		   - A new node `(Schema { service: "OrderService", version: "v1.2", removed: ["province"], ...})`
	   - **Compatibility Check**:
		   - The validation engine runs a Cypher query:
		   - This retrieves any *"consumer:Service"* that claims to require the *province* parameter. 
			```SQL
	MATCH (o:Schema {service: "OrderService", version: "v1.2"})
     -[:REMOVED_FIELD]->(f:Field {name: "Province"})
	 <-[:USES_FIELD]-(consumer:Service)
RETURN consumer 
```

 - **Result and Action**:
	 - If the query returns **(CheckoutService)**, the registry knows removing *province* might break **CheckoutService**.
	 - The system can either block the change or require manual override.
 
# **Decision**:
Developing a profitable B2B SaaS as a solo founder is challenging yet promising. Based on preliminary research, a **Service Schema Validator & Repository** stands out as a strong candidate for commercialization. A critical pain point in microservice architectures arises when the service landscape becomes so extensive that no single developer can reliably track all interdependencies. In response, many organizations allocate significant resources to build proprietary tools that manage schema versioning and compatibility checks. This gap indicates a clear market opportunity for an off-the-shelf solution that centralizes, validates, and enforces schema integrity across distributed services.
# **Consequences**:
- **Market Differentiation:** By offering a ready-made solution for schema validation and repository management, the product can fill a significant gap in microservice governance. This may help small to midsize companies avoid costly in-house development efforts and speed up their DevOps cycles.
- **Technical Complexity:** Implementing robust validation logic, managing version compatibility, and providing a user-friendly repository interface will require careful planning and iterative development. A lack of deep domain expertise or feature maturity may limit early adoption.
- **Operational Overhead:** Hosting and maintaining a centralized registry for client data introduces additional security, reliability, and performance requirements. Ensuring high uptime, data protection, and regulatory compliance could become challenging, especially for a solo founder.
- **Long-Term Scalability:** If successful, the solution will need to handle increasing volumes of schemas and more complex dependency graphs. The underlying architecture must be designed to accommodate these growth scenarios without significant refactoring or infrastructure overhauls.