package management

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"strings"
	"sync"

	"github.com/gorilla/mux"
)

// ReceivedRequest represents a request that was sent to a mock server
type ReceivedRequest struct {
	ServerName string
	URL        string
	Method     string
	StatusCode int
}

func (request ReceivedRequest) String() string {
	return fmt.Sprintf(
		"server: %s; url: %s; method: %s; response status: %d",
		request.ServerName,
		request.URL,
		request.Method,
		request.StatusCode,
	)
}

type requestsCounter map[ReceivedRequest]int

func (counter requestsCounter) MarshalJSON() ([]byte, error) {
	buffer := bytes.NewBufferString("[")
	length := len(counter)
	count := 0
	for request, requestsCount := range counter {
		buffer.WriteString("{")
		buffer.WriteString(fmt.Sprintf("\"url\":\"%s\",", request.URL))
		buffer.WriteString(fmt.Sprintf("\"method\":\"%s\",", request.Method))
		buffer.WriteString(fmt.Sprintf("\"count\":%s", strconv.Itoa(requestsCount)))
		buffer.WriteString("}")
		count++
		if count < length {
			buffer.WriteString(",")
		}
	}
	buffer.WriteString("]")
	return buffer.Bytes(), nil
}

type statisticsStorage struct {
	mutex           sync.RWMutex
	RequestsChannel chan ReceivedRequest
	requests        requestsCounter
}

func newStatisticsStorage() *statisticsStorage {
	storage := new(statisticsStorage)
	storage.requests = make(requestsCounter)
	storage.RequestsChannel = make(chan ReceivedRequest)
	return storage
}

func (storage *statisticsStorage) add(request ReceivedRequest) {
	storage.mutex.Lock()
	defer storage.mutex.Unlock()
	storage.requests[request]++
}

func (storage *statisticsStorage) get(request ReceivedRequest) int {
	storage.mutex.RLock()
	defer storage.mutex.RUnlock()

	return storage.requests[request]
}

func (storage *statisticsStorage) iter(f func(request ReceivedRequest, count int) bool) {
	storage.mutex.RLock()
	defer storage.mutex.RUnlock()

	for request, count := range storage.requests {
		if !f(request, count) {
			return
		}
	}
}

func (storage *statisticsStorage) Run(done <-chan bool) {
	log.Printf("[Statistics storage] Starting...")

	defer log.Printf("[Statistics storage] Stopped")

	for {
		select {
		case request, ok := <-storage.RequestsChannel:
			if !ok {
				return
			}
			storage.add(request)
		case <-done:
			return
		}
	}
}

func (storage *statisticsStorage) getRequestStatistics(request *ReceivedRequest) requestsCounter {
	records := make(requestsCounter)

	storage.iter(func(collectedRequest ReceivedRequest, count int) bool {
		if request.ServerName != collectedRequest.ServerName {
			return true
		}
		if request.URL != "" && request.URL != collectedRequest.URL {
			return true
		}
		if request.Method != "" && request.Method != collectedRequest.Method {
			return true
		}

		records[collectedRequest] = count

		return true
	})
	return records
}

func (storage *statisticsStorage) HTTPHandler(w http.ResponseWriter, req *http.Request) {
	vars := mux.Vars(req)
	serverName := vars["serverName"]

	request := ReceivedRequest{
		ServerName: serverName,
	}

	urls, ok := req.URL.Query()["url"]
	if ok && len(urls) > 0 {
		request.URL = urls[0]
	}
	methods, ok := req.URL.Query()["method"]
	if ok && len(methods) > 0 {
		request.Method = strings.ToUpper(methods[0])
	}

	statistics := storage.getRequestStatistics(&request)
	payload, err := json.Marshal(statistics)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("Internal server error"))
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write(payload)
}
