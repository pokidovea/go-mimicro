package validator

import (
	"bytes"
	"io"
	"net/http/httptest"
	"testing"

	"mimicro/internal/spec"
)

func TestValidateRequestBody(t *testing.T) {
	tests := []struct {
		name      string
		schema    *spec.Schema
		body      string
		reqBody   *spec.RequestBody
		wantError bool
		errorMsg  string
	}{
		{
			name: "valid object",
			reqBody: &spec.RequestBody{
				Required: true,
				Content: map[string]spec.MediaTypeObject{
					"application/json": {
						Schema: &spec.Schema{
							Type: "object",
							Properties: map[string]*spec.Schema{
								"name": {Type: "string"},
								"age":  {Type: "integer"},
							},
							Required: []string{"name"},
						},
					},
				},
			},
			body:      `{"name": "John", "age": 30}`,
			wantError: false,
		},
		{
			name: "missing required field",
			reqBody: &spec.RequestBody{
				Required: true,
				Content: map[string]spec.MediaTypeObject{
					"application/json": {
						Schema: &spec.Schema{
							Type: "object",
							Properties: map[string]*spec.Schema{
								"name": {Type: "string"},
								"age":  {Type: "integer"},
							},
							Required: []string{"name"},
						},
					},
				},
			},
			body:      `{"age": 30}`,
			wantError: true,
			errorMsg:  "Required field is missing",
		},
		{
			name: "invalid type",
			reqBody: &spec.RequestBody{
				Required: true,
				Content: map[string]spec.MediaTypeObject{
					"application/json": {
						Schema: &spec.Schema{
							Type: "object",
							Properties: map[string]*spec.Schema{
								"age": {Type: "integer"},
							},
						},
					},
				},
			},
			body:      `{"age": "not a number"}`,
			wantError: true,
			errorMsg:  "Must be an integer",
		},
		{
			name: "string too short",
			reqBody: &spec.RequestBody{
				Required: true,
				Content: map[string]spec.MediaTypeObject{
					"application/json": {
						Schema: &spec.Schema{
							Type: "object",
							Properties: map[string]*spec.Schema{
								"name": {
									Type:      "string",
									MinLength: intPtr(5),
								},
							},
						},
					},
				},
			},
			body:      `{"name": "Bob"}`,
			wantError: true,
			errorMsg:  "Must be at least 5 characters long",
		},
		{
			name: "number below minimum",
			reqBody: &spec.RequestBody{
				Required: true,
				Content: map[string]spec.MediaTypeObject{
					"application/json": {
						Schema: &spec.Schema{
							Type: "object",
							Properties: map[string]*spec.Schema{
								"age": {
									Type:    "integer",
									Minimum: float64Ptr(18),
								},
							},
						},
					},
				},
			},
			body:      `{"age": 10}`,
			wantError: true,
			errorMsg:  "Must be at least 18",
		},
		{
			name: "invalid email format",
			reqBody: &spec.RequestBody{
				Required: true,
				Content: map[string]spec.MediaTypeObject{
					"application/json": {
						Schema: &spec.Schema{
							Type: "object",
							Properties: map[string]*spec.Schema{
								"email": {
									Type:   "string",
									Format: "email",
								},
							},
						},
					},
				},
			},
			body:      `{"email": "notanemail"}`,
			wantError: true,
			errorMsg:  "Must be a valid email address",
		},
		{
			name: "valid array",
			reqBody: &spec.RequestBody{
				Required: true,
				Content: map[string]spec.MediaTypeObject{
					"application/json": {
						Schema: &spec.Schema{
							Type: "array",
							Items: &spec.Schema{
								Type: "string",
							},
						},
					},
				},
			},
			body:      `["item1", "item2", "item3"]`,
			wantError: false,
		},
		{
			name: "invalid JSON",
			reqBody: &spec.RequestBody{
				Required: true,
				Content: map[string]spec.MediaTypeObject{
					"application/json": {
						Schema: &spec.Schema{
							Type: "object",
						},
					},
				},
			},
			body:      `{invalid json}`,
			wantError: true,
			errorMsg:  "Invalid JSON",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			apiSpec := &spec.OpenAPISpec{}
			v := New(apiSpec)

			req := httptest.NewRequest("POST", "/test", bytes.NewBufferString(tt.body))
			req.Header.Set("Content-Type", "application/json")

			operation := &spec.Operation{
				RequestBody: tt.reqBody,
			}

			err := v.ValidateRequestBody(req, operation)

			if tt.wantError {
				if err == nil {
					t.Errorf("expected error but got none")
					return
				}
				if tt.errorMsg != "" && !contains(err.Error(), tt.errorMsg) {
					t.Errorf("expected error containing %q, got %q", tt.errorMsg, err.Error())
				}
			} else {
				if err != nil {
					t.Errorf("unexpected error: %v", err)
				}
			}
		})
	}
}

