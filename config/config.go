package config

import (
	"github.com/ghodss/yaml"
	"github.com/pokidovea/mimicro/mock_server"
	"io/ioutil"
)

type MockServerCollection struct {
	Servers []mock_server.MockServer `json:"servers"`
}

func Load(configPath string) (*MockServerCollection, error) {
	data, err := ioutil.ReadFile(configPath)

	if err != nil {
		return nil, err
	}

	var server_collection MockServerCollection

	err = yaml.Unmarshal(data, &server_collection)

	if err != nil {
		return nil, err
	}

	return &server_collection, nil
}
