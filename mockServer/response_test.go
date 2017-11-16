package mockServer

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"path"
	"runtime"
	"testing"
	"text/template"

	"github.com/ghodss/yaml"
	"github.com/gorilla/mux"
	"github.com/stretchr/testify/assert"
)

func TestWriteTemplateResponse(t *testing.T) {
	tmpl := template.New("template")
	tmpl.Parse(`{"passed_value": "{{.var}}"}`)

	response := Response{tmpl, nil, "application/json", http.StatusCreated}
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

func TestWriteTemplateResponseWithoutVarsInURL(t *testing.T) {
	tmpl := template.New("template")
	tmpl.Parse(`{"passed_value": "{{.var}}"}`)

	response := Response{tmpl, nil, "application/json", http.StatusCreated}
	router := mux.NewRouter()
	router.HandleFunc("/simple_url", response.WriteResponse)

	w := httptest.NewRecorder()
	req := httptest.NewRequest("POST", "/simple_url", nil)

	router.ServeHTTP(w, req)

	resp := w.Result()
	body, _ := ioutil.ReadAll(resp.Body)

	assert.Equal(t, "application/json", resp.Header.Get("Content-Type"))
	assert.Equal(t, http.StatusCreated, resp.StatusCode)
	assert.Equal(t, `{"passed_value": "<no value>"}`, string(body))
}

func TestWriteTemplateResponseWithoutVarsInTemplate(t *testing.T) {
	tmpl := template.New("template")
	tmpl.Parse(`{"passed_value": "2"}`)

	response := Response{tmpl, nil, "application/json", http.StatusCreated}
	router := mux.NewRouter()
	router.HandleFunc("/simple_url/{var}", response.WriteResponse)

	w := httptest.NewRecorder()
	req := httptest.NewRequest("POST", "/simple_url/1", nil)

	router.ServeHTTP(w, req)

	resp := w.Result()
	body, _ := ioutil.ReadAll(resp.Body)

	assert.Equal(t, "application/json", resp.Header.Get("Content-Type"))
	assert.Equal(t, http.StatusCreated, resp.StatusCode)
	assert.Equal(t, `{"passed_value": "2"}`, string(body))
}

func TestWriteFileResponse(t *testing.T) {
	_, filename, _, _ := runtime.Caller(0)

	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/simple_url", nil)

	filepath := path.Join(path.Dir(filename), "..", "examples", "mimicro.png")

	tmpl := template.New("template")
	tmpl.Parse(filepath)

	var response = Response{nil, tmpl, "", http.StatusOK}

	response.WriteResponse(w, req)

	resp := w.Result()
	body, _ := ioutil.ReadAll(resp.Body)

	assert.Equal(t, "image/png", resp.Header.Get("Content-Type"))
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	fileContent, err := ioutil.ReadFile(filepath)
	assert.Nil(t, err)
	assert.Equal(t, fileContent, body)
}

func TestWriteFileResponseWithVarInPath(t *testing.T) {
	_, filename, _, _ := runtime.Caller(0)
	filepath := path.Join(path.Dir(filename), "..", "examples", "{{.var}}micro.png")
	realFilepath := path.Join(path.Dir(filename), "..", "examples", "mimicro.png")

	tmpl := template.New("template")
	tmpl.Parse(filepath)

	response := Response{nil, tmpl, "", http.StatusOK}
	router := mux.NewRouter()
	router.HandleFunc("/{var}/in/filepath", response.WriteResponse)

	w := httptest.NewRecorder()
	req := httptest.NewRequest("POST", "/mi/in/filepath", nil)

	router.ServeHTTP(w, req)

	resp := w.Result()
	body, _ := ioutil.ReadAll(resp.Body)

	assert.Equal(t, "image/png", resp.Header.Get("Content-Type"))
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	fileContent, err := ioutil.ReadFile(realFilepath)
	assert.Nil(t, err)
	assert.Equal(t, fileContent, body)
}

func createResponseFromConfig(config string) Response {
	var response Response
	err := yaml.Unmarshal([]byte(config), &response)

	if err != nil {
		panic(err)
	}

	return response
}

func executeTemplate(tmpl *template.Template, vars map[string]string) string {
	w := bytes.NewBufferString("")
	err := tmpl.Execute(w, vars)

	if err != nil {
		panic(err)
	}

	return w.String()
}

func TestUnmarshalTemplateString(t *testing.T) {
	config := `
template: "var = {{.var}}"
content_type: application/json
status_code: 201
    `

	response := createResponseFromConfig(config)

	assert.Nil(t, response.file)
	assert.NotNil(t, response.template)
	assert.Equal(t, "var = 42", executeTemplate(response.template, map[string]string{"var": "42"}))

	assert.Equal(t, "application/json", response.ContentType)
	assert.Equal(t, http.StatusCreated, response.StatusCode)
}

func TestUnmarshalTemplateFile(t *testing.T) {
	_, filename, _, _ := runtime.Caller(0)
	configPath = path.Join(path.Dir(filename), "..", "examples", "config.yaml")

	cases := []string{
		path.Join(path.Dir(filename), "..", "examples", "response_with_var.json"), // absolute path
		"./response_with_var.json",
		"../examples/response_with_var.json",
		"../../mimicro/examples/response_with_var.json",
		"response_with_var.json",
	}

	for _, filePath := range cases {
		config := fmt.Sprintf(`template: file://%s`, filePath)

		response := createResponseFromConfig(config)

		assert.Nil(t, response.file)
		assert.NotNil(t, response.template)
		assert.Equal(
			t,
			"{\n    \"passed_value\": \"43\"\n}\n",
			executeTemplate(response.template, map[string]string{"var": "43"}),
		)

		assert.Equal(t, "text/plain", response.ContentType)
		assert.Equal(t, http.StatusOK, response.StatusCode)
	}
}

func TestUnmarshalBinaryFile(t *testing.T) {
	_, filename, _, _ := runtime.Caller(0)
	configPath = path.Join(path.Dir(filename), "..", "examples", "config.yaml")
	absoluteFilePath := path.Join(path.Dir(filename), "..", "examples", "mimicro.png")

	cases := []string{
		absoluteFilePath,
		"./mimicro.png",
		"../examples/mimicro.png",
		"../../mimicro/examples/mimicro.png",
		"mimicro.png",
		"{{.var}}micro.png",
	}

	for _, filePath := range cases {
		config := fmt.Sprintf(`file: file://%s`, filePath)

		response := createResponseFromConfig(config)

		assert.Nil(t, response.template)
		assert.Equal(
			t,
			absoluteFilePath,
			executeTemplate(response.file, map[string]string{"var": "mi"}),
		)

		assert.Equal(t, "", response.ContentType)
		assert.Equal(t, http.StatusOK, response.StatusCode)
	}
}

func TestUnmarshalNonexistingTemplate(t *testing.T) {
	var response Response

	config := "template: file:///wrong_file"

	err := yaml.Unmarshal([]byte(config), &response)
	assert.NotNil(t, err)
	assert.Equal(t, "error unmarshaling JSON: File does not exist /wrong_file", err.Error())

}
