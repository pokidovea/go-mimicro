package response

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"path"
	"runtime"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestUnmarshallingOnlyBody(t *testing.T) {
	str := `{"body":"OK"}`

	var response Response
	err := json.Unmarshal([]byte(str), &response)
	assert.Nil(t, err)

	assert.Equal(t, "OK", response.Body)
	assert.Equal(t, "text/plain", response.ContentType)
	assert.Equal(t, http.StatusOK, response.StatusCode)
}

func TestUnmarshallingAllFields(t *testing.T) {
	str := `{"body":"{}", "content_type":"application/json", "status_code": 201}`

	var response Response
	err := json.Unmarshal([]byte(str), &response)
	assert.Nil(t, err)

	assert.Equal(t, "{}", response.Body)
	assert.Equal(t, "application/json", response.ContentType)
	assert.Equal(t, http.StatusCreated, response.StatusCode)
}

func TestWriteSimpleResponse(t *testing.T) {
	w := httptest.NewRecorder()
	r := httptest.NewRequest("POST", "/simple_url", nil)
	var response = Response{"{\"a\":1}", "application/json", http.StatusCreated, false}

	response.WriteResponse(w, r)

	resp := w.Result()
	body, _ := ioutil.ReadAll(resp.Body)

	assert.Equal(t, "application/json", resp.Header.Get("Content-Type"))
	assert.Equal(t, http.StatusCreated, resp.StatusCode)
	assert.Equal(t, "{\"a\":1}", string(body))
}

func TestWriteFileResponse(t *testing.T) {
	_, filename, _, _ := runtime.Caller(0)

	w := httptest.NewRecorder()
	r := httptest.NewRequest("GET", "/simple_url", nil)

	filepath := path.Join(path.Dir(filename), "..", "..", "config", "fixtures", "mimicro.png")
	var response = Response{filepath, "", http.StatusOK, true}

	response.WriteResponse(w, r)

	resp := w.Result()
	body, _ := ioutil.ReadAll(resp.Body)

	assert.Equal(t, "image/png", resp.Header.Get("Content-Type"))
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	fileContent, err := ioutil.ReadFile(filepath)
	assert.Nil(t, err)
	assert.Equal(t, fileContent, body)
}
