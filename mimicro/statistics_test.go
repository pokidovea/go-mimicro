package mimicro

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gorilla/mux"
	"github.com/stretchr/testify/assert"
)

func TestAddAndGetRequests(t *testing.T) {
	storage := newStatisticsStorage()
	request1 := ReceivedRequest{
		ServerName: "Simple server",
		URL:        "/some_url",
		Method:     "POST",
	}
	storage.add(request1)
	storage.add(request1)

	request2 := ReceivedRequest{
		ServerName: "Simple server",
		URL:        "/some_url",
		Method:     "GET",
	}
	storage.add(request2)

	assert.Equal(t, 2, storage.get(request1))
	assert.Equal(t, 1, storage.get(request2))
}

func TestDeleteRequests(t *testing.T) {
	storage := newStatisticsStorage()
	request1 := ReceivedRequest{
		ServerName: "Simple server",
		URL:        "/some_url",
		Method:     "POST",
		StatusCode: 0,
	}
	request2 := ReceivedRequest{
		ServerName: "Simple server",
		URL:        "/another_url",
		Method:     "GET",
		StatusCode: 0,
	}
	request3 := ReceivedRequest{
		ServerName: "Another server",
		URL:        "/another_url",
		Method:     "GET",
		StatusCode: 0,
	}

	storage.add(request1)
	storage.add(request2)
	storage.add(request3)

	pattern := requestPattern{
		ServerName: "Simple server",
		URL:        "*",
		Method:     "*",
	}

	storage.del(pattern)

	assert.Equal(t, 1, len(storage.requests))
}

func TestCollectFromChannel(t *testing.T) {
	storage := newStatisticsStorage()

	request := ReceivedRequest{
		ServerName: "Simple server",
		URL:        "/some_url",
		Method:     "POST",
	}

	storage.Start()

	storage.RequestsChannel <- request
	storage.RequestsChannel <- request
	storage.Stop()

	time.Sleep(10 * time.Millisecond)

	assert.Equal(t, 2, storage.get(request))
}

func TestStringifyRequest(t *testing.T) {
	request := ReceivedRequest{
		ServerName: "Simple server",
		URL:        "/some_url",
		Method:     "POST",
		StatusCode: http.StatusCreated,
	}

	assert.Equal(
		t,
		"server: Simple server; url: /some_url; method: POST; response status: 201",
		fmt.Sprintf("%s", request),
	)
}

func TestGetStatisticsHandlerWhenNothingPassed(t *testing.T) {
	router := mux.NewRouter()
	storage := newStatisticsStorage()
	request1 := ReceivedRequest{
		ServerName: "server_1",
		URL:        "/some_url",
		Method:     "POST",
		StatusCode: 0,
	}
	request2 := ReceivedRequest{
		ServerName: "server_2",
		URL:        "/another_url",
		Method:     "GET",
		StatusCode: 0,
	}

	storage.add(request1)
	storage.add(request1)
	storage.add(request2)

	router.HandleFunc("/url", storage.GetStatisticsHandler)

	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/url", nil)

	router.ServeHTTP(w, req)

	resp := w.Result()
	body, _ := ioutil.ReadAll(resp.Body)

	assert.Equal(t, "application/json", resp.Header.Get("Content-Type"))
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	expectedValues := []string{
		`[{"server":"server_1","url":"/some_url","method":"POST","count":2},` +
			`{"server":"server_2","url":"/another_url","method":"GET","count":1}]`,
		`[{"server":"server_2","url":"/another_url","method":"GET","count":1},` +
			`{"server":"server_1","url":"/some_url","method":"POST","count":2}]`,
	}

	assert.Contains(t, expectedValues, string(body))
}

func TestGetStatisticsHandlerWhenServerNamePassed(t *testing.T) {
	router := mux.NewRouter()
	storage := newStatisticsStorage()
	request1 := ReceivedRequest{
		ServerName: "server_1",
		URL:        "/some_url",
		Method:     "POST",
		StatusCode: 0,
	}
	request2 := ReceivedRequest{
		ServerName: "server_1",
		URL:        "/another_url",
		Method:     "GET",
		StatusCode: 0,
	}
	request3 := ReceivedRequest{
		ServerName: "server_2",
		URL:        "/another_url",
		Method:     "POST",
		StatusCode: 0,
	}

	storage.add(request1)
	storage.add(request1)
	storage.add(request2)
	storage.add(request3)

	router.HandleFunc("/url", storage.GetStatisticsHandler)

	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/url?server=server_1", nil)

	router.ServeHTTP(w, req)

	resp := w.Result()
	body, _ := ioutil.ReadAll(resp.Body)

	assert.Equal(t, "application/json", resp.Header.Get("Content-Type"))
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	expectedValues := []string{
		`[{"server":"server_1","url":"/some_url","method":"POST","count":2},` +
			`{"server":"server_1","url":"/another_url","method":"GET","count":1}]`,
		`[{"server":"server_1","url":"/another_url","method":"GET","count":1},` +
			`{"server":"server_1","url":"/some_url","method":"POST","count":2}]`,
	}

	assert.Contains(t, expectedValues, string(body))
}

