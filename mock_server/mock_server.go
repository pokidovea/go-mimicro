package mock_server

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"sync"
	"time"

	"github.com/pokidovea/mimicro/mock_server/endpoint"
	"github.com/pokidovea/mimicro/statistics"
)

type MockServer struct {
	Name      string              `json:"name"`
	Port      int                 `json:"port"`
	Endpoints []endpoint.Endpoint `json:"endpoints"`
}

func (mockServer MockServer) startHttpServer(statisticsChannel chan statistics.Request) *http.Server {
	mux := http.NewServeMux()

	for _, endpoint := range mockServer.Endpoints {
		endpoint.CollectStatistics(statisticsChannel, mockServer.Name)
		mux.HandleFunc(endpoint.Url, endpoint.GetHandler())
	}

	srv := &http.Server{
		Addr:           ":" + strconv.Itoa(mockServer.Port),
		Handler:        mux,
		ReadTimeout:    10 * time.Second,
		WriteTimeout:   10 * time.Second,
		MaxHeaderBytes: 1 << 20,
	}

	go func() {
		if err := srv.ListenAndServe(); err != nil {
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
	ctx, _ := context.WithTimeout(context.Background(), 1*time.Second)
	srv.Shutdown(ctx)
	log.Printf("[%s] Server stopped", mockServer.Name)

	wg.Done()
}
