package mockServer

import (
	"errors"
	"fmt"
	"io/ioutil"
	"path/filepath"

	"github.com/ghodss/yaml"
	"github.com/pokidovea/mimicro/settings"
	"github.com/xeipuuv/gojsonschema"
)

// ServerCollection сontains a full configuration of servers
type ServerCollection struct {
	CollectStatistics bool         `json:"collect_statistics"`
	Servers           []MockServer `json:"servers"`
}

func validateSchema(data []byte) error {
	jsonData, err := yaml.YAMLToJSON(data)
	if err != nil {
		return err
	}

	schemaLoader := gojsonschema.NewStringLoader(schema)
	documentLoader := gojsonschema.NewStringLoader(string(jsonData))

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

// Load function loads the config from file into the ServerCollection structure
// Returns ServerCollection structure
func Load(configPath string) (*ServerCollection, error) {
	data, err := ioutil.ReadFile(configPath)

	if err != nil {
		return nil, err
	}

	settings.CONFIG_PATH, err = filepath.Abs(configPath)

	if err != nil {
		return nil, err
	}

	return parseConfig(data)
}

// CheckConfig checks the json schema of pased config file
func CheckConfig(configPath string) error {
	data, err := ioutil.ReadFile(configPath)

	if err != nil {
		return err
	}

	return validateSchema(data)
}
