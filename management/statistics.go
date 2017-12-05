package management

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"sync"
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

type requestPattern struct {
	ServerName string
	URL        string
	Method     string
}

func (pattern requestPattern) matches(request ReceivedRequest) bool {
	if pattern.ServerName != "*" && pattern.ServerName != request.ServerName {
		return false
	}
	if pattern.URL != "*" && pattern.URL != request.URL {
		return false
	}
	if pattern.Method != "*" && pattern.Method != request.Method {
		return false
	}

	return true
}

func createRequestPatternFromQuery(URL *url.URL) requestPattern {
	var pattern requestPattern

	servers, ok := URL.Query()["server"]
	if ok && len(servers) > 0 {
		pattern.ServerName = servers[0]
	} else {
		pattern.ServerName = "*"
	}

	urls, ok := URL.Query()["url"]
	if ok && len(urls) > 0 {
		pattern.URL = urls[0]
	} else {
		pattern.URL = "*"
	}

	methods, ok := URL.Query()["method"]
	if ok && len(methods) > 0 {
		pattern.Method = strings.ToUpper(methods[0])
	} else {
		pattern.Method = "*"
	}

	return pattern
}

type requestsCounter map[ReceivedRequest]int

func (counter requestsCounter) MarshalJSON() ([]byte, error) {
	buffer := bytes.NewBufferString("[")
	length := len(counter)
	count := 0
	for request, requestsCount := range counter {
		buffer.WriteString("{")
		buffer.WriteString(fmt.Sprintf("\"server\":\"%s\",", request.ServerName))
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
	storage.RequestsChannel = make(chan ReceivedRequest, 100)
	return storage
}

func (storage *statisticsStorage) add(request ReceivedRequest) {
	storage.mutex.Lock()
	defer storage.mutex.Unlock()
	storage.requests[request]++
}

func (storage *statisticsStorage) del(pattern requestPattern) {
	storage.mutex.Lock()
	defer storage.mutex.Unlock()

	for request := range storage.requests {
		if pattern.matches(request) {
			delete(storage.requests, request)
		}
	}
}

func (storage *statisticsStorage) get(request ReceivedRequest) int {
	storage.mutex.RLock()
	defer storage.mutex.RUnlock()

	return storage.requests[request]
}

func (storage *statisticsStorage) iterate(f func(request ReceivedRequest, count int) bool) {
	storage.mutex.RLock()
	defer storage.mutex.RUnlock()

	for request, count := range storage.requests {
		if !f(request, count) {
			return
		}
	}
}

func (storage *statisticsStorage) filter(pattern requestPattern) requestsCounter {
	records := make(requestsCounter)

	storage.iterate(func(request ReceivedRequest, count int) bool {
		if !pattern.matches(request) {
			return true
		}

		records[request] = count

		return true
	})
	return records
}

func (storage *statisticsStorage) run() {
	log.Printf("[Statistics storage] Starting...")

	defer log.Printf("[Statistics storage] Stopped")

	for request := range storage.RequestsChannel {
		storage.add(request)
	}
}

func (storage *statisticsStorage) Start() {
	go storage.run()
}

func (storage *statisticsStorage) Stop() {
	close(storage.RequestsChannel)
}

func (storage *statisticsStorage) GetStatisticsHandler(w http.ResponseWriter, req *http.Request) {
	pattern := createRequestPatternFromQuery(req.URL)

	statistics := storage.filter(pattern)
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

func (storage *statisticsStorage) DeleteStatisticsHandler(w http.ResponseWriter, req *http.Request) {
	pattern := createRequestPatternFromQuery(req.URL)

	storage.del(pattern)
	w.Header().Set("Content-Type", "text/plain")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("OK"))
}
