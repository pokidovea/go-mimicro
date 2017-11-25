package management

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
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

type statisticsRecord struct {
	URL    string `json:"url"`
	Method string `json:"method"`
	Count  int    `json:"count"`
}

type statisticsStorage struct {
	mutex           sync.RWMutex
	RequestsChannel chan ReceivedRequest
	requests        map[ReceivedRequest]int
}

func newStatisticsStorage() *statisticsStorage {
	storage := new(statisticsStorage)
	storage.requests = make(map[ReceivedRequest]int)
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

func (storage *statisticsStorage) getRequestStatistics(request *ReceivedRequest) []statisticsRecord {
	var records []statisticsRecord

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

		records = append(
			records,
			statisticsRecord{
				URL:    collectedRequest.URL,
				Method: collectedRequest.Method,
				Count:  count,
			},
		)

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
		request.Method = methods[0]
	}

	statistics := storage.getRequestStatistics(&request)
	payload, err := json.Marshal(statistics)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("intervalServerError"))
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write(payload)
}
