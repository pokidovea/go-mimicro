package mock_server

import (
	"net/http"
	"strconv"
	"time"

	"github.com/pokidovea/mimicro/mock_server/endpoint"
	"github.com/pokidovea/mimicro/statistics"
)

type MockServer struct {
	Name      string              `json:"name"`
	Port      int                 `json:"port"`
	Endpoints []endpoint.Endpoint `json:"endpoints"`
}

func (mockServer MockServer) Serve(statisticsChannel chan statistics.Request, done chan bool) {
	mux := http.NewServeMux()

	for _, endpoint := range mockServer.Endpoints {
		endpoint.CollectStatistics(statisticsChannel, mockServer.Name)
		mux.HandleFunc(endpoint.Url, endpoint.GetHandler())
	}

	s := &http.Server{
		Addr:           ":" + strconv.Itoa(mockServer.Port),
		Handler:        mux,
		ReadTimeout:    10 * time.Second,
		WriteTimeout:   10 * time.Second,
		MaxHeaderBytes: 1 << 20,
	}

	s.ListenAndServe()

	done <- true

}
