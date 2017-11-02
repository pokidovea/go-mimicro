package mock_server

import (
	"fmt"
	"net/http"
	"strconv"
	"time"
)

type ResponseMethod struct {
	Response    string `json:"response"`
	ContentType string `json:"content_type"`
}

type Endpoint struct {
	Url  string         `json:"url"`
	GET  ResponseMethod `json:"GET"`
	POST ResponseMethod `json:"POST"`
}

type MockServer struct {
	Name      string     `json:"name"`
	Port      int        `json:"port"`
	Endpoints []Endpoint `json:"endpoints"`
}

func (responseMethod ResponseMethod) writeResponse(w http.ResponseWriter) {
	w.Header().Set("Content-Type", responseMethod.ContentType)
	w.WriteHeader(http.StatusOK)

	fmt.Fprintf(w, responseMethod.Response)
}

func (endpoint Endpoint) getHandler() func(w http.ResponseWriter, req *http.Request) {
	return func(w http.ResponseWriter, req *http.Request) {
		if req.Method == "GET" && endpoint.GET.Response != "" {
			endpoint.GET.writeResponse(w)
			return
		}
		if req.Method == "POST" && endpoint.POST.Response != "" {
			endpoint.POST.writeResponse(w)
			return
		}
		http.NotFound(w, req)
	}
}

func (mockServer MockServer) Serve(done chan bool) {
	mux := http.NewServeMux()

	for _, endpoint := range mockServer.Endpoints {
		mux.HandleFunc(endpoint.Url, endpoint.getHandler())
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
