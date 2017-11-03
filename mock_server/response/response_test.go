package response

import (
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

	if resp.Header.Get("Content-Type") != "text/plain" {
		t.Errorf("Expected content type to be 'text/plain', but was '%v'", resp.Header.Get("Content-Type"))
	}

	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected code to be %v, but was %v", http.StatusOK, resp.StatusCode)
	}

	if string(body) != "OK" {
		t.Errorf("Expected body to be %v, but was %v", "OK", string(body))
	}
}

func TestAllFieldsAreDefined(t *testing.T) {
	w := httptest.NewRecorder()
	var response = Response{"{\"a\":1}", "application/json", http.StatusCreated}

	response.WriteResponse(w)

	resp := w.Result()
	body, _ := ioutil.ReadAll(resp.Body)

	if resp.Header.Get("Content-Type") != "application/json" {
		t.Errorf("Expected content type to be 'application/json', but was '%v'", resp.Header.Get("Content-Type"))
	}

	if resp.StatusCode != http.StatusCreated {
		t.Errorf("Expected code to be %v, but was %v", http.StatusCreated, resp.StatusCode)
	}

	if string(body) != "{\"a\":1}" {
		t.Errorf("Expected body to be %v, but was %v", "{\"a\":1}", string(body))
	}
}
