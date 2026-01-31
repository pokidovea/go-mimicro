package generator

import (
	"fmt"
	"math/rand"
	"strings"
	"time"

	"mimicro/internal/spec"
)

type ResponseGenerator struct {
	spec *spec.OpenAPISpec
	rnd  *rand.Rand
}

func New(apiSpec *spec.OpenAPISpec) *ResponseGenerator {
	return &ResponseGenerator{
		spec: apiSpec,
		rnd:  rand.New(rand.NewSource(time.Now().UnixNano())),
	}
}

// Generate creates random data based on schema
func (g *ResponseGenerator) Generate(schema *spec.Schema) interface{} {
	if schema == nil {
		return nil
	}

	// Handle $ref
	if schema.Ref != "" {
		resolvedSchema := g.resolveRef(schema.Ref)
		if resolvedSchema != nil {
			return g.Generate(resolvedSchema)
		}
	}

	// Use example if provided
	if schema.Example != nil {
		return schema.Example
	}

	// Handle enum
	if len(schema.Enum) > 0 {
		return schema.Enum[g.rnd.Intn(len(schema.Enum))]
	}

	// Generate based on type
	switch schema.Type {
	case "object":
		return g.generateObject(schema)
	case "array":
		return g.generateArray(schema)
	case "string":
		return g.generateString(schema)
	case "integer":
		return g.generateInteger(schema)
	case "number":
		return g.generateNumber(schema)
	case "boolean":
		return g.rnd.Intn(2) == 1
	default:
		return nil
	}
}

func (g *ResponseGenerator) generateObject(schema *spec.Schema) map[string]interface{} {
	result := make(map[string]interface{})

	for propName, propSchema := range schema.Properties {
		result[propName] = g.Generate(propSchema)
	}

	return result
}

func (g *ResponseGenerator) generateArray(schema *spec.Schema) []interface{} {
	// Generate array with 1-5 items
	count := g.rnd.Intn(5) + 1
	result := make([]interface{}, count)

	for i := 0; i < count; i++ {
		result[i] = g.Generate(schema.Items)
	}

	return result
}

func (g *ResponseGenerator) generateString(schema *spec.Schema) string {
	// Handle format
	switch schema.Format {
	case "date":
		return time.Now().Format("2006-01-02")
	case "date-time":
		return time.Now().Format(time.RFC3339)
	case "email":
		return fmt.Sprintf("user%d@example.com", g.rnd.Intn(1000))
	case "uuid":
		return fmt.Sprintf("%08x-%04x-%04x-%04x-%012x",
			g.rnd.Uint32(), g.rnd.Uint32()&0xffff, g.rnd.Uint32()&0xffff,
			g.rnd.Uint32()&0xffff, g.rnd.Uint64()&0xffffffffffff)
	case "uri":
		return fmt.Sprintf("https://example.com/resource/%d", g.rnd.Intn(1000))
	}

	// Generate random string
	minLen := 5
	maxLen := 20

	if schema.MinLength != nil {
		minLen = *schema.MinLength
	}
	if schema.MaxLength != nil {
		maxLen = *schema.MaxLength
	}

	if minLen > maxLen {
		maxLen = minLen
	}

	length := minLen
	if maxLen > minLen {
		length = minLen + g.rnd.Intn(maxLen-minLen+1)
	}

	const letters = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	result := make([]byte, length)
	for i := range result {
		result[i] = letters[g.rnd.Intn(len(letters))]
	}

	return string(result)
}

func (g *ResponseGenerator) generateInteger(schema *spec.Schema) int {
	min := 0
	max := 1000

	if schema.Minimum != nil {
		min = int(*schema.Minimum)
	}
	if schema.Maximum != nil {
		max = int(*schema.Maximum)
	}

	if min > max {
		max = min
	}

	if max == min {
		return min
	}

	return min + g.rnd.Intn(max-min+1)
}

func (g *ResponseGenerator) generateNumber(schema *spec.Schema) float64 {
	min := 0.0
	max := 1000.0

	if schema.Minimum != nil {
		min = *schema.Minimum
	}
	if schema.Maximum != nil {
		max = *schema.Maximum
	}

	if min > max {
		max = min
	}

	return min + g.rnd.Float64()*(max-min)
}

func (g *ResponseGenerator) resolveRef(ref string) *spec.Schema {
	// Handle $ref like "#/components/schemas/User"
	if !strings.HasPrefix(ref, "#/components/schemas/") {
		return nil
	}

	schemaName := strings.TrimPrefix(ref, "#/components/schemas/")

	if g.spec.Components != nil && g.spec.Components.Schemas != nil {
		if schema, ok := g.spec.Components.Schemas[schemaName]; ok {
			return schema
		}
	}

	return nil
}
