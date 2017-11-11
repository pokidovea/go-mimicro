package response

import (
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"path"
	"runtime"
	"testing"
	"text/template"

	"github.com/gorilla/mux"
	"github.com/stretchr/testify/assert"
)

func TestWriteTemplateResponse(t *testing.T) {
	tmpl := template.New("template")
	tmpl.Parse(`{"passed_value": "{{.var}}"}`)

	response := Response{tmpl, "", "application/json", http.StatusCreated}
	router := mux.NewRouter()
	router.HandleFunc("/simple_url/{var}", response.WriteResponse)

	w := httptest.NewRecorder()
	req := httptest.NewRequest("POST", "/simple_url/1", nil)

	router.ServeHTTP(w, req)

	resp := w.Result()
	body, _ := ioutil.ReadAll(resp.Body)

	assert.Equal(t, "application/json", resp.Header.Get("Content-Type"))
	assert.Equal(t, http.StatusCreated, resp.StatusCode)
	assert.Equal(t, `{"passed_value": "1"}`, string(body))
}

func TestWriteFileResponse(t *testing.T) {
	_, filename, _, _ := runtime.Caller(0)

	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/simple_url", nil)

	filepath := path.Join(path.Dir(filename), "..", "..", "examples", "mimicro.png")
	var response = Response{nil, filepath, "", http.StatusOK}

	response.WriteResponse(w, req)

	resp := w.Result()
	body, _ := ioutil.ReadAll(resp.Body)

	assert.Equal(t, "image/png", resp.Header.Get("Content-Type"))
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	fileContent, err := ioutil.ReadFile(filepath)
	assert.Nil(t, err)
	assert.Equal(t, fileContent, body)
}

// func TestUnmarshallingOnlyBody(t *testing.T) {
// 	str := `{"body":"OK"}`

// 	var response Response
// 	err := json.Unmarshal([]byte(str), &response)
// 	assert.Nil(t, err)

// 	assert.Equal(t, "OK", response.Body)
// 	assert.Equal(t, "text/plain", response.ContentType)
// 	assert.Equal(t, http.StatusOK, response.StatusCode)
// }

// func TestUnmarshallingAllFields(t *testing.T) {
// 	str := `{"body":"{}", "content_type":"application/json", "status_code": 201}`

// 	var response Response
// 	err := json.Unmarshal([]byte(str), &response)
// 	assert.Nil(t, err)

// 	assert.Equal(t, "{}", response.Body)
// 	assert.Equal(t, "application/json", response.ContentType)
// 	assert.Equal(t, http.StatusCreated, response.StatusCode)
// }

// func TestWriteSimpleResponse(t *testing.T) {
// 	w := httptest.NewRecorder()
// 	r := httptest.NewRequest("POST", "/simple_url", nil)
// 	var response = Response{"{\"a\":1}", "application/json", http.StatusCreated, false}

// 	response.WriteResponse(w, r)

// 	resp := w.Result()
// 	body, _ := ioutil.ReadAll(resp.Body)

// 	assert.Equal(t, "application/json", resp.Header.Get("Content-Type"))
// 	assert.Equal(t, http.StatusCreated, resp.StatusCode)
// 	assert.Equal(t, "{\"a\":1}", string(body))
// }
