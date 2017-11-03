package response

import (
	"github.com/stretchr/testify/assert"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestResponseHasOnlyBody(t *testing.T) {
	w := httptest.NewRecorder()
	var response = Response{"OK", "", 0}

	response.WriteResponse(w)

	resp := w.Result()
	body, _ := ioutil.ReadAll(resp.Body)

	assert.Equal(t, resp.Header.Get("Content-Type"), "text/plain")
	assert.Equal(t, resp.StatusCode, http.StatusOK)
	assert.Equal(t, string(body), "OK")
}

func TestAllFieldsAreDefined(t *testing.T) {
	w := httptest.NewRecorder()
	var response = Response{"{\"a\":1}", "application/json", http.StatusCreated}

	response.WriteResponse(w)

	resp := w.Result()
	body, _ := ioutil.ReadAll(resp.Body)

	assert.Equal(t, resp.Header.Get("Content-Type"), "application/json")
	assert.Equal(t, resp.StatusCode, http.StatusCreated)
	assert.Equal(t, string(body), "{\"a\":1}")
}
