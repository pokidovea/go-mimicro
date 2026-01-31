package generator

import (
	"testing"

	"mimicro/internal/spec"
)

func TestGenerateString(t *testing.T) {
	apiSpec := &spec.OpenAPISpec{}
	g := New(apiSpec)

	tests := []struct {
		name   string
		schema *spec.Schema
		check  func(interface{}) bool
	}{
		{
			name:   "basic string",
			schema: &spec.Schema{Type: "string"},
			check: func(v interface{}) bool {
				s, ok := v.(string)
				return ok && len(s) >= 5 && len(s) <= 20
			},
		},
		{
			name: "string with minLength",
			schema: &spec.Schema{
				Type:      "string",
				MinLength: intPtr(10),
			},
			check: func(v interface{}) bool {
				s, ok := v.(string)
				return ok && len(s) >= 10
			},
		},
		{
			name: "string with maxLength",
			schema: &spec.Schema{
				Type:      "string",
				MaxLength: intPtr(8),
			},
			check: func(v interface{}) bool {
				s, ok := v.(string)
				return ok && len(s) <= 8
			},
		},
		{
			name: "string with format email",
			schema: &spec.Schema{
				Type:   "string",
				Format: "email",
			},
			check: func(v interface{}) bool {
				s, ok := v.(string)
				return ok && len(s) > 0 && containsStr(s, "@")
			},
		},
		{
			name: "string with format uuid",
			schema: &spec.Schema{
				Type:   "string",
				Format: "uuid",
			},
			check: func(v interface{}) bool {
				s, ok := v.(string)
				return ok && len(s) == 36 && countChar(s, '-') == 4
			},
		},
		{
			name: "string with format date",
			schema: &spec.Schema{
				Type:   "string",
				Format: "date",
			},
			check: func(v interface{}) bool {
				s, ok := v.(string)
				return ok && len(s) == 10 && countChar(s, '-') == 2
			},
		},
		{
			name: "string with format date-time",
			schema: &spec.Schema{
				Type:   "string",
				Format: "date-time",
			},
			check: func(v interface{}) bool {
				s, ok := v.(string)
				return ok && containsStr(s, "T")
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := g.Generate(tt.schema)
			if !tt.check(result) {
				t.Errorf("generated value %v doesn't match expectations", result)
			}
		})
	}
}

func TestGenerateInteger(t *testing.T) {
	apiSpec := &spec.OpenAPISpec{}
	g := New(apiSpec)

	tests := []struct {
		name   string
		schema *spec.Schema
		check  func(interface{}) bool
	}{
		{
			name:   "basic integer",
			schema: &spec.Schema{Type: "integer"},
			check: func(v interface{}) bool {
				i, ok := v.(int)
				return ok && i >= 0 && i <= 1000
			},
		},
		{
			name: "integer with minimum",
			schema: &spec.Schema{
				Type:    "integer",
				Minimum: float64Ptr(100),
			},
			check: func(v interface{}) bool {
				i, ok := v.(int)
				return ok && i >= 100
			},
		},
		{
			name: "integer with maximum",
			schema: &spec.Schema{
				Type:    "integer",
				Maximum: float64Ptr(50),
			},
			check: func(v interface{}) bool {
				i, ok := v.(int)
				return ok && i <= 50
			},
		},
		{
			name: "integer with min and max",
			schema: &spec.Schema{
				Type:    "integer",
				Minimum: float64Ptr(10),
				Maximum: float64Ptr(20),
			},
			check: func(v interface{}) bool {
				i, ok := v.(int)
				return ok && i >= 10 && i <= 20
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := g.Generate(tt.schema)
			if !tt.check(result) {
				t.Errorf("generated value %v doesn't match expectations", result)
			}
		})
	}
}

func TestGenerateNumber(t *testing.T) {
	apiSpec := &spec.OpenAPISpec{}
	g := New(apiSpec)

	tests := []struct {
		name   string
		schema *spec.Schema
		check  func(interface{}) bool
	}{
		{
			name:   "basic number",
			schema: &spec.Schema{Type: "number"},
			check: func(v interface{}) bool {
				f, ok := v.(float64)
				return ok && f >= 0.0 && f <= 1000.0
			},
		},
		{
			name: "number with minimum",
			schema: &spec.Schema{
				Type:    "number",
				Minimum: float64Ptr(50.5),
			},
			check: func(v interface{}) bool {
				f, ok := v.(float64)
				return ok && f >= 50.5
			},
		},
		{
			name: "number with maximum",
			schema: &spec.Schema{
				Type:    "number",
				Maximum: float64Ptr(100.5),
			},
			check: func(v interface{}) bool {
				f, ok := v.(float64)
				return ok && f <= 100.5
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := g.Generate(tt.schema)
			if !tt.check(result) {
				t.Errorf("generated value %v doesn't match expectations", result)
			}
		})
	}
}

func TestGenerateBoolean(t *testing.T) {
	apiSpec := &spec.OpenAPISpec{}
	g := New(apiSpec)

	schema := &spec.Schema{Type: "boolean"}
	result := g.Generate(schema)

	_, ok := result.(bool)
	if !ok {
		t.Errorf("expected boolean, got %T", result)
	}
}

