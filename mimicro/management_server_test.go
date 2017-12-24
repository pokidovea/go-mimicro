package mimicro

import (
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewServerWithoutStatistics(t *testing.T) {
	substitutionStorage := NewSubstitutionStorage()
	server := NewManagementServer(4534, false, substitutionStorage)

	assert.Equal(t, 4534, server.Port)
	assert.Nil(t, server.statisticsStorage)
	assert.Equal(t, substitutionStorage, server.substitutionStorage)
}

func TestNewServerWithStatistics(t *testing.T) {
	substitutionStorage := NewSubstitutionStorage()
	server := NewManagementServer(4534, true, substitutionStorage)

	assert.Equal(t, 4534, server.Port)
	assert.NotNil(t, server.statisticsStorage)
}

func TestWriteRequestLogWithoutStatistics(t *testing.T) {
	substitutionStorage := NewSubstitutionStorage()
	server := NewManagementServer(4534, false, substitutionStorage)

	server.WriteRequestLog("server_1", "/some/url", "GET", http.StatusOK)
}

func TestWriteRequestLogWithStatistics(t *testing.T) {
	substitutionStorage := NewSubstitutionStorage()
	server := NewManagementServer(4534, true, substitutionStorage)

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