func TestGetStatisticsHandlerWhenURLPassed(t *testing.T) {
	router := mux.NewRouter()
	storage := newStatisticsStorage()
	request1 := ReceivedRequest{
		ServerName: "server_1",
		URL:        "/some_url",
		Method:     "POST",
		StatusCode: 0,
	}
	request2 := ReceivedRequest{
		ServerName: "server_1",
		URL:        "/another_url",
		Method:     "GET",
		StatusCode: 0,
	}
	request3 := ReceivedRequest{
		ServerName: "server_2",
		URL:        "/another_url",
		Method:     "POST",
		StatusCode: 0,
	}

	storage.add(request1)
	storage.add(request2)
	storage.add(request2)
	storage.add(request3)

	router.HandleFunc("/url", storage.GetStatisticsHandler)

	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/url?url=/another_url", nil)

	router.ServeHTTP(w, req)

	resp := w.Result()
	body, _ := ioutil.ReadAll(resp.Body)

	assert.Equal(t, "application/json", resp.Header.Get("Content-Type"))
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	expectedValues := []string{
		`[{"server":"server_1","url":"/another_url","method":"GET","count":2},` +
			`{"server":"server_2","url":"/another_url","method":"POST","count":1}]`,
		`[{"server":"server_2","url":"/another_url","method":"POST","count":1},` +
			`{"server":"server_1","url":"/another_url","method":"GET","count":2}]`,
	}

	assert.Contains(t, expectedValues, string(body))
}

func TestGetStatisticsHandlerWhenMethodPassed(t *testing.T) {
	router := mux.NewRouter()
	storage := newStatisticsStorage()
	request1 := ReceivedRequest{
		ServerName: "server_1",
		URL:        "/some_url",
		Method:     "POST",
		StatusCode: 0,
	}
	request2 := ReceivedRequest{
		ServerName: "server_1",
		URL:        "/another_url",
		Method:     "GET",
		StatusCode: 0,
	}
	request3 := ReceivedRequest{
		ServerName: "server_2",
		URL:        "/another_url",
		Method:     "POST",
		StatusCode: 0,
	}

	storage.add(request1)
	storage.add(request2)
	storage.add(request3)
	storage.add(request3)

	router.HandleFunc("/url", storage.GetStatisticsHandler)

	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/url?method=post", nil)

	router.ServeHTTP(w, req)

	resp := w.Result()
	body, _ := ioutil.ReadAll(resp.Body)

	assert.Equal(t, "application/json", resp.Header.Get("Content-Type"))
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	expectedValues := []string{
		`[{"server":"server_1","url":"/some_url","method":"POST","count":1},` +
			`{"server":"server_2","url":"/another_url","method":"POST","count":2}]`,
		`[{"server":"server_2","url":"/another_url","method":"POST","count":2},` +
			`{"server":"server_1","url":"/some_url","method":"POST","count":1}]`,
	}

	assert.Contains(t, expectedValues, string(body))
}

func TestGetStatisticsHandlerWhenPassedAllParams(t *testing.T) {
	router := mux.NewRouter()
	storage := newStatisticsStorage()
	request1 := ReceivedRequest{
		ServerName: "server_1",
		URL:        "/some_url",
		Method:     "POST",
		StatusCode: 0,
	}
	request2 := ReceivedRequest{
		ServerName: "server_1",
		URL:        "/another_url",
		Method:     "POST",
		StatusCode: 0,
	}
	request3 := ReceivedRequest{
		ServerName: "server_2",
		URL:        "/another_url",
		Method:     "POST",
		StatusCode: 0,
	}
	request4 := ReceivedRequest{
		ServerName: "server_2",
		URL:        "/another_url",
		Method:     "GET",
		StatusCode: 0,
	}

	storage.add(request1)
	storage.add(request2)
	storage.add(request3)
	storage.add(request4)

	router.HandleFunc("/url", storage.GetStatisticsHandler)

	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/url?server=server_2&method=post&url=/another_url", nil)

	router.ServeHTTP(w, req)

	resp := w.Result()
	body, _ := ioutil.ReadAll(resp.Body)

	assert.Equal(t, "application/json", resp.Header.Get("Content-Type"))
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	assert.Equal(t, `[{"server":"server_2","url":"/another_url","method":"POST","count":1}]`, string(body))
}

