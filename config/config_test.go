package config

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"path"
	"runtime"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSimpleConfig(t *testing.T) {
	config := `
collectStats: true
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
	assert.True(t, serverCollection.CollectStat)

	server := serverCollection.Servers[0]
	assert.Equal(t, 4573, server.Port)
	assert.Equal(t, "server_1", server.Name)
	assert.Equal(t, 1, len(server.Endpoints))

	endpoint := server.Endpoints[0]
	assert.Equal(t, "/simple_url", endpoint.Url)

	get_response := endpoint.GET
	assert.Equal(t, "{}", get_response.Body)
	assert.Equal(t, "application/json", get_response.ContentType)
	assert.Equal(t, http.StatusOK, get_response.StatusCode)

	post_response := endpoint.POST
	assert.Equal(t, "OK", post_response.Body)
	assert.Equal(t, "text/plain", post_response.ContentType)
	assert.Equal(t, http.StatusCreated, post_response.StatusCode)

	patch_response := endpoint.PATCH
	assert.Nil(t, patch_response)

	put_response := endpoint.PUT
	assert.Nil(t, put_response)

	delete_response := endpoint.DELETE
	assert.Nil(t, delete_response)
}

func TestResponseBodyFromFile(t *testing.T) {
	_, filename, _, _ := runtime.Caller(0)
	filepath := path.Join(path.Dir(filename), "fixtures", "server_1_simple_response.json")

	config := fmt.Sprintf(`
collectStats: false
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
	assert.False(t, serverCollection.CollectStat)

	server := serverCollection.Servers[0]
	assert.Equal(t, 4573, server.Port)
	assert.Equal(t, "server_1", server.Name)
	assert.Equal(t, 1, len(server.Endpoints))

	endpoint := server.Endpoints[0]
	assert.Equal(t, "/response_from_file", endpoint.Url)

	get_response := endpoint.GET
	body, err := ioutil.ReadFile(filepath)
	assert.Equal(t, string(body), get_response.Body)
	assert.Equal(t, "application/json", get_response.ContentType)
	assert.Equal(t, http.StatusOK, get_response.StatusCode)

	post_response := endpoint.POST
	assert.Nil(t, post_response)

	patch_response := endpoint.PATCH
	assert.Nil(t, patch_response)

	put_response := endpoint.PUT
	assert.Nil(t, put_response)

	delete_response := endpoint.DELETE
	assert.Nil(t, delete_response)

}
