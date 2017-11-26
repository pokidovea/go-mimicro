package management

import (
	"fmt"
	"net/http"
	"testing"
	"time"

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

	storage := newStatisticsStorage()
	storage.add(request1)
	storage.add(request1)
	storage.add(request2)

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
