package mimicro

import (
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewServerWithoutStatistics(t *testing.T) {
	server := NewManagementServer(4534, false)

	assert.Equal(t, server.Port, 4534)
	assert.Nil(t, server.statisticsStorage)
}

func TestNewServerWithStatistics(t *testing.T) {
	server := NewManagementServer(4534, true)

	assert.Equal(t, server.Port, 4534)
	assert.NotNil(t, server.statisticsStorage)
}

func TestWriteRequestLogWithoutStatistics(t *testing.T) {
	server := NewManagementServer(4534, false)

	server.WriteRequestLog("server_1", "/some/url", "GET", http.StatusOK)
}

func TestWriteRequestLogWithStatistics(t *testing.T) {
	server := NewManagementServer(4534, true)

	// make the channel buffered to test in one thread
	server.statisticsStorage.RequestsChannel = make(chan ReceivedRequest, 1)

	server.WriteRequestLog("server_1", "/some/url", "GET", http.StatusOK)

	var request ReceivedRequest
	request = <-server.statisticsStorage.RequestsChannel

	expectedRequest := ReceivedRequest{
		ServerName: "server_1",
		URL:        "/some/url",
		Method:     "GET",
		StatusCode: http.StatusOK,
	}

	assert.Equal(t, expectedRequest, request)
}
