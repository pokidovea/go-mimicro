package config

import (
	"errors"
	"fmt"
	"github.com/ghodss/yaml"
	"github.com/pokidovea/mimicro/mock_server"
	"github.com/xeipuuv/gojsonschema"
	"io/ioutil"
)

type MockServerCollection struct {
	Servers []mock_server.MockServer `json:"servers"`
}

func validateSchema(data string) error {
	schemaLoader := gojsonschema.NewStringLoader(schema)
	documentLoader := gojsonschema.NewStringLoader(data)

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

	return parseConfig(data)
}

func CheckConfig(configPath string) error {
	data, err := ioutil.ReadFile(configPath)

	if err != nil {
		return err
	}

	return validateSchema(string(data))
}
