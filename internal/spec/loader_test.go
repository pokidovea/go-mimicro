package spec

import (
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
)

func TestLoadFromFile(t *testing.T) {
	// Create temporary directory
	tmpDir := t.TempDir()

	tests := []struct {
		name      string
		content   string
		filename  string
		wantError bool
	}{
		{
			name: "valid YAML",
			content: `openapi: 3.0.0
info:
  title: Test API
  version: 1.0.0
paths:
  /test:
    get:
      responses:
        '200':
          description: Success
`,
			filename:  "test.yaml",
			wantError: false,
		},
		{
			name: "valid JSON",
			content: `{
  "openapi": "3.0.0",
  "info": {
    "title": "Test API",
    "version": "1.0.0"
  },
  "paths": {
    "/test": {
      "get": {
        "responses": {
          "200": {
            "description": "Success"
          }
        }
      }
    }
  }
}`,
			filename:  "test.json",
			wantError: false,
		},
		{
			name:      "invalid YAML/JSON",
			content:   `this is not valid yaml or json { [`,
			filename:  "invalid.yaml",
			wantError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			filePath := filepath.Join(tmpDir, tt.filename)
			err := os.WriteFile(filePath, []byte(tt.content), 0644)
			if err != nil {
				t.Fatalf("failed to create test file: %v", err)
			}

			loader := NewLoader()
			spec, err := loader.Load(filePath)

			if tt.wantError {
				if err == nil {
					t.Error("expected error but got none")
				}
			} else {
				if err != nil {
					t.Errorf("unexpected error: %v", err)
				}
				if spec == nil {
					t.Error("expected spec but got nil")
				}
				if spec.Info.Title != "Test API" {
					t.Errorf("expected title 'Test API', got %q", spec.Info.Title)
				}
				if spec.Info.Version != "1.0.0" {
					t.Errorf("expected version '1.0.0', got %q", spec.Info.Version)
				}
			}
		})
	}
}

func TestLoadFromURL(t *testing.T) {
	tests := []struct {
		name         string
		content      string
		contentType  string
		statusCode   int
		wantError    bool
		expectedInfo Info
	}{
		{
			name: "valid YAML from URL",
			content: `openapi: 3.0.0
info:
  title: Remote API
  version: 2.0.0
paths:
  /remote:
    get:
      responses:
        '200':
          description: Success
`,
			contentType: "application/yaml",
			statusCode:  http.StatusOK,
			wantError:   false,
			expectedInfo: Info{
				Title:   "Remote API",
				Version: "2.0.0",
			},
		},
		{
			name: "valid JSON from URL",
			content: `{
  "openapi": "3.0.0",
  "info": {
    "title": "Remote JSON API",
    "version": "3.0.0"
  },
  "paths": {}
}`,
			contentType: "application/json",
			statusCode:  http.StatusOK,
			wantError:   false,
			expectedInfo: Info{
				Title:   "Remote JSON API",
				Version: "3.0.0",
			},
		},
		{
			name:        "HTTP error",
			content:     "Not Found",
			contentType: "text/plain",
			statusCode:  http.StatusNotFound,
			wantError:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create test server
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Type", tt.contentType)
				w.WriteHeader(tt.statusCode)
				w.Write([]byte(tt.content))
			}))
			defer server.Close()

			loader := NewLoader()
			spec, err := loader.Load(server.URL)

			if tt.wantError {
				if err == nil {
					t.Error("expected error but got none")
				}
			} else {
				if err != nil {
					t.Errorf("unexpected error: %v", err)
				}
				if spec == nil {
					t.Error("expected spec but got nil")
				} else {
					if spec.Info.Title != tt.expectedInfo.Title {
						t.Errorf("expected title %q, got %q", tt.expectedInfo.Title, spec.Info.Title)
					}
					if spec.Info.Version != tt.expectedInfo.Version {
						t.Errorf("expected version %q, got %q", tt.expectedInfo.Version, spec.Info.Version)
					}
				}
			}
		})
	}
}

func TestLoadNonExistentFile(t *testing.T) {
	loader := NewLoader()
	_, err := loader.Load("/nonexistent/path/to/spec.yaml")

	if err == nil {
		t.Error("expected error for nonexistent file")
	}
}

func TestParseComplexSpec(t *testing.T) {
	content := `openapi: 3.0.0
info:
  title: Complex API
  version: 1.0.0
paths:
  /users:
    get:
      summary: Get users
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
      summary: Create user
      requestBody:
        required: true
        content:
          application/json:
            schema:
              $ref: '#/components/schemas/User'
      responses:
        '201':
          description: Created
  /users/{id}:
    get:
      parameters:
        - name: id
          in: path
          required: true
          schema:
            type: integer
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
        age:
          type: integer
          minimum: 0
          maximum: 150
`

	tmpDir := t.TempDir()
	filePath := filepath.Join(tmpDir, "complex.yaml")
	err := os.WriteFile(filePath, []byte(content), 0644)
	if err != nil {
		t.Fatalf("failed to create test file: %v", err)
	}

	loader := NewLoader()
	spec, err := loader.Load(filePath)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Check basic info
	if spec.Info.Title != "Complex API" {
		t.Errorf("expected title 'Complex API', got %q", spec.Info.Title)
	}

	// Check paths
	if len(spec.Paths) != 2 {
		t.Errorf("expected 2 paths, got %d", len(spec.Paths))
	}

	// Check /users path
	usersPath, exists := spec.Paths["/users"]
	if !exists {
		t.Fatal("expected /users path")
	}

	if usersPath.Get == nil {
		t.Error("expected GET operation on /users")
	}

	if usersPath.Post == nil {
		t.Error("expected POST operation on /users")
	}

	if usersPath.Post.RequestBody == nil {
		t.Error("expected request body on POST /users")
	}

	// Check /users/{id} path
	userByIdPath, exists := spec.Paths["/users/{id}"]
	if !exists {
		t.Fatal("expected /users/{id} path")
	}

	if userByIdPath.Get == nil {
		t.Error("expected GET operation on /users/{id}")
	}

	if len(userByIdPath.Get.Parameters) != 1 {
		t.Errorf("expected 1 parameter, got %d", len(userByIdPath.Get.Parameters))
	} else {
		param := userByIdPath.Get.Parameters[0]
		if param.Name != "id" {
			t.Errorf("expected parameter name 'id', got %q", param.Name)
		}
		if param.In != "path" {
			t.Errorf("expected parameter in 'path', got %q", param.In)
		}
		if !param.Required {
			t.Error("expected parameter to be required")
		}
	}

	// Check components
	if spec.Components == nil {
		t.Fatal("expected components")
	}

	if len(spec.Components.Schemas) != 1 {
		t.Errorf("expected 1 schema, got %d", len(spec.Components.Schemas))
	}

	userSchema, exists := spec.Components.Schemas["User"]
	if !exists {
		t.Fatal("expected User schema")
	}

	if userSchema.Type != "object" {
		t.Errorf("expected User type 'object', got %q", userSchema.Type)
	}

	if len(userSchema.Required) != 2 {
		t.Errorf("expected 2 required fields, got %d", len(userSchema.Required))
	}

	if len(userSchema.Properties) != 4 {
		t.Errorf("expected 4 properties, got %d", len(userSchema.Properties))
	}

	// Check email property
	emailProp, exists := userSchema.Properties["email"]
	if !exists {
		t.Fatal("expected email property")
	}

	if emailProp.Format != "email" {
		t.Errorf("expected email format, got %q", emailProp.Format)
	}
}
