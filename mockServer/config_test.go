package mockServer

import (
	"fmt"
	"net/http"
	"path"
	"runtime"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSimpleConfig(t *testing.T) {
	config := `
collect_statistics: true
servers:
- name: server_1
  port: 4573
  endpoints:
    - url: /simple_url
      GET:
        template: "{}"
        content_type: application/json
      POST:
        template: "OK"
        status_code: 201
    `

	err := validateSchema([]byte(config))
	assert.Nil(t, err)

	serverCollection, err := parseConfig([]byte(config))
	assert.Nil(t, err)
	assert.Equal(t, len(serverCollection.Servers), 1)
	assert.True(t, serverCollection.CollectStatistics)

	server := serverCollection.Servers[0]
	assert.Equal(t, 4573, server.Port)
	assert.Equal(t, "server_1", server.Name)
	assert.Equal(t, 1, len(server.Endpoints))

	endpoint := server.Endpoints[0]
	assert.Equal(t, "/simple_url", endpoint.Url)

	getResponse := endpoint.GET
	assert.NotNil(t, getResponse.template)
	assert.Equal(t, "application/json", getResponse.ContentType)
	assert.Equal(t, http.StatusOK, getResponse.StatusCode)

	postResponse := endpoint.POST
	assert.NotNil(t, postResponse.template)
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
	filepath := path.Join(path.Dir(filename), "..", "examples", "response_with_var.json")

	config := fmt.Sprintf(`
collect_statistics: false
servers:
- name: server_1
  port: 4573
  endpoints:
    - url: /response_from_file
      GET:
        file: file://%s
        content_type: application/json
    `, filepath)

	err := validateSchema([]byte(config))
	assert.Nil(t, err)

	serverCollection, err := parseConfig([]byte(config))
	assert.Nil(t, err)
	assert.Equal(t, len(serverCollection.Servers), 1)
	assert.False(t, serverCollection.CollectStatistics)

	server := serverCollection.Servers[0]
	assert.Equal(t, 4573, server.Port)
	assert.Equal(t, "server_1", server.Name)
	assert.Equal(t, 1, len(server.Endpoints))

	endpoint := server.Endpoints[0]
	assert.Equal(t, "/response_from_file", endpoint.Url)

	getResponse := endpoint.GET
	assert.Equal(t, filepath, getResponse.file)
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
	configPath = path.Join(path.Dir(filename), "..", "examples", "config.yaml")

	cases := []string{
		"./response_with_var.json",
		"../examples/response_with_var.json",
		"../../mimicro/examples/response_with_var.json",
		"response_with_var.json",
	}

	for _, filepath := range cases {
		fullFilePath := path.Join(path.Dir(configPath), filepath)

		config := fmt.Sprintf(`
            collect_statistics: false
            servers:
            - name: server_1
              port: 4573
              endpoints:
                - url: /response_from_file
                  GET:
                    file: file://%s
                    content_type: application/json
        `, filepath)

		err := validateSchema([]byte(config))
		assert.Nil(t, err)

		serverCollection, err := parseConfig([]byte(config))
		assert.Nil(t, err)
		assert.Equal(t, len(serverCollection.Servers), 1)
		assert.False(t, serverCollection.CollectStatistics)

		server := serverCollection.Servers[0]
		assert.Equal(t, 4573, server.Port)
		assert.Equal(t, "server_1", server.Name)
		assert.Equal(t, 1, len(server.Endpoints))

		endpoint := server.Endpoints[0]
		assert.Equal(t, "/response_from_file", endpoint.Url)

		getResponse := endpoint.GET

		assert.Equal(t, fullFilePath, getResponse.file)
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
	configPath = path.Join(path.Dir(filename), "..", "examples", "config.yaml")
	filepath := "mimicro.png"
	fullFilePath := path.Join(path.Dir(configPath), filepath)

	// we can set any content type for binary file, but if it is empty, it is autodetected while serving
	ctypes := []string{
		"",
		"application/json",
	}

	for _, ctype := range ctypes {
		// server ignores status code from config while serving file. It's always 200
		config := fmt.Sprintf(`
        collect_statistics: false
        servers:
        - name: server_1
          port: 4573
          endpoints:
            - url: /get_picture
              GET:
                file: file://%s
                status_code: 200
                content_type: "%s"
            `, filepath, ctype)

		err := validateSchema([]byte(config))
		assert.Nil(t, err)

		serverCollection, err := parseConfig([]byte(config))
		assert.Nil(t, err)
		assert.Equal(t, len(serverCollection.Servers), 1)
		assert.False(t, serverCollection.CollectStatistics)

		server := serverCollection.Servers[0]
		assert.Equal(t, 4573, server.Port)
		assert.Equal(t, "server_1", server.Name)
		assert.Equal(t, 1, len(server.Endpoints))

		endpoint := server.Endpoints[0]
		assert.Equal(t, "/get_picture", endpoint.Url)

		getResponse := endpoint.GET
		assert.Equal(t, fullFilePath, getResponse.file)
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
