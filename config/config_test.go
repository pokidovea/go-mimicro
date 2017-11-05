package config

import (
	"github.com/stretchr/testify/assert"
	"net/http"
	"testing"
)

func Test1(t *testing.T) {
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
	assert.Equal(t, server.Port, 4573)
	assert.Equal(t, len(server.Endpoints), 1)

	endpoint := server.Endpoints[0]
	assert.Equal(t, endpoint.Url, "/simple_url")

	get_response := endpoint.GET
	assert.Equal(t, get_response.Body, "{}")
	assert.Equal(t, get_response.ContentType, "application/json")
	assert.Equal(t, get_response.StatusCode, http.StatusOK)

	post_response := endpoint.POST
	assert.Equal(t, post_response.Body, "{}")
	assert.Equal(t, get_response.ContentType, "text/plain")
	assert.Equal(t, get_response.StatusCode, http.StatusCreated)

	patch_response := endpoint.PATCH
	assert.Nil(t, patch_response)

	put_response := endpoint.PUT
	assert.Nil(t, put_response)

	delete_response := endpoint.DELETE
	assert.Nil(t, delete_response)
}
