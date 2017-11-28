package management

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

func TestCollectFromChannel(t *testing.T) {
	storage := newStatisticsStorage()

	done := make(chan bool, 1)
	defer close(done)

	request := ReceivedRequest{
		ServerName: "Simple server",
		URL:        "/some_url",
		Method:     "POST",
	}

	go storage.Run(done)

	storage.RequestsChannel <- request
	storage.RequestsChannel <- request
	done <- true

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

func TestGetRequestStatisticsWhenNothingFound(t *testing.T) {
	request := ReceivedRequest{
		ServerName: "Simple server",
		URL:        "/some_url",
		Method:     "POST",
		StatusCode: 0,
	}

	storage := newStatisticsStorage()

	result := storage.getRequestStatistics(&request)

	assert.Empty(t, result)
}

func TestGetRequestStatisticsByServerName(t *testing.T) {
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

	storage := newStatisticsStorage()
	storage.add(request1)
	storage.add(request1)
	storage.add(request2)
	storage.add(request3)

	request := ReceivedRequest{
		ServerName: "Simple server",
		URL:        "",
		Method:     "",
		StatusCode: 0,
	}

	result := storage.getRequestStatistics(&request)

	expectedResult := requestsCounter{
		request1: 2,
		request2: 1,
	}

	assert.Equal(t, expectedResult, result)

}

func TestGetRequestStatisticsByURL(t *testing.T) {
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

	storage := newStatisticsStorage()
	storage.add(request1)
	storage.add(request1)
	storage.add(request2)

	request := ReceivedRequest{
		ServerName: "Simple server",
		URL:        "/some_url",
		Method:     "",
		StatusCode: 0,
	}

	result := storage.getRequestStatistics(&request)

	expectedResult := requestsCounter{
		request1: 2,
	}

	assert.Equal(t, expectedResult, result)
}

func TestGetRequestStatisticsByMethod(t *testing.T) {
	request1 := ReceivedRequest{
		ServerName: "Simple server",
		URL:        "/some_url",
		Method:     "POST",
		StatusCode: 0,
	}
	request2 := ReceivedRequest{
		ServerName: "Simple server",
		URL:        "/another_url",
		Method:     "POST",
		StatusCode: 0,
	}
	request3 := ReceivedRequest{
		ServerName: "Simple server",
		URL:        "/another_url",
		Method:     "GET",
		StatusCode: 0,
	}

	storage := newStatisticsStorage()
	storage.add(request1)
	storage.add(request1)
	storage.add(request2)
	storage.add(request3)

	request := ReceivedRequest{
		ServerName: "Simple server",
		URL:        "",
		Method:     "POST",
		StatusCode: 0,
	}

	result := storage.getRequestStatistics(&request)

	expectedResult := requestsCounter{
		request1: 2,
		request2: 1,
	}

	assert.Equal(t, expectedResult, result)
}

func TestHTTPHandlerOnlyServerName(t *testing.T) {
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

	router.HandleFunc("/url/{serverName}", storage.HTTPHandler)

	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/url/server_1", nil)

	router.ServeHTTP(w, req)

	resp := w.Result()
	body, _ := ioutil.ReadAll(resp.Body)

	assert.Equal(t, "application/json", resp.Header.Get("Content-Type"))
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	expectedValues := []string{
		`[{"url":"/some_url","method":"POST","count":2},{"url":"/another_url","method":"GET","count":1}]`,
		`[{"url":"/another_url","method":"GET","count":1},{"url":"/some_url","method":"POST","count":2}]`,
	}

	assert.Contains(t, expectedValues, string(body))
}

func TestHTTPHandlerWhenURLPassed(t *testing.T) {
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
		ServerName: "server_1",
		URL:        "/another_url",
		Method:     "POST",
		StatusCode: 0,
	}

	storage.add(request1)
	storage.add(request2)
	storage.add(request2)
	storage.add(request3)

	router.HandleFunc("/url/{serverName}", storage.HTTPHandler)

	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/url/server_1?url=/another_url", nil)

	router.ServeHTTP(w, req)

	resp := w.Result()
	body, _ := ioutil.ReadAll(resp.Body)

	assert.Equal(t, "application/json", resp.Header.Get("Content-Type"))
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	expectedValues := []string{
		`[{"url":"/another_url","method":"POST","count":1},{"url":"/another_url","method":"GET","count":2}]`,
		`[{"url":"/another_url","method":"GET","count":2},{"url":"/another_url","method":"POST","count":1}]`,
	}

	assert.Contains(t, expectedValues, string(body))
}

func TestHTTPHandlerWhenMethodPassed(t *testing.T) {
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
		ServerName: "server_1",
		URL:        "/another_url",
		Method:     "POST",
		StatusCode: 0,
	}

	storage.add(request1)
	storage.add(request2)
	storage.add(request3)

	router.HandleFunc("/url/{serverName}", storage.HTTPHandler)

	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/url/server_1?method=post", nil)

	router.ServeHTTP(w, req)

	resp := w.Result()
	body, _ := ioutil.ReadAll(resp.Body)

	assert.Equal(t, "application/json", resp.Header.Get("Content-Type"))
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	expectedValues := []string{
		`[{"url":"/some_url","method":"POST","count":1},{"url":"/another_url","method":"POST","count":1}]`,
		`[{"url":"/another_url","method":"POST","count":1},{"url":"/some_url","method":"POST","count":1}]`,
	}

	assert.Contains(t, expectedValues, string(body))
}

func TestHTTPHandlerWhenPassedBothMethodAndURL(t *testing.T) {
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
		ServerName: "server_1",
		URL:        "/another_url",
		Method:     "POST",
		StatusCode: 0,
	}

	storage.add(request1)
	storage.add(request2)
	storage.add(request3)

	router.HandleFunc("/url/{serverName}", storage.HTTPHandler)

	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/url/server_1?method=post&url=/some_url", nil)

	router.ServeHTTP(w, req)

	resp := w.Result()
	body, _ := ioutil.ReadAll(resp.Body)

	assert.Equal(t, "application/json", resp.Header.Get("Content-Type"))
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	assert.Equal(t, `[{"url":"/some_url","method":"POST","count":1}]`, string(body))
}