func TestValidatePathParameters(t *testing.T) {
	tests := []struct {
		name       string
		parameters []spec.Parameter
		pathValues map[string]string
		wantError  bool
		errorMsg   string
	}{
		{
			name: "valid integer parameter",
			parameters: []spec.Parameter{
				{
					Name:     "id",
					In:       "path",
					Required: true,
					Schema:   &spec.Schema{Type: "integer"},
				},
			},
			pathValues: map[string]string{"id": "123"},
			wantError:  false,
		},
		{
			name: "invalid integer parameter",
			parameters: []spec.Parameter{
				{
					Name:     "id",
					In:       "path",
					Required: true,
					Schema:   &spec.Schema{Type: "integer"},
				},
			},
			pathValues: map[string]string{"id": "abc"},
			wantError:  true,
			errorMsg:   "Must be an integer",
		},
		{
			name: "missing required parameter",
			parameters: []spec.Parameter{
				{
					Name:     "id",
					In:       "path",
					Required: true,
					Schema:   &spec.Schema{Type: "string"},
				},
			},
			pathValues: map[string]string{},
			wantError:  true,
			errorMsg:   "Required path parameter is missing",
		},
		{
			name: "integer below minimum",
			parameters: []spec.Parameter{
				{
					Name:     "id",
					In:       "path",
					Required: true,
					Schema: &spec.Schema{
						Type:    "integer",
						Minimum: float64Ptr(1),
					},
				},
			},
			pathValues: map[string]string{"id": "0"},
			wantError:  true,
			errorMsg:   "Must be at least 1",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			apiSpec := &spec.OpenAPISpec{}
			v := New(apiSpec)

			req := httptest.NewRequest("GET", "/test", nil)

			// Simulate path values
			for key, value := range tt.pathValues {
				req.SetPathValue(key, value)
			}

			operation := &spec.Operation{
				Parameters: tt.parameters,
			}

			err := v.ValidatePathParameters(req, operation)

			if tt.wantError {
				if err == nil {
					t.Errorf("expected error but got none")
					return
				}
				if tt.errorMsg != "" && !contains(err.Error(), tt.errorMsg) {
					t.Errorf("expected error containing %q, got %q", tt.errorMsg, err.Error())
				}
			} else {
				if err != nil {
					t.Errorf("unexpected error: %v", err)
				}
			}
		})
	}
}

