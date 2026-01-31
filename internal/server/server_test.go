package server

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"mimicro/internal/spec"
)

func TestHandleRequest(t *testing.T) {
	apiSpec := &spec.OpenAPISpec{
		Info: spec.Info{
			Title:   "Test API",
			Version: "1.0.0",
		},
		Paths: map[string]spec.PathItem{
			"/users": {
				Get: &spec.Operation{
					Responses: map[string]spec.Response{
						"200": {
							Description: "Success",
							Content: map[string]spec.MediaTypeObject{
								"application/json": {
									Schema: &spec.Schema{
										Type: "array",
										Items: &spec.Schema{
											Type: "object",
											Properties: map[string]*spec.Schema{
												"id":   {Type: "integer"},
												"name": {Type: "string"},
											},
										},
									},
								},
							},
						},
					},
				},
			},
		},
	}

	server := New(apiSpec, 8080)

	req := httptest.NewRequest("GET", "/users", nil)
	w := httptest.NewRecorder()

	handler := server.handleRequest("/users", "GET", apiSpec.Paths["/users"].Get)
	handler(w, req)

	resp := w.Result()
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("expected status 200, got %d", resp.StatusCode)
	}

	if contentType := resp.Header.Get("Content-Type"); contentType != "application/json" {
		t.Errorf("expected Content-Type application/json, got %q", contentType)
	}

	var result []interface{}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if len(result) == 0 {
		t.Error("expected non-empty array")
	}
}

func TestValidationErrors(t *testing.T) {
	tests := []struct {
		name           string
		method         string
		path           string
		body           string
		operation      *spec.Operation
		expectedStatus int
		checkError     bool
	}{
		{
			name:   "valid request body",
			method: "POST",
			path:   "/users",
			body:   `{"name": "John", "email": "john@example.com"}`,
			operation: &spec.Operation{
				RequestBody: &spec.RequestBody{
					Required: true,
					Content: map[string]spec.MediaTypeObject{
						"application/json": {
							Schema: &spec.Schema{
								Type: "object",
								Properties: map[string]*spec.Schema{
									"name":  {Type: "string"},
									"email": {Type: "string", Format: "email"},
								},
								Required: []string{"name", "email"},
							},
						},
					},
				},
				Responses: map[string]spec.Response{
					"201": {
						Description: "Created",
						Content: map[string]spec.MediaTypeObject{
							"application/json": {
								Schema: &spec.Schema{
									Type: "object",
									Properties: map[string]*spec.Schema{
										"id":   {Type: "integer"},
										"name": {Type: "string"},
									},
								},
							},
						},
					},
				},
			},
			expectedStatus: http.StatusCreated,
			checkError:     false,
		},
		{
			name:   "missing required field",
			method: "POST",
			path:   "/users",
			body:   `{"name": "John"}`,
			operation: &spec.Operation{
				RequestBody: &spec.RequestBody{
					Required: true,
					Content: map[string]spec.MediaTypeObject{
						"application/json": {
							Schema: &spec.Schema{
								Type: "object",
								Properties: map[string]*spec.Schema{
									"name":  {Type: "string"},
									"email": {Type: "string"},
								},
								Required: []string{"name", "email"},
							},
						},
					},
				},
				Responses: map[string]spec.Response{
					"201": {Description: "Created"},
				},
			},
			expectedStatus: http.StatusBadRequest,
			checkError:     true,
		},
		{
			name:   "invalid email format",
			method: "POST",
			path:   "/users",
			body:   `{"name": "John", "email": "notanemail"}`,
			operation: &spec.Operation{
				RequestBody: &spec.RequestBody{
					Required: true,
					Content: map[string]spec.MediaTypeObject{
						"application/json": {
							Schema: &spec.Schema{
								Type: "object",
								Properties: map[string]*spec.Schema{
									"name":  {Type: "string"},
									"email": {Type: "string", Format: "email"},
								},
								Required: []string{"name", "email"},
							},
						},
					},
				},
				Responses: map[string]spec.Response{
					"201": {Description: "Created"},
				},
			},
			expectedStatus: http.StatusBadRequest,
			checkError:     true,
		},
		{
			name:   "invalid JSON",
			method: "POST",
			path:   "/users",
			body:   `{invalid json}`,
			operation: &spec.Operation{
				RequestBody: &spec.RequestBody{
					Required: true,
					Content: map[string]spec.MediaTypeObject{
						"application/json": {
							Schema: &spec.Schema{
								Type: "object",
							},
						},
					},
				},
				Responses: map[string]spec.Response{
					"201": {Description: "Created"},
				},
			},
			expectedStatus: http.StatusBadRequest,
			checkError:     true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			apiSpec := &spec.OpenAPISpec{
				Info: spec.Info{
					Title:   "Test API",
					Version: "1.0.0",
				},
				Paths: map[string]spec.PathItem{
					tt.path: {
						Post: tt.operation,
					},
				},
			}

			server := New(apiSpec, 8080)

			req := httptest.NewRequest(tt.method, tt.path, bytes.NewBufferString(tt.body))
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()

			handler := server.handleRequest(tt.path, tt.method, tt.operation)
			handler(w, req)

			resp := w.Result()
			defer resp.Body.Close()

			if resp.StatusCode != tt.expectedStatus {
				t.Errorf("expected status %d, got %d", tt.expectedStatus, resp.StatusCode)
			}

			if tt.checkError {
				var errorResp map[string]string
				if err := json.NewDecoder(resp.Body).Decode(&errorResp); err != nil {
					t.Fatalf("failed to decode error response: %v", err)
				}

				if _, exists := errorResp["error"]; !exists {
					t.Error("expected 'error' field in response")
				}
			}
		})
	}
}

