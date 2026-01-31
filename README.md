# Mimicro

[![codecov](https://codecov.io/github/pokidovea/go-mimicro/graph/badge.svg?token=DERg3q4rVa)](https://codecov.io/github/pokidovea/go-mimicro)

Mock HTTP server based on OpenAPI specifications.

## Description

Mimicro is a lightweight mock server that reads an OpenAPI (Swagger) specification and automatically creates HTTP endpoints that return random data matching the schema definitions. It's designed to be used as a stub for testing integrations with external services.

## Features

- **OpenAPI 3.0 Support**: Reads OpenAPI specifications in JSON or YAML format
- **Local and Remote Specs**: Load specifications from local files or URLs
- **Automatic Endpoint Registration**: All paths and HTTP methods from the spec are registered automatically
- **Request Validation**: Validates request body, path parameters, and query parameters
- **Smart Error Responses**: Returns 400 Bad Request for validation errors, 405 Method Not Allowed for unsupported methods
- **Smart Data Generation**: Generates random data that matches schema types and constraints
- **Type Support**: object, array, string, integer, number, boolean
- **Format Support**: date, date-time, email, uuid, uri
- **Constraint Support**: min/max for numbers, minLength/maxLength for strings, enum values, pattern matching
- **Component References**: Resolves `$ref` references to `#/components/schemas`

## Installation

### Prerequisites

- Go 1.20 or higher

### Build from source

```bash
git clone <repository-url>
cd mimicro
go build -o mimicro
```

## Usage

### Basic usage

```bash
./mimicro -spec <path-or-url-to-openapi.yaml>
```

### With a custom port

```bash
./mimicro -spec openapi.yaml -port 3000
```

### Command-line flags

- `-spec` (required): Path or URL to OpenAPI specification file (YAML or JSON)
- `-port` (optional): Port to run the mock server on (default: 8080)

## Examples

### Example 1: Local file

```bash
./mimicro -spec ./api-spec.yaml
```

### Example 2: Remote URL

```bash
./mimicro -spec https://petstore3.swagger.io/api/v3/openapi.json
```

### Example 3: Custom port

```bash
./mimicro -spec ./my-api.yaml -port 9000
```

## Sample OpenAPI Spec

```yaml
openapi: 3.0.0
info:
  title: Sample API
  version: 1.0.0
paths:
  /users:
    get:
      summary: Get all users
      responses:
        '200':
          description: Success
          content:
            application/json:
              schema:
                type: array
                items:
                  $ref: '#/components/schemas/User'
    post:
      summary: Create a new user
      requestBody:
        required: true
        content:
          application/json:
            schema:
              $ref: '#/components/schemas/User'
      responses:
        '201':
          description: Created
          content:
            application/json:
              schema:
                $ref: '#/components/schemas/User'
  /users/{id}:
    get:
      summary: Get user by ID
      parameters:
        - name: id
          in: path
          required: true
          schema:
            type: integer
            minimum: 1
      responses:
        '200':
          description: Success
          content:
            application/json:
              schema:
                $ref: '#/components/schemas/User'
components:
  schemas:
    User:
      type: object
      required:
        - name
        - email
      properties:
        id:
          type: integer
        name:
          type: string
          minLength: 3
          maxLength: 50
        email:
          type: string
          format: email
        created_at:
          type: string
          format: date-time
```

### Successful response example

```bash
curl http://localhost:8080/users/123
```

```json
{
  "id": 42,
  "name": "kXmPqRsT",
  "email": "user573@example.com",
  "created_at": "2026-01-31T10:15:30Z"
}
```

### Validation error examples

#### Invalid path parameter

```bash
curl http://localhost:8080/users/abc
```

```json
{
  "error": "id: Must be an integer"
}
```

#### Missing the required field in the request body

```bash
curl -X POST http://localhost:8080/users \
  -H "Content-Type: application/json" \
  -d '{"name": "John"}'
```

```json
{
  "error": "body.email: Required field is missing"
}
```

#### Method is not allowed

```bash
curl -X PUT http://localhost:8080/users
```

```json
{
  "error": "Method PUT not allowed. Allowed methods: GET, POST"
}
```

## Data Generation Rules

### Strings
- Default length: 5–20 characters
- Respects `minLength` and `maxLength` constraints
- Special formats:
  - `date`: YYYY-MM-DD
  - `date-time`: RFC3339 format
  - `email`: user{random}@example.com
  - `uuid`: valid UUID v4 format
  - `uri`: https://example.com/resource/{random}

### Numbers
- **integer**: Random integer between min and max (default: 0-1000)
- **number**: Random float between min and max (default: 0.0-1000.0)
- Respects `minimum` and `maximum` constraints

### Arrays
- Generates 1-5 items by default
- Each item matches the `items` schema

### Objects
- Generates all properties defined in the schema
- Each property matches its defined schema

### Enums
- Randomly selects one value from the enum list

### References
- Resolves `$ref` to `#/components/schemas/{SchemaName}`
- Supports nested references

## Validation

Mimicro validates all incoming requests against the OpenAPI specification:

### Request Body Validation
- Validates JSON structure against schema
- Checks required fields
- Validates data types (string, integer, number, boolean, object, array)
- Enforces constraints (min/max, minLength/maxLength, pattern, enum)
- Validates formats (email, uuid, uri, date, date-time)
- Returns **400 Bad Request** with error details if validation fails

### Path Parameters Validation
- Validates parameter types (string, integer, number, boolean)
- Enforces constraints (min/max for numbers)
- Checks required parameters
- Returns **400 Bad Request** with error details if validation fails

### Query Parameters Validation
- Validates parameter types
- Enforces constraints
- Checks required parameters
- Returns **400 Bad Request** with error details if validation fails

### Method Validation
- Returns **405 Method Not Allowed** if the endpoint doesn't support the HTTP method
- Includes `Allow` header with supported methods

## Limitations

- Only supports OpenAPI 3.0 specifications
- Always returns the first 2xx response defined in the spec
- Does not support authentication/authorization
- Does not persist data between requests
- Basic format validation (not fully RFC-compliant)

## License

MIT

## Contributing

Contributions are welcome! Please feel free to submit a Pull Request.
