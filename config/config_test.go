package config

import (
	"github.com/stretchr/testify/assert"
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
        body: 'OK'
        content_type: text/plain
    `

	serverCollection, err := parseConfig([]byte(config))
	assert.Nil(t, err)
	assert.Equal(t, len(serverCollection.Servers), 1)
}