func TestValidateQueryParameters(t *testing.T) {
	tests := []struct {
		name       string
		parameters []spec.Parameter
		queryURL   string
		wantError  bool
		errorMsg   string
	}{
		{
			name: "valid query parameter",
			parameters: []spec.Parameter{
				{
					Name:     "limit",
					In:       "query",
					Required: false,
					Schema:   &spec.Schema{Type: "integer"},
				},
			},
			queryURL:  "/test?limit=10",
			wantError: false,
		},
		{
			name: "invalid query parameter type",
			parameters: []spec.Parameter{
				{
					Name:     "limit",
					In:       "query",
					Required: true,
					Schema:   &spec.Schema{Type: "integer"},
				},
			},
			queryURL:  "/test?limit=abc",
			wantError: true,
			errorMsg:  "Must be an integer",
		},
		{
			name: "missing required query parameter",
			parameters: []spec.Parameter{
				{
					Name:     "limit",
					In:       "query",
					Required: true,
					Schema:   &spec.Schema{Type: "integer"},
				},
			},
			queryURL:  "/test",
			wantError: true,
			errorMsg:  "Required query parameter is missing",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			apiSpec := &spec.OpenAPISpec{}
			v := New(apiSpec)

			req := httptest.NewRequest("GET", tt.queryURL, nil)

			operation := &spec.Operation{
				Parameters: tt.parameters,
			}

			err := v.ValidateQueryParameters(req, operation)

			if tt.wantError {
				if err == nil {
					t.Errorf("expected error but got none")
					return
				}
				if tt.errorMsg != "" && !contains(err.Error(), tt.errorMsg) {
					t.Errorf("expected error containing %q, got %q", tt.errorMsg, err.Error())
				}
			} else {
				if err != nil {
					t.Errorf("unexpected error: %v", err)
				}
			}
		})
	}
}

func TestValidateEnum(t *testing.T) {
	apiSpec := &spec.OpenAPISpec{}
	v := New(apiSpec)

	reqBody := &spec.RequestBody{
		Required: true,
		Content: map[string]spec.MediaTypeObject{
			"application/json": {
				Schema: &spec.Schema{
					Type: "object",
					Properties: map[string]*spec.Schema{
						"status": {
							Type: "string",
							Enum: []interface{}{"active", "inactive", "pending"},
						},
					},
				},
			},
		},
	}

	// Valid enum value
	req := httptest.NewRequest("POST", "/test", bytes.NewBufferString(`{"status": "active"}`))
	req.Header.Set("Content-Type", "application/json")

	operation := &spec.Operation{RequestBody: reqBody}
	err := v.ValidateRequestBody(req, operation)
	if err != nil {
		t.Errorf("unexpected error for valid enum: %v", err)
	}

	// Invalid enum value
	req = httptest.NewRequest("POST", "/test", bytes.NewBufferString(`{"status": "invalid"}`))
	req.Header.Set("Content-Type", "application/json")
	req.Body = io.NopCloser(bytes.NewBufferString(`{"status": "invalid"}`))

	err = v.ValidateRequestBody(req, operation)
	if err == nil {
		t.Errorf("expected error for invalid enum value")
	}
}

func TestValidatePattern(t *testing.T) {
	apiSpec := &spec.OpenAPISpec{}
	v := New(apiSpec)

	reqBody := &spec.RequestBody{
		Required: true,
		Content: map[string]spec.MediaTypeObject{
			"application/json": {
				Schema: &spec.Schema{
					Type: "object",
					Properties: map[string]*spec.Schema{
						"code": {
							Type:    "string",
							Pattern: "^[A-Z]{3}$",
						},
					},
				},
			},
		},
	}

	// Valid pattern
	req := httptest.NewRequest("POST", "/test", bytes.NewBufferString(`{"code": "ABC"}`))
	req.Header.Set("Content-Type", "application/json")

	operation := &spec.Operation{RequestBody: reqBody}
	err := v.ValidateRequestBody(req, operation)
	if err != nil {
		t.Errorf("unexpected error for valid pattern: %v", err)
	}

	// Invalid pattern
	req = httptest.NewRequest("POST", "/test", bytes.NewBufferString(`{"code": "abc"}`))
	req.Header.Set("Content-Type", "application/json")

	err = v.ValidateRequestBody(req, operation)
	if err == nil {
		t.Errorf("expected error for invalid pattern")
	}
}

// Helper functions
func intPtr(i int) *int {
	return &i
}

func float64Ptr(f float64) *float64 {
	return &f
}

func contains(s, substr string) bool {
	return bytes.Contains([]byte(s), []byte(substr))
}
