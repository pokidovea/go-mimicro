package spec

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"

	"gopkg.in/yaml.v3"
)

type Loader struct{}

func NewLoader() *Loader {
	return &Loader{}
}

// Load reads OpenAPI spec from file path or URL
func (l *Loader) Load(path string) (*OpenAPISpec, error) {
	var data []byte
	var err error

	if strings.HasPrefix(path, "http://") || strings.HasPrefix(path, "https://") {
		data, err = l.loadFromURL(path)
	} else {
		data, err = l.loadFromFile(path)
	}

	if err != nil {
		return nil, err
	}

	return l.parse(data, path)
}

func (l *Loader) loadFromFile(path string) ([]byte, error) {
	return os.ReadFile(path)
}

func (l *Loader) loadFromURL(url string) ([]byte, error) {
	resp, err := http.Get(url)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch URL: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("HTTP error: %s", resp.Status)
	}

	return io.ReadAll(resp.Body)
}

func (l *Loader) parse(data []byte, path string) (*OpenAPISpec, error) {
	var spec OpenAPISpec

	// Try JSON first
	if err := json.Unmarshal(data, &spec); err == nil {
		return &spec, nil
	}

	// Try YAML
	if err := yaml.Unmarshal(data, &spec); err != nil {
		return nil, fmt.Errorf("failed to parse as JSON or YAML: %w", err)
	}

	return &spec, nil
}
