package validator

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"regexp"
	"strconv"
	"strings"

	"mimicro/internal/spec"
)

type Validator struct {
	spec *spec.OpenAPISpec
}

func New(apiSpec *spec.OpenAPISpec) *Validator {
	return &Validator{
		spec: apiSpec,
	}
}

type ValidationError struct {
	Field   string
	Message string
}

func (e *ValidationError) Error() string {
	if e.Field != "" {
		return fmt.Sprintf("%s: %s", e.Field, e.Message)
	}
	return e.Message
}

// ValidateRequestBody validates the request body against the schema
func (v *Validator) ValidateRequestBody(r *http.Request, operation *spec.Operation) error {
	if operation.RequestBody == nil {
		return nil
	}

	contentType := r.Header.Get("Content-Type")
	if contentType == "" && operation.RequestBody.Required {
		return &ValidationError{Message: "Content-Type header is required"}
	}

	// Extract base content type (without charset, etc.)
	baseContentType := strings.Split(contentType, ";")[0]
	baseContentType = strings.TrimSpace(baseContentType)

	mediaType, ok := operation.RequestBody.Content[baseContentType]
	if !ok {
		// Try to find any JSON content type
		for ct, mt := range operation.RequestBody.Content {
			if strings.Contains(ct, "json") {
				mediaType = mt
				ok = true
				break
			}
		}
		if !ok {
			return &ValidationError{Message: fmt.Sprintf("Unsupported content type: %s", contentType)}
		}
	}

	if mediaType.Schema == nil {
		return nil
	}

	// Read and parse body
	body, err := io.ReadAll(r.Body)
	if err != nil {
		return &ValidationError{Message: "Failed to read request body"}
	}

	if len(body) == 0 && operation.RequestBody.Required {
		return &ValidationError{Message: "Request body is required"}
	}

	if len(body) == 0 {
		return nil
	}

	var data interface{}
	if err := json.Unmarshal(body, &data); err != nil {
		return &ValidationError{Message: "Invalid JSON: " + err.Error()}
	}

	// Validate against schema
	return v.validateValue(data, mediaType.Schema, "body")
}

// ValidatePathParameters validates path parameters
func (v *Validator) ValidatePathParameters(r *http.Request, operation *spec.Operation) error {
	if operation.Parameters == nil {
		return nil
	}

	for _, param := range operation.Parameters {
		if param.In != "path" {
			continue
		}

		// Extract path parameter value from URL
		value := r.PathValue(param.Name)

		if value == "" && param.Required {
			return &ValidationError{
				Field:   param.Name,
				Message: "Required path parameter is missing",
			}
		}

		if value == "" {
			continue
		}

		// Validate against schema
		if param.Schema != nil {
			if err := v.validatePathParam(value, param.Schema, param.Name); err != nil {
				return err
			}
		}
	}

	return nil
}

// ValidateQueryParameters validates query parameters
func (v *Validator) ValidateQueryParameters(r *http.Request, operation *spec.Operation) error {
	if operation.Parameters == nil {
		return nil
	}

	for _, param := range operation.Parameters {
		if param.In != "query" {
			continue
		}

		value := r.URL.Query().Get(param.Name)

		if value == "" && param.Required {
			return &ValidationError{
				Field:   param.Name,
				Message: "Required query parameter is missing",
			}
		}

		if value == "" {
			continue
		}

		// Validate against schema
		if param.Schema != nil {
			if err := v.validateQueryParam(value, param.Schema, param.Name); err != nil {
				return err
			}
		}
	}

	return nil
}

func (v *Validator) validatePathParam(value string, schema *spec.Schema, fieldName string) error {
	return v.validateStringParam(value, schema, fieldName)
}

func (v *Validator) validateQueryParam(value string, schema *spec.Schema, fieldName string) error {
	return v.validateStringParam(value, schema, fieldName)
}