func TestGenerateObject(t *testing.T) {
	apiSpec := &spec.OpenAPISpec{}
	g := New(apiSpec)

	schema := &spec.Schema{
		Type: "object",
		Properties: map[string]*spec.Schema{
			"name":   {Type: "string"},
			"age":    {Type: "integer"},
			"active": {Type: "boolean"},
		},
	}

	result := g.Generate(schema)
	obj, ok := result.(map[string]interface{})
	if !ok {
		t.Fatalf("expected object, got %T", result)
	}

	if _, exists := obj["name"]; !exists {
		t.Error("missing 'name' property")
	}
	if _, exists := obj["age"]; !exists {
		t.Error("missing 'age' property")
	}
	if _, exists := obj["active"]; !exists {
		t.Error("missing 'active' property")
	}

	if _, ok := obj["name"].(string); !ok {
		t.Error("'name' should be string")
	}
	if _, ok := obj["age"].(int); !ok {
		t.Error("'age' should be integer")
	}
	if _, ok := obj["active"].(bool); !ok {
		t.Error("'active' should be boolean")
	}
}

func TestGenerateArray(t *testing.T) {
	apiSpec := &spec.OpenAPISpec{}
	g := New(apiSpec)

	schema := &spec.Schema{
		Type: "array",
		Items: &spec.Schema{
			Type: "string",
		},
	}

	result := g.Generate(schema)
	arr, ok := result.([]interface{})
	if !ok {
		t.Fatalf("expected array, got %T", result)
	}

	if len(arr) < 1 || len(arr) > 5 {
		t.Errorf("expected array length between 1 and 5, got %d", len(arr))
	}

	for i, item := range arr {
		if _, ok := item.(string); !ok {
			t.Errorf("item %d should be string, got %T", i, item)
		}
	}
}

func TestGenerateEnum(t *testing.T) {
	apiSpec := &spec.OpenAPISpec{}
	g := New(apiSpec)

	schema := &spec.Schema{
		Type: "string",
		Enum: []interface{}{"active", "inactive", "pending"},
	}

	// Generate multiple times to check it's always one of enum values
	for i := 0; i < 10; i++ {
		result := g.Generate(schema)
		str, ok := result.(string)
		if !ok {
			t.Fatalf("expected string, got %T", result)
		}

		valid := str == "active" || str == "inactive" || str == "pending"
		if !valid {
			t.Errorf("generated value %q is not in enum", str)
		}
	}
}

func TestGenerateWithExample(t *testing.T) {
	apiSpec := &spec.OpenAPISpec{}
	g := New(apiSpec)

	schema := &spec.Schema{
		Type:    "string",
		Example: "example-value",
	}

	result := g.Generate(schema)
	if result != "example-value" {
		t.Errorf("expected 'example-value', got %v", result)
	}
}

func TestGenerateWithRef(t *testing.T) {
	apiSpec := &spec.OpenAPISpec{
		Components: &spec.Components{
			Schemas: map[string]*spec.Schema{
				"User": {
					Type: "object",
					Properties: map[string]*spec.Schema{
						"id":   {Type: "integer"},
						"name": {Type: "string"},
					},
				},
			},
		},
	}
	g := New(apiSpec)

	schema := &spec.Schema{
		Ref: "#/components/schemas/User",
	}

	result := g.Generate(schema)
	obj, ok := result.(map[string]interface{})
	if !ok {
		t.Fatalf("expected object, got %T", result)
	}

	if _, exists := obj["id"]; !exists {
		t.Error("missing 'id' property")
	}
	if _, exists := obj["name"]; !exists {
		t.Error("missing 'name' property")
	}
}

func TestGenerateNestedObject(t *testing.T) {
	apiSpec := &spec.OpenAPISpec{}
	g := New(apiSpec)

	schema := &spec.Schema{
		Type: "object",
		Properties: map[string]*spec.Schema{
			"user": {
				Type: "object",
				Properties: map[string]*spec.Schema{
					"name": {Type: "string"},
					"email": {
						Type:   "string",
						Format: "email",
					},
				},
			},
			"tags": {
				Type: "array",
				Items: &spec.Schema{
					Type: "string",
				},
			},
		},
	}

	result := g.Generate(schema)
	obj, ok := result.(map[string]interface{})
	if !ok {
		t.Fatalf("expected object, got %T", result)
	}

	user, ok := obj["user"].(map[string]interface{})
	if !ok {
		t.Fatal("'user' should be object")
	}

	if _, ok := user["name"].(string); !ok {
		t.Error("'user.name' should be string")
	}
	if email, ok := user["email"].(string); !ok {
		t.Error("'user.email' should be string")
	} else if !containsStr(email, "@") {
		t.Error("'user.email' should contain @")
	}

	tags, ok := obj["tags"].([]interface{})
	if !ok {
		t.Fatal("'tags' should be array")
	}

	if len(tags) < 1 {
		t.Error("'tags' array should not be empty")
	}
}

// Helper functions
func intPtr(i int) *int {
	return &i
}

func float64Ptr(f float64) *float64 {
	return &f
}

func containsStr(s, substr string) bool {
	return len(s) > 0 && len(substr) > 0 && contains(s, substr)
}

func contains(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

func countChar(s string, c rune) int {
	count := 0
	for _, ch := range s {
		if ch == c {
			count++
		}
	}
	return count
}
