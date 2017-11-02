package mock_server

import (
	"fmt"
	"net/http"
	"strconv"
	"time"
)

type Endpoint struct {
	Url      string `json:"url"`
	Response string `json:"response"`
}

type MockServer struct {
	Name      string     `json:"name"`
	Port      int        `json:"port"`
	Endpoints []Endpoint `json:"endpoints"`
}

func (mockServer MockServer) Serve() {
	mux := http.NewServeMux()

	for _, endpoint := range mockServer.Endpoints {
		response := endpoint.Response
		mux.HandleFunc(endpoint.Url, func(w http.ResponseWriter, req *http.Request) {
			fmt.Fprintf(w, response)
		})
	}

	s := &http.Server{
		Addr:           ":" + strconv.Itoa(mockServer.Port),
		Handler:        mux,
		ReadTimeout:    10 * time.Second,
		WriteTimeout:   10 * time.Second,
		MaxHeaderBytes: 1 << 20,
	}

	s.ListenAndServe()
}
