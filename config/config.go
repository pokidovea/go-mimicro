package config

import (
	"errors"
	"fmt"
	"io/ioutil"
	"path/filepath"

	"github.com/ghodss/yaml"
	"github.com/pokidovea/mimicro/mock_server"
	"github.com/pokidovea/mimicro/settings"
	"github.com/xeipuuv/gojsonschema"
)

type MockServerCollection struct {
	Servers []mock_server.MockServer `json:"servers"`
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

	} else {
		var errorString string
		for _, desc := range result.Errors() {
			errorString = fmt.Sprintf("%s%s\n", errorString, desc)
		}

		return errors.New(errorString)
	}
}

func parseConfig(data []byte) (*MockServerCollection, error) {
	var serverCollection MockServerCollection

	err := yaml.Unmarshal(data, &serverCollection)

	if err != nil {
		return nil, err
	}

	return &serverCollection, nil
}

func Load(configPath string) (*MockServerCollection, error) {
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

func CheckConfig(configPath string) error {
	data, err := ioutil.ReadFile(configPath)

	if err != nil {
		return err
	}

	return validateSchema(data)
}
