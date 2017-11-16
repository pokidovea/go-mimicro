package mockServer

import (
	"net/http"
	"path"
	"runtime"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestValidateCorrectConfigFromFile(t *testing.T) {
	_, filename, _, _ := runtime.Caller(0)
	configPath = path.Join(path.Dir(filename), "..", "examples", "config.yaml")
	err := CheckConfig(configPath)

	assert.Nil(t, err)
}

func TestValidateNonexistingConfigFile(t *testing.T) {
	err := CheckConfig("/wrong_file.yaml")

	assert.NotNil(t, err)
	assert.Equal(t, "open /wrong_file.yaml: no such file or directory", err.Error())
}

func TestValidateWrongConfig(t *testing.T) {
	config := `
servers:
  - name: server_1
    port: 4573
    endpoints:
      - url: /simple_url
        GET:
          template: "{}"
          content_type: application/json
          status_code: "201"
    `
	expectedError := `servers.0.endpoints.0.GET: Must validate one and only one schema (oneOf)
servers.0.endpoints.0.GET.status_code: Invalid type. Expected: integer, given: string
`
	err := validateSchema([]byte(config))

	assert.NotNil(t, err)
	assert.Equal(t, expectedError, err.Error())
}

func TestLoadConfigFromFile(t *testing.T) {
	_, filename, _, _ := runtime.Caller(0)
	configPath = path.Join(path.Dir(filename), "..", "examples", "config.yaml")

	serverCollection, err := Load(configPath)

	assert.Nil(t, err)
	assert.Equal(t, len(serverCollection.Servers), 1)
	assert.True(t, serverCollection.CollectStatistics)

	server := serverCollection.Servers[0]
	assert.Equal(t, 4573, server.Port)
	assert.Equal(t, "server_1", server.Name)
	assert.Equal(t, 5, len(server.Endpoints))

	// endpoint 0
	endpoint := server.Endpoints[0]
	assert.Equal(t, "/simple_url", endpoint.URL)

	getResponse := endpoint.GET
	assert.NotNil(t, getResponse.template)
	assert.Nil(t, getResponse.file)
	assert.Equal(t, "application/json", getResponse.ContentType)
	assert.Equal(t, http.StatusOK, getResponse.StatusCode)

	postResponse := endpoint.POST
	assert.NotNil(t, postResponse.template)
	assert.Nil(t, postResponse.file)
	assert.Equal(t, "text/plain", postResponse.ContentType)
	assert.Equal(t, http.StatusCreated, postResponse.StatusCode)

	patchResponse := endpoint.PATCH
	assert.Nil(t, patchResponse)

	putResponse := endpoint.PUT
	assert.Nil(t, putResponse)

	deleteResponse := endpoint.DELETE
	assert.Nil(t, deleteResponse)

	// endpoint 1
	endpoint = server.Endpoints[1]
	assert.Equal(t, "/picture", endpoint.URL)

	getResponse = endpoint.GET
	assert.Nil(t, getResponse.template)
	assert.NotNil(t, getResponse.file)
	assert.Equal(t, "", getResponse.ContentType)
	assert.Equal(t, http.StatusOK, getResponse.StatusCode)

	postResponse = endpoint.POST
	assert.Nil(t, postResponse)

	patchResponse = endpoint.PATCH
	assert.Nil(t, patchResponse)

	putResponse = endpoint.PUT
	assert.Nil(t, putResponse)

	deleteResponse = endpoint.DELETE
	assert.Nil(t, deleteResponse)

	// endpoint 2
	endpoint = server.Endpoints[2]
	assert.Equal(t, "/{var}/in/filepath", endpoint.URL)

	getResponse = endpoint.GET
	assert.Nil(t, getResponse.template)
	assert.NotNil(t, getResponse.file)
	assert.Equal(t, "", getResponse.ContentType)
	assert.Equal(t, http.StatusOK, getResponse.StatusCode)

	postResponse = endpoint.POST
	assert.Nil(t, postResponse)

	patchResponse = endpoint.PATCH
	assert.Nil(t, patchResponse)

	putResponse = endpoint.PUT
	assert.Nil(t, putResponse)

	deleteResponse = endpoint.DELETE
	assert.Nil(t, deleteResponse)

	// endpoint 3
	endpoint = server.Endpoints[3]
	assert.Equal(t, "/template_from_file/{var}", endpoint.URL)

	getResponse = endpoint.GET
	assert.Nil(t, getResponse)

	postResponse = endpoint.POST
	assert.Nil(t, postResponse)

	patchResponse = endpoint.PATCH
	assert.Nil(t, patchResponse)

	putResponse = endpoint.PUT
	assert.NotNil(t, putResponse.template)
	assert.Nil(t, putResponse.file)
	assert.Equal(t, "application/json", putResponse.ContentType)
	assert.Equal(t, http.StatusOK, putResponse.StatusCode)

	deleteResponse = endpoint.DELETE
	assert.Nil(t, deleteResponse)

	// endpoint 4
	endpoint = server.Endpoints[4]
	assert.Equal(t, "/string_template/{var}", endpoint.URL)

	getResponse = endpoint.GET
	assert.Nil(t, getResponse)

	postResponse = endpoint.POST
	assert.Nil(t, postResponse)

	patchResponse = endpoint.PATCH
	assert.Nil(t, patchResponse)

	putResponse = endpoint.PUT
	assert.Nil(t, putResponse)

	deleteResponse = endpoint.DELETE
	assert.NotNil(t, deleteResponse.template)
	assert.Nil(t, deleteResponse.file)
	assert.Equal(t, "text/plain", deleteResponse.ContentType)
	assert.Equal(t, http.StatusForbidden, deleteResponse.StatusCode)
}
