package mockServer

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
	"github.com/pokidovea/mimicro/statistics"
)

type MockServer struct {
	Name      string     `json:"name"`
	Port      int        `json:"port"`
	Endpoints []Endpoint `json:"endpoints"`
}

func (mockServer MockServer) startHttpServer(statisticsChannel chan statistics.Request) *http.Server {
	router := mux.NewRouter()

	for _, endpoint := range mockServer.Endpoints {
		endpoint.CollectStatistics(statisticsChannel, mockServer.Name)
		router.HandleFunc(endpoint.Url, endpoint.GetHandler())
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

func (mockServer MockServer) Serve(statisticsChannel chan statistics.Request, wg *sync.WaitGroup) {
	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt)
	defer close(interrupt)
	defer signal.Stop(interrupt)

	srv := mockServer.startHttpServer(statisticsChannel)
	<-interrupt

	log.Printf("[%s] Stopping a server...", mockServer.Name)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		log.Printf("[%s] Shutdown error: %s", mockServer.Name, err)
	}

	log.Printf("[%s] Server stopped", mockServer.Name)

	wg.Done()
}