func (v *Validator) validateStringParam(value string, schema *spec.Schema, fieldName string) error {
	// Resolve $ref if present
	if schema.Ref != "" {
		schema = v.resolveRef(schema.Ref)
		if schema == nil {
			return nil // Skip validation if ref can't be resolved
		}
	}

	switch schema.Type {
	case "integer":
		val, err := strconv.Atoi(value)
		if err != nil {
			return &ValidationError{Field: fieldName, Message: "Must be an integer"}
		}
		return v.validateInteger(val, schema, fieldName)

	case "number":
		val, err := strconv.ParseFloat(value, 64)
		if err != nil {
			return &ValidationError{Field: fieldName, Message: "Must be a number"}
		}
		return v.validateNumber(val, schema, fieldName)

	case "boolean":
		if value != "true" && value != "false" {
			return &ValidationError{Field: fieldName, Message: "Must be a boolean (true or false)"}
		}

	case "string":
		return v.validateString(value, schema, fieldName)
	}

	return nil
}

func (v *Validator) validateValue(value interface{}, schema *spec.Schema, fieldName string) error {
	if schema == nil {
		return nil
	}

	// Resolve $ref if present
	if schema.Ref != "" {
		schema = v.resolveRef(schema.Ref)
		if schema == nil {
			return nil // Skip validation if ref can't be resolved
		}
	}

	// Check enum
	if len(schema.Enum) > 0 {
		found := false
		for _, enumVal := range schema.Enum {
			if enumVal == value {
				found = true
				break
			}
		}
		if !found {
			return &ValidationError{Field: fieldName, Message: "Value must be one of the allowed enum values"}
		}
		return nil
	}

	switch schema.Type {
	case "object":
		return v.validateObject(value, schema, fieldName)
	case "array":
		return v.validateArray(value, schema, fieldName)
	case "string":
		str, ok := value.(string)
		if !ok {
			return &ValidationError{Field: fieldName, Message: "Must be a string"}
		}
		return v.validateString(str, schema, fieldName)
	case "integer":
		// JSON unmarshals numbers as float64
		num, ok := value.(float64)
		if !ok {
			return &ValidationError{Field: fieldName, Message: "Must be an integer"}
		}
		if num != float64(int(num)) {
			return &ValidationError{Field: fieldName, Message: "Must be an integer"}
		}
		return v.validateInteger(int(num), schema, fieldName)
	case "number":
		num, ok := value.(float64)
		if !ok {
			return &ValidationError{Field: fieldName, Message: "Must be a number"}
		}
		return v.validateNumber(num, schema, fieldName)
	case "boolean":
		_, ok := value.(bool)
		if !ok {
			return &ValidationError{Field: fieldName, Message: "Must be a boolean"}
		}
	}

	return nil
}

func (v *Validator) validateObject(value interface{}, schema *spec.Schema, fieldName string) error {
	obj, ok := value.(map[string]interface{})
	if !ok {
		return &ValidationError{Field: fieldName, Message: "Must be an object"}
	}

	// Check required fields
	for _, reqField := range schema.Required {
		if _, exists := obj[reqField]; !exists {
			return &ValidationError{
				Field:   fmt.Sprintf("%s.%s", fieldName, reqField),
				Message: "Required field is missing",
			}
		}
	}

	// Validate properties
	for propName, propSchema := range schema.Properties {
		if propValue, exists := obj[propName]; exists {
			fullFieldName := fmt.Sprintf("%s.%s", fieldName, propName)
			if err := v.validateValue(propValue, propSchema, fullFieldName); err != nil {
				return err
			}
		}
	}

	return nil
}

func (v *Validator) validateArray(value interface{}, schema *spec.Schema, fieldName string) error {
	arr, ok := value.([]interface{})
	if !ok {
		return &ValidationError{Field: fieldName, Message: "Must be an array"}
	}

	if schema.Items != nil {
		for i, item := range arr {
			itemFieldName := fmt.Sprintf("%s[%d]", fieldName, i)
			if err := v.validateValue(item, schema.Items, itemFieldName); err != nil {
				return err
			}
		}
	}

	return nil
}

