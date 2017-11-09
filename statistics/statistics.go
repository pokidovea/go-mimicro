package statistics

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"sync"
	"time"

	"github.com/gorilla/mux"
)

type Request struct {
	ServerName string `json:"server"`
	Url        string `json:"endpoint"`
	Method     string `json:"method"`
	StatusCode int    `json:"status_code"`
}

type serverStatistic struct {
	Url    string `json:"url"`
	Method string `json:"method"`
	Count  int    `json:"count"`
}

func (request Request) String() string {
	return fmt.Sprintf(
		"server: %s; url: %s; method: %s; response status: %d",
		request.ServerName,
		request.Url,
		request.Method,
		request.StatusCode,
	)
}

type Collector struct {
	Chan     chan Request
	requests map[Request]int
}

func (collector *Collector) Add(request Request) {
	if collector.requests == nil {
		collector.requests = make(map[Request]int)
	}
	collector.requests[request]++
}

func (collector Collector) Get(request Request) int {
	output := collector.requests[request]
	return output
}

func (collector *Collector) Run(wg *sync.WaitGroup) {
	signalChannel := make(chan os.Signal, 1)
	signal.Notify(signalChannel, os.Interrupt)

	defer close(signalChannel)
	defer signal.Stop(signalChannel)
	defer log.Printf("Statistics collector stopped")
	defer wg.Done()

	for {
		select {
		case request, ok := <-collector.Chan:
			if !ok {
				return
			}
			collector.Add(request)
		case <-signalChannel:
			return
		}
	}
}

func (collector *Collector) getRequestStatistics(request *Request) []serverStatistic {
	var statistics []serverStatistic

	for collectedRequest, count := range collector.requests {
		if request.ServerName != collectedRequest.ServerName {
			continue
		}
		if request.Url != "" && request.Url != collectedRequest.Url {
			continue
		}
		if request.Method != "" && request.Method != collectedRequest.Method {
			continue
		}

		statistics = append(statistics, serverStatistic{
			Url:    collectedRequest.Url,
			Method: collectedRequest.Method,
			Count:  count,
		})
	}
	return statistics
}

func (collector *Collector) ServerStatisticsAPIHandler(w http.ResponseWriter, req *http.Request) {
	vars := mux.Vars(req)
	serverName := vars["serverName"]

	request := Request{
		ServerName: serverName,
	}

	urls_q, ok := req.URL.Query()["url"]
	if ok && len(urls_q) > 0 {
		request.Url = urls_q[0]
	}
	methods_q, ok := req.URL.Query()["method"]
	if ok && len(methods_q) > 0 {
		request.Method = methods_q[0]
	}

	statistics := collector.getRequestStatistics(&request)
	if len(statistics) == 0 {
		statistics = make([]serverStatistic, 0)
	}
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

func (collector *Collector) startHttpServer(port int) *http.Server {
	router := mux.NewRouter()
	router.HandleFunc("/servers/{serverName}", collector.ServerStatisticsAPIHandler)

	srv := &http.Server{
		Addr:           ":" + strconv.Itoa(port),
		Handler:        router,
		ReadTimeout:    10 * time.Second,
		WriteTimeout:   10 * time.Second,
		MaxHeaderBytes: 1 << 20,
	}

	go func() {
		if err := srv.ListenAndServe(); err != http.ErrServerClosed {
			// cannot panic, because this probably is an intentional close
			log.Printf("Httpserver: ListenAndServe() error: %s", err)
		}
	}()
	return srv
}

func (collector *Collector) Serve(port int, wg *sync.WaitGroup) {
	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt)
	defer close(interrupt)
	defer signal.Stop(interrupt)

	srv := collector.startHttpServer(port)
	<-interrupt

	log.Printf("[Statistics] Stopping a server...")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		log.Printf("[Statistics] Shutdown error: %s", err)
	}

	log.Printf("[Statistics] Server stopped")

	wg.Done()
}
