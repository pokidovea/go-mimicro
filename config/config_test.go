package config

import (
	"fmt"
	"net/http"
	"path"
	"runtime"
	"testing"

	"github.com/pokidovea/mimicro/settings"
	"github.com/stretchr/testify/assert"
)

func TestSimpleConfig(t *testing.T) {
	config := `
servers:
- name: server_1
  port: 4573
  endpoints:
    - url: /simple_url
      GET:
        body: "{}"
        content_type: application/json
      POST:
        body: "OK"
        status_code: 201
    `

	err := validateSchema([]byte(config))
	assert.Nil(t, err)

	serverCollection, err := parseConfig([]byte(config))
	assert.Nil(t, err)
	assert.Equal(t, len(serverCollection.Servers), 1)

	server := serverCollection.Servers[0]
	assert.Equal(t, 4573, server.Port)
	assert.Equal(t, "server_1", server.Name)
	assert.Equal(t, 1, len(server.Endpoints))

	endpoint := server.Endpoints[0]
	assert.Equal(t, "/simple_url", endpoint.Url)

	getResponse := endpoint.GET
	assert.Equal(t, "{}", getResponse.Body)
	assert.Equal(t, "application/json", getResponse.ContentType)
	assert.Equal(t, http.StatusOK, getResponse.StatusCode)

	postResponse := endpoint.POST
	assert.Equal(t, "OK", postResponse.Body)
	assert.Equal(t, "text/plain", postResponse.ContentType)
	assert.Equal(t, http.StatusCreated, postResponse.StatusCode)

	patchResponse := endpoint.PATCH
	assert.Nil(t, patchResponse)

	putResponse := endpoint.PUT
	assert.Nil(t, putResponse)

	deleteResponse := endpoint.DELETE
	assert.Nil(t, deleteResponse)
}

func TestResponseBodyFromFileByAbsolutePath(t *testing.T) {
	_, filename, _, _ := runtime.Caller(0)
	filepath := path.Join(path.Dir(filename), "fixtures", "server_1_simple_response.json")

	config := fmt.Sprintf(`
servers:
- name: server_1
  port: 4573
  endpoints:
    - url: /response_from_file
      GET:
        body: file://%s
        content_type: application/json
    `, filepath)

	err := validateSchema([]byte(config))
	assert.Nil(t, err)

	serverCollection, err := parseConfig([]byte(config))
	assert.Nil(t, err)
	assert.Equal(t, len(serverCollection.Servers), 1)

	server := serverCollection.Servers[0]
	assert.Equal(t, 4573, server.Port)
	assert.Equal(t, "server_1", server.Name)
	assert.Equal(t, 1, len(server.Endpoints))

	endpoint := server.Endpoints[0]
	assert.Equal(t, "/response_from_file", endpoint.Url)

	getResponse := endpoint.GET
	assert.Equal(t, filepath, getResponse.Body)
	assert.Equal(t, "application/json", getResponse.ContentType)
	assert.Equal(t, http.StatusOK, getResponse.StatusCode)

	postResponse := endpoint.POST
	assert.Nil(t, postResponse)

	patchResponse := endpoint.PATCH
	assert.Nil(t, patchResponse)

	putResponse := endpoint.PUT
	assert.Nil(t, putResponse)

	deleteResponse := endpoint.DELETE
	assert.Nil(t, deleteResponse)
}

func TestResponseBodyFromFileByRelativePath(t *testing.T) {
	_, filename, _, _ := runtime.Caller(0)
	settings.CONFIG_PATH = path.Join(path.Dir(filename), "fixtures", "config.yaml")

	cases := []string{
		"./server_1_simple_response.json",
		"../fixtures/server_1_simple_response.json",
		"../../config/fixtures/server_1_simple_response.json",
		"server_1_simple_response.json",
	}

	for _, filepath := range cases {
		fullFilePath := path.Join(path.Dir(settings.CONFIG_PATH), filepath)

		config := fmt.Sprintf(`
            servers:
            - name: server_1
              port: 4573
              endpoints:
                - url: /response_from_file
                  GET:
                    body: file://%s
                    content_type: application/json
        `, filepath)

		err := validateSchema([]byte(config))
		assert.Nil(t, err)

		serverCollection, err := parseConfig([]byte(config))
		assert.Nil(t, err)
		assert.Equal(t, len(serverCollection.Servers), 1)

		server := serverCollection.Servers[0]
		assert.Equal(t, 4573, server.Port)
		assert.Equal(t, "server_1", server.Name)
		assert.Equal(t, 1, len(server.Endpoints))

		endpoint := server.Endpoints[0]
		assert.Equal(t, "/response_from_file", endpoint.Url)

		getResponse := endpoint.GET

		assert.Equal(t, fullFilePath, getResponse.Body)
		assert.Equal(t, "application/json", getResponse.ContentType)
		assert.Equal(t, http.StatusOK, getResponse.StatusCode)

		postResponse := endpoint.POST
		assert.Nil(t, postResponse)

		patchResponse := endpoint.PATCH
		assert.Nil(t, patchResponse)

		putResponse := endpoint.PUT
		assert.Nil(t, putResponse)

		deleteResponse := endpoint.DELETE
		assert.Nil(t, deleteResponse)
	}
}

func TestBinaryFile(t *testing.T) {
	_, filename, _, _ := runtime.Caller(0)
	settings.CONFIG_PATH = path.Join(path.Dir(filename), "fixtures", "config.yaml")
	filepath := "mimicro.png"
	fullFilePath := path.Join(path.Dir(settings.CONFIG_PATH), filepath)

	// we can set any content type for binary file, but if it is empty, it is autodetected while serving
	ctypes := []string{
		"",
		"application/json",
	}

	for _, ctype := range ctypes {
		// server ignores status code from config while serving file. It's always 200
		config := fmt.Sprintf(`
        servers:
        - name: server_1
          port: 4573
          endpoints:
            - url: /get_picture
              GET:
                body: file://%s
                status_code: 201
                content_type: "%s"
            `, filepath, ctype)

		err := validateSchema([]byte(config))
		assert.Nil(t, err)

		serverCollection, err := parseConfig([]byte(config))
		assert.Nil(t, err)
		assert.Equal(t, len(serverCollection.Servers), 1)

		server := serverCollection.Servers[0]
		assert.Equal(t, 4573, server.Port)
		assert.Equal(t, "server_1", server.Name)
		assert.Equal(t, 1, len(server.Endpoints))

		endpoint := server.Endpoints[0]
		assert.Equal(t, "/get_picture", endpoint.Url)

		getResponse := endpoint.GET
		assert.Equal(t, fullFilePath, getResponse.Body)
		assert.Equal(t, ctype, getResponse.ContentType)
		assert.Equal(t, http.StatusOK, getResponse.StatusCode)

		postResponse := endpoint.POST
		assert.Nil(t, postResponse)

		patchResponse := endpoint.PATCH
		assert.Nil(t, patchResponse)

		putResponse := endpoint.PUT
		assert.Nil(t, putResponse)

		deleteResponse := endpoint.DELETE
		assert.Nil(t, deleteResponse)
	}
}