func TestPathParameterValidation(t *testing.T) {
	tests := []struct {
		name           string
		pathValue      string
		operation      *spec.Operation
		expectedStatus int
		checkError     bool
	}{
		{
			name:      "valid integer parameter",
			pathValue: "123",
			operation: &spec.Operation{
				Parameters: []spec.Parameter{
					{
						Name:     "id",
						In:       "path",
						Required: true,
						Schema:   &spec.Schema{Type: "integer"},
					},
				},
				Responses: map[string]spec.Response{
					"200": {
						Description: "Success",
						Content: map[string]spec.MediaTypeObject{
							"application/json": {
								Schema: &spec.Schema{
									Type: "object",
									Properties: map[string]*spec.Schema{
										"id": {Type: "integer"},
									},
								},
							},
						},
					},
				},
			},
			expectedStatus: http.StatusOK,
			checkError:     false,
		},
		{
			name:      "invalid integer parameter",
			pathValue: "abc",
			operation: &spec.Operation{
				Parameters: []spec.Parameter{
					{
						Name:     "id",
						In:       "path",
						Required: true,
						Schema:   &spec.Schema{Type: "integer"},
					},
				},
				Responses: map[string]spec.Response{
					"200": {Description: "Success"},
				},
			},
			expectedStatus: http.StatusBadRequest,
			checkError:     true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			apiSpec := &spec.OpenAPISpec{
				Info: spec.Info{
					Title:   "Test API",
					Version: "1.0.0",
				},
			}

			server := New(apiSpec, 8080)

			req := httptest.NewRequest("GET", "/users/"+tt.pathValue, nil)
			req.SetPathValue("id", tt.pathValue)
			w := httptest.NewRecorder()

			handler := server.handleRequest("/users/{id}", "GET", tt.operation)
			handler(w, req)

			resp := w.Result()
			defer resp.Body.Close()

			if resp.StatusCode != tt.expectedStatus {
				t.Errorf("expected status %d, got %d", tt.expectedStatus, resp.StatusCode)
			}

			if tt.checkError {
				var errorResp map[string]string
				if err := json.NewDecoder(resp.Body).Decode(&errorResp); err != nil {
					t.Fatalf("failed to decode error response: %v", err)
				}

				if _, exists := errorResp["error"]; !exists {
					t.Error("expected 'error' field in response")
				}
			}
		})
	}
}

func TestMethodNotAllowed(t *testing.T) {
	pathItem := spec.PathItem{
		Get: &spec.Operation{
			Responses: map[string]spec.Response{
				"200": {Description: "Success"},
			},
		},
		Post: &spec.Operation{
			Responses: map[string]spec.Response{
				"201": {Description: "Created"},
			},
		},
	}

	apiSpec := &spec.OpenAPISpec{
		Info: spec.Info{
			Title:   "Test API",
			Version: "1.0.0",
		},
	}

	server := New(apiSpec, 8080)

	req := httptest.NewRequest("PUT", "/users", nil)
	w := httptest.NewRecorder()

	handler := server.handleMethodNotAllowed("/users", pathItem)
	handler(w, req)

	resp := w.Result()
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusMethodNotAllowed {
		t.Errorf("expected status 405, got %d", resp.StatusCode)
	}

	allow := resp.Header.Get("Allow")
	if allow == "" {
		t.Error("expected 'Allow' header")
	}

	var errorResp map[string]string
	if err := json.NewDecoder(resp.Body).Decode(&errorResp); err != nil {
		t.Fatalf("failed to decode error response: %v", err)
	}

	if _, exists := errorResp["error"]; !exists {
		t.Error("expected 'error' field in response")
	}
}

func TestGetAllowedMethods(t *testing.T) {
	tests := []struct {
		name     string
		pathItem spec.PathItem
		expected []string
	}{
		{
			name: "single method",
			pathItem: spec.PathItem{
				Get: &spec.Operation{},
			},
			expected: []string{"GET"},
		},
		{
			name: "multiple methods",
			pathItem: spec.PathItem{
				Get:    &spec.Operation{},
				Post:   &spec.Operation{},
				Delete: &spec.Operation{},
			},
			expected: []string{"GET", "POST", "DELETE"},
		},
		{
			name: "all methods",
			pathItem: spec.PathItem{
				Get:    &spec.Operation{},
				Post:   &spec.Operation{},
				Put:    &spec.Operation{},
				Delete: &spec.Operation{},
				Patch:  &spec.Operation{},
			},
			expected: []string{"GET", "POST", "PUT", "DELETE", "PATCH"},
		},
		{
			name:     "no methods",
			pathItem: spec.PathItem{},
			expected: []string{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			apiSpec := &spec.OpenAPISpec{}
			server := New(apiSpec, 8080)

			methods := server.getAllowedMethods(tt.pathItem)

			if len(methods) != len(tt.expected) {
				t.Errorf("expected %d methods, got %d", len(tt.expected), len(methods))
			}

			for i, method := range methods {
				if i >= len(tt.expected) {
					break
				}
				if method != tt.expected[i] {
					t.Errorf("expected method %q at position %d, got %q", tt.expected[i], i, method)
				}
			}
		})
	}
}

func TestJoinMethods(t *testing.T) {
	tests := []struct {
		name     string
		methods  []string
		expected string
	}{
		{
			name:     "single method",
			methods:  []string{"GET"},
			expected: "GET",
		},
		{
			name:     "multiple methods",
			methods:  []string{"GET", "POST", "PUT"},
			expected: "GET, POST, PUT",
		},
		{
			name:     "empty",
			methods:  []string{},
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := joinMethods(tt.methods)
			if result != tt.expected {
				t.Errorf("expected %q, got %q", tt.expected, result)
			}
		})
	}
}