func (v *Validator) validateString(value string, schema *spec.Schema, fieldName string) error {
	// Check length constraints
	if schema.MinLength != nil && len(value) < *schema.MinLength {
		return &ValidationError{
			Field:   fieldName,
			Message: fmt.Sprintf("Must be at least %d characters long", *schema.MinLength),
		}
	}

	if schema.MaxLength != nil && len(value) > *schema.MaxLength {
		return &ValidationError{
			Field:   fieldName,
			Message: fmt.Sprintf("Must be at most %d characters long", *schema.MaxLength),
		}
	}

	// Check pattern
	if schema.Pattern != "" {
		matched, err := regexp.MatchString(schema.Pattern, value)
		if err != nil {
			return &ValidationError{Field: fieldName, Message: "Invalid pattern in schema"}
		}
		if !matched {
			return &ValidationError{
				Field:   fieldName,
				Message: fmt.Sprintf("Must match pattern: %s", schema.Pattern),
			}
		}
	}

	// Validate format
	if schema.Format != "" {
		if err := v.validateFormat(value, schema.Format, fieldName); err != nil {
			return err
		}
	}

	return nil
}

func (v *Validator) validateInteger(value int, schema *spec.Schema, fieldName string) error {
	if schema.Minimum != nil && float64(value) < *schema.Minimum {
		return &ValidationError{
			Field:   fieldName,
			Message: fmt.Sprintf("Must be at least %v", *schema.Minimum),
		}
	}

	if schema.Maximum != nil && float64(value) > *schema.Maximum {
		return &ValidationError{
			Field:   fieldName,
			Message: fmt.Sprintf("Must be at most %v", *schema.Maximum),
		}
	}

	return nil
}

func (v *Validator) validateNumber(value float64, schema *spec.Schema, fieldName string) error {
	if schema.Minimum != nil && value < *schema.Minimum {
		return &ValidationError{
			Field:   fieldName,
			Message: fmt.Sprintf("Must be at least %v", *schema.Minimum),
		}
	}

	if schema.Maximum != nil && value > *schema.Maximum {
		return &ValidationError{
			Field:   fieldName,
			Message: fmt.Sprintf("Must be at most %v", *schema.Maximum),
		}
	}

	return nil
}

func (v *Validator) validateFormat(value, format, fieldName string) error {
	switch format {
	case "email":
		if !strings.Contains(value, "@") {
			return &ValidationError{Field: fieldName, Message: "Must be a valid email address"}
		}
	case "uuid":
		uuidPattern := `^[0-9a-fA-F]{8}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{12}$`
		matched, _ := regexp.MatchString(uuidPattern, value)
		if !matched {
			return &ValidationError{Field: fieldName, Message: "Must be a valid UUID"}
		}
	case "uri", "url":
		if !strings.HasPrefix(value, "http://") && !strings.HasPrefix(value, "https://") {
			return &ValidationError{Field: fieldName, Message: "Must be a valid URL"}
		}
	case "date":
		datePattern := `^\d{4}-\d{2}-\d{2}$`
		matched, _ := regexp.MatchString(datePattern, value)
		if !matched {
			return &ValidationError{Field: fieldName, Message: "Must be a valid date (YYYY-MM-DD)"}
		}
	case "date-time":
		// Basic RFC3339 validation
		if !strings.Contains(value, "T") || (!strings.Contains(value, "Z") && !strings.Contains(value, "+") && !strings.Contains(value, "-")) {
			return &ValidationError{Field: fieldName, Message: "Must be a valid date-time (RFC3339)"}
		}
	}

	return nil
}

func (v *Validator) resolveRef(ref string) *spec.Schema {
	if !strings.HasPrefix(ref, "#/components/schemas/") {
		return nil
	}

	schemaName := strings.TrimPrefix(ref, "#/components/schemas/")

	if v.spec.Components != nil && v.spec.Components.Schemas != nil {
		if schema, ok := v.spec.Components.Schemas[schemaName]; ok {
			return schema
		}
	}

	return nil
}
