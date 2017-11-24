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

func createEndpoint() Endpoint {
	str := `
url: /simple_url
GET:
    template: "{}"
    headers:
        content-type: application/json
POST:
    template: OK
    status_code: 201
`

	var endpoint Endpoint
	err := yaml.Unmarshal([]byte(str), &endpoint)

	if err != nil {
		panic(err)
	}

	return endpoint
}

func TestHandleGETResponse(t *testing.T) {
	endpoint := createEndpoint()
	logMessage := new(responseLogMessage)

	w := httptest.NewRecorder()
	r := httptest.NewRequest("GET", "/simple_url", nil)

	handler := endpoint.GetHandler(logMessage.writeResponseLog, "server_name")
	handler(w, r)

	resp := w.Result()

	body, _ := ioutil.ReadAll(resp.Body)
	assert.Equal(t, "application/json", resp.Header.Get("Content-Type"))
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	assert.Equal(t, "{}", string(body))

	assert.Equal(t, "server_name", logMessage.ServerName)
	assert.Equal(t, "/simple_url", logMessage.URL)
	assert.Equal(t, "GET", logMessage.Method)
	assert.Equal(t, 200, logMessage.StatusCode)
}

func TestHandlePOSTResponse(t *testing.T) {
	endpoint := createEndpoint()
	logMessage := new(responseLogMessage)

	w := httptest.NewRecorder()
	r := httptest.NewRequest("POST", "/simple_url", nil)

	handler := endpoint.GetHandler(logMessage.writeResponseLog, "server_name")
	handler(w, r)

	resp := w.Result()

	body, _ := ioutil.ReadAll(resp.Body)
	assert.Equal(t, http.StatusCreated, resp.StatusCode)
	assert.Equal(t, "OK", string(body))

	assert.Equal(t, "server_name", logMessage.ServerName)
	assert.Equal(t, "/simple_url", logMessage.URL)
	assert.Equal(t, "POST", logMessage.Method)
	assert.Equal(t, 201, logMessage.StatusCode)
}

func TestHandleNonexistingResponses(t *testing.T) {
	endpoint := createEndpoint()
	logMessage := new(responseLogMessage)

	handler := endpoint.GetHandler(logMessage.writeResponseLog, "server_name")

	methods := [...]string{"PATCH", "PUT", "DELETE"}

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