func TestDeleteStatisticsHandlerWhenNothingPassed(t *testing.T) {
	router := mux.NewRouter()
	storage := newStatisticsStorage()
	request1 := ReceivedRequest{
		ServerName: "server_1",
		URL:        "/some_url",
		Method:     "POST",
		StatusCode: 0,
	}
	request2 := ReceivedRequest{
		ServerName: "server_2",
		URL:        "/another_url",
		Method:     "GET",
		StatusCode: 0,
	}

	storage.add(request1)
	storage.add(request1)
	storage.add(request2)

	router.HandleFunc("/url", storage.DeleteStatisticsHandler)

	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/url", nil)

	router.ServeHTTP(w, req)

	resp := w.Result()
	body, _ := ioutil.ReadAll(resp.Body)

	assert.Equal(t, "text/plain", resp.Header.Get("Content-Type"))
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	assert.Contains(t, "OK", string(body))

	assert.Len(t, storage.requests, 0)
}

func TestDeleteStatisticsHandlerWhenServerNamePassed(t *testing.T) {
	router := mux.NewRouter()
	storage := newStatisticsStorage()
	request1 := ReceivedRequest{
		ServerName: "server_1",
		URL:        "/some_url",
		Method:     "POST",
		StatusCode: 0,
	}
	request2 := ReceivedRequest{
		ServerName: "server_2",
		URL:        "/another_url",
		Method:     "GET",
		StatusCode: 0,
	}

	storage.add(request1)
	storage.add(request1)
	storage.add(request2)

	router.HandleFunc("/url", storage.DeleteStatisticsHandler)

	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/url?server=server_2", nil)

	router.ServeHTTP(w, req)

	resp := w.Result()
	body, _ := ioutil.ReadAll(resp.Body)

	assert.Equal(t, "text/plain", resp.Header.Get("Content-Type"))
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	assert.Contains(t, "OK", string(body))

	assert.Equal(t, requestsCounter{request1: 2}, storage.requests)
}

func TestDeleteStatisticsHandlerWhenURLPassed(t *testing.T) {
	router := mux.NewRouter()
	storage := newStatisticsStorage()
	request1 := ReceivedRequest{
		ServerName: "server_1",
		URL:        "/some_url",
		Method:     "POST",
		StatusCode: 0,
	}
	request2 := ReceivedRequest{
		ServerName: "server_2",
		URL:        "/another_url",
		Method:     "GET",
		StatusCode: 0,
	}
	request3 := ReceivedRequest{
		ServerName: "server_3",
		URL:        "/another_url",
		Method:     "GET",
		StatusCode: 0,
	}
	storage.add(request1)
	storage.add(request2)
	storage.add(request3)

	router.HandleFunc("/url", storage.DeleteStatisticsHandler)

	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/url?url=/another_url", nil)

	router.ServeHTTP(w, req)

	resp := w.Result()
	body, _ := ioutil.ReadAll(resp.Body)

	assert.Equal(t, "text/plain", resp.Header.Get("Content-Type"))
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	assert.Contains(t, "OK", string(body))

	assert.Equal(t, requestsCounter{request1: 1}, storage.requests)
}

func TestDeleteStatisticsHandlerWhenMethodPassed(t *testing.T) {
	router := mux.NewRouter()
	storage := newStatisticsStorage()
	request1 := ReceivedRequest{
		ServerName: "server_1",
		URL:        "/some_url",
		Method:     "POST",
		StatusCode: 0,
	}
	request2 := ReceivedRequest{
		ServerName: "server_2",
		URL:        "/another_url",
		Method:     "GET",
		StatusCode: 0,
	}
	request3 := ReceivedRequest{
		ServerName: "server_3",
		URL:        "/another_url",
		Method:     "GET",
		StatusCode: 0,
	}
	storage.add(request1)
	storage.add(request2)
	storage.add(request3)

	router.HandleFunc("/url", storage.DeleteStatisticsHandler)

	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/url?method=post", nil)

	router.ServeHTTP(w, req)

	resp := w.Result()
	body, _ := ioutil.ReadAll(resp.Body)

	assert.Equal(t, "text/plain", resp.Header.Get("Content-Type"))
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	assert.Contains(t, "OK", string(body))

	assert.Equal(t, requestsCounter{request2: 1, request3: 1}, storage.requests)
}
