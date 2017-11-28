package mockServer

import (
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/ghodss/yaml"
	"github.com/stretchr/testify/assert"
)

type responseLogMessage struct {
	ServerName, URL, Method string
	StatusCode              int
}

func (msg *responseLogMessage) writeResponseLog(serverName, URL, method string, statusCode int) {
	msg.ServerName = serverName
	msg.URL = URL
	msg.Method = method
	msg.StatusCode = statusCode
}

func TestHandleResponse(t *testing.T) {
	str := `
url: /simple_url
GET:
    template: GET
    headers:
        content-type: application/json
    status_code: 201
POST:
    template: POST
    headers:
        content-type: application/json
    status_code: 201
PUT:
    template: PUT
    headers:
        content-type: application/json
    status_code: 201
PATCH:
    template: PATCH
    headers:
        content-type: application/json
    status_code: 201
DELETE:
    template: DELETE
    headers:
        content-type: application/json
    status_code: 201
`

	var endpoint Endpoint
	yaml.Unmarshal([]byte(str), &endpoint)

	logMessage := new(responseLogMessage)

	methods := [...]string{"GET", "POST", "PATCH", "PUT", "DELETE"}
	handler := endpoint.GetHandler(logMessage.writeResponseLog, "server_name")

	for _, method := range methods {
		w := httptest.NewRecorder()
		r := httptest.NewRequest(method, "/simple_url", nil)

		handler(w, r)

		resp := w.Result()

		body, _ := ioutil.ReadAll(resp.Body)
		assert.Equal(t, "application/json", resp.Header.Get("Content-Type"))
		assert.Equal(t, http.StatusCreated, resp.StatusCode)
		assert.Equal(t, method, string(body))

		assert.Equal(t, "server_name", logMessage.ServerName)
		assert.Equal(t, "/simple_url", logMessage.URL)
		assert.Equal(t, method, logMessage.Method)
		assert.Equal(t, http.StatusCreated, logMessage.StatusCode)
	}
}

func TestHandleNonexistingResponses(t *testing.T) {
	str := `
url: /simple_url
GET:
    template: GET
    headers:
        content-type: application/json
    status_code: 201
`

	var endpoint Endpoint
	yaml.Unmarshal([]byte(str), &endpoint)

	logMessage := new(responseLogMessage)

	methods := [...]string{"POST", "PATCH", "PUT", "DELETE"}
	handler := endpoint.GetHandler(logMessage.writeResponseLog, "server_name")

	for _, method := range methods {
		w := httptest.NewRecorder()
		r := httptest.NewRequest(method, "/simple_url", nil)
		handler(w, r)

		resp := w.Result()

		body, _ := ioutil.ReadAll(resp.Body)
		assert.Equal(t, "text/plain; charset=utf-8", resp.Header.Get("Content-Type"))
		assert.Equal(t, http.StatusNotFound, resp.StatusCode)
		assert.Equal(t, "404 page not found\n", string(body))

		assert.Equal(t, "server_name", logMessage.ServerName)
		assert.Equal(t, "/simple_url", logMessage.URL)
		assert.Equal(t, method, logMessage.Method)
		assert.Equal(t, 404, logMessage.StatusCode)
	}
}
