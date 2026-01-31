package spec

// OpenAPISpec represents a simplified OpenAPI 3.0 specification
type OpenAPISpec struct {
	OpenAPI    string              `json:"openapi" yaml:"openapi"`
	Info       Info                `json:"info" yaml:"info"`
	Paths      map[string]PathItem `json:"paths" yaml:"paths"`
	Components *Components         `json:"components,omitempty" yaml:"components,omitempty"`
}

type Info struct {
	Title   string `json:"title" yaml:"title"`
	Version string `json:"version" yaml:"version"`
}

type PathItem struct {
	Get    *Operation `json:"get,omitempty" yaml:"get,omitempty"`
	Post   *Operation `json:"post,omitempty" yaml:"post,omitempty"`
	Put    *Operation `json:"put,omitempty" yaml:"put,omitempty"`
	Delete *Operation `json:"delete,omitempty" yaml:"delete,omitempty"`
	Patch  *Operation `json:"patch,omitempty" yaml:"patch,omitempty"`
}

type Operation struct {
	OperationID string              `json:"operationId,omitempty" yaml:"operationId,omitempty"`
	Summary     string              `json:"summary,omitempty" yaml:"summary,omitempty"`
	Parameters  []Parameter         `json:"parameters,omitempty" yaml:"parameters,omitempty"`
	RequestBody *RequestBody        `json:"requestBody,omitempty" yaml:"requestBody,omitempty"`
	Responses   map[string]Response `json:"responses" yaml:"responses"`
}

type Parameter struct {
	Name     string  `json:"name" yaml:"name"`
	In       string  `json:"in" yaml:"in"` // query, path, header, cookie
	Required bool    `json:"required,omitempty" yaml:"required,omitempty"`
	Schema   *Schema `json:"schema,omitempty" yaml:"schema,omitempty"`
}

type RequestBody struct {
	Required bool                       `json:"required,omitempty" yaml:"required,omitempty"`
	Content  map[string]MediaTypeObject `json:"content" yaml:"content"`
}

type Response struct {
	Description string                     `json:"description" yaml:"description"`
	Content     map[string]MediaTypeObject `json:"content,omitempty" yaml:"content,omitempty"`
}

type MediaTypeObject struct {
	Schema *Schema `json:"schema,omitempty" yaml:"schema,omitempty"`
}

type Schema struct {
	Type       string             `json:"type,omitempty" yaml:"type,omitempty"`
	Format     string             `json:"format,omitempty" yaml:"format,omitempty"`
	Properties map[string]*Schema `json:"properties,omitempty" yaml:"properties,omitempty"`
	Items      *Schema            `json:"items,omitempty" yaml:"items,omitempty"`
	Required   []string           `json:"required,omitempty" yaml:"required,omitempty"`
	Enum       []interface{}      `json:"enum,omitempty" yaml:"enum,omitempty"`
	Minimum    *float64           `json:"minimum,omitempty" yaml:"minimum,omitempty"`
	Maximum    *float64           `json:"maximum,omitempty" yaml:"maximum,omitempty"`
	MinLength  *int               `json:"minLength,omitempty" yaml:"minLength,omitempty"`
	MaxLength  *int               `json:"maxLength,omitempty" yaml:"maxLength,omitempty"`
	Pattern    string             `json:"pattern,omitempty" yaml:"pattern,omitempty"`
	Ref        string             `json:"$ref,omitempty" yaml:"$ref,omitempty"`
	Example    interface{}        `json:"example,omitempty" yaml:"example,omitempty"`
}

type Components struct {
	Schemas map[string]*Schema `json:"schemas,omitempty" yaml:"schemas,omitempty"`
}
