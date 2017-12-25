package mimicro

import (
	"errors"
	"fmt"
	"io/ioutil"
	"path/filepath"

	"github.com/ghodss/yaml"
	"github.com/xeipuuv/gojsonschema"
)

var ConfigPath = ""

// ServerCollection сontains a full configuration of servers
type ServerCollection struct {
	Servers []MockServer `json:"servers"`
}

func ValidateSchema(data []byte, schema string) error {
	schemaLoader := gojsonschema.NewStringLoader(schema)
	documentLoader := gojsonschema.NewStringLoader(string(data))

	result, err := gojsonschema.Validate(schemaLoader, documentLoader)
	if err != nil {
		return err
	}

	if result.Valid() {
		return nil
	}

	var errorString string
	for _, desc := range result.Errors() {
		errorString = fmt.Sprintf("%s%s\n", errorString, desc)
	}

	return errors.New(errorString)
}

func parseConfig(data []byte) (*ServerCollection, error) {
	var serverCollection ServerCollection

	err := yaml.Unmarshal(data, &serverCollection)

	if err != nil {
		return nil, err
	}

	return &serverCollection, nil
}

// LoadConfig function loads the config from file into the ServerCollection structure
// Returns ServerCollection structure
func LoadConfig(configPath string) (*ServerCollection, error) {
	data, err := ioutil.ReadFile(configPath)

	if err != nil {
		return nil, err
	}

	ConfigPath, err = filepath.Abs(configPath)

	if err != nil {
		return nil, err
	}

	return parseConfig(data)
}

// CheckConfig checks the json schema of passed config file
func CheckConfig(configPath string) error {
	data, err := ioutil.ReadFile(configPath)

	if err != nil {
		return err
	}

	jsonData, err := yaml.YAMLToJSON(data)
	if err != nil {
		return err
	}

	return ValidateSchema(jsonData, ConfigSchema)
}
