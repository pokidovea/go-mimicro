package config

import (
	"github.com/ghodss/yaml"
	"io/ioutil"
)

type Endpoint struct {
	Url      string `json:"url"`
	Response string `json:"response"`
}

type Server struct {
	Name      string     `json:"name"`
	Port      int        `json:"port"`
	Endpoints []Endpoint `json:"endpoints"`
}

type ServerCollection struct {
	Servers []Server `json:"servers"`
}

func Load(configPath string) (*ServerCollection, error) {
	data, err := ioutil.ReadFile(configPath)

	if err != nil {
		return nil, err
	}

	var server_collection ServerCollection

	err = yaml.Unmarshal(data, &server_collection)

	if err != nil {
		return nil, err
	}

	return &server_collection, nil
}
