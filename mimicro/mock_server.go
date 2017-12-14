package mimicro

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"sync"
	"time"

	"github.com/gorilla/mux"
)

// RequestLogWriter is signature of method, which should be passed to the mock server to write requests log
type RequestLogWriter func(serverName, URL, method string, statusCode int)

// MockServer represents a standalone mock server with its name, port and collection of endpoints
type MockServer struct {
	Name      string     `json:"name"`
	Port      int        `json:"port"`
	Endpoints []Endpoint `json:"endpoints"`
}

func (mockServer MockServer) startHTTPServer(logWriter RequestLogWriter) *http.Server {
	router := mux.NewRouter()

	for _, endpoint := range mockServer.Endpoints {
		router.HandleFunc(endpoint.URL, endpoint.GetHandler(logWriter, mockServer.Name))
	}

	srv := &http.Server{
		Addr:           ":" + strconv.Itoa(mockServer.Port),
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

// Serve method starts the server and does some operations after it stops
func (mockServer MockServer) Serve(logWriter RequestLogWriter, wg *sync.WaitGroup) {
	log.Printf("[%s] Starting...", mockServer.Name)
	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt)
	defer close(interrupt)
	defer signal.Stop(interrupt)

	srv := mockServer.startHTTPServer(logWriter)
	<-interrupt

	log.Printf("[%s] Stopping...", mockServer.Name)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		log.Printf("[%s] Shutdown error: %s", mockServer.Name, err)
	}

	log.Printf("[%s] Stopped", mockServer.Name)

	wg.Done()
}
