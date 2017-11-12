package mockServer

import (
	"log"
	"net/http"

	"github.com/pokidovea/mimicro/statistics"
)

// Endpoint represents an URL, wich accepts one ore several types of requests
type Endpoint struct {
	statisticsChannel chan statistics.Request
	serverName        string
	URL               string    `json:"url"`
	GET               *Response `json:"GET"`
	POST              *Response `json:"POST"`
	PATCH             *Response `json:"PATCH"`
	PUT               *Response `json:"PUT"`
	DELETE            *Response `json:"DELETE"`
}

// CollectStatistics sets statisticsChannel and serverName for the endpoint
// TODO: Give better name for this function
func (endpoint *Endpoint) CollectStatistics(statisticsChannel chan statistics.Request, serverName string) {
	endpoint.statisticsChannel = statisticsChannel
	endpoint.serverName = serverName
}

// GetHandler returns a function to register it as handler in mux
func (endpoint Endpoint) GetHandler() func(w http.ResponseWriter, req *http.Request) {
	return func(w http.ResponseWriter, req *http.Request) {
		var response *Response

		if req.Method == "GET" && endpoint.GET != nil {
			response = endpoint.GET
		} else if req.Method == "POST" && endpoint.POST != nil {
			response = endpoint.POST
		} else if req.Method == "PATCH" && endpoint.PATCH != nil {
			response = endpoint.PATCH
		} else if req.Method == "PUT" && endpoint.PUT != nil {
			response = endpoint.PUT
		} else if req.Method == "DELETE" && endpoint.DELETE != nil {
			response = endpoint.DELETE
		}

		statisticsRequest := statistics.Request{
			ServerName: endpoint.serverName,
			Url:        req.URL.String(),
			Method:     req.Method,
		}

		if response != nil {
			response.WriteResponse(w, req)
			statisticsRequest.StatusCode = response.StatusCode
		} else {
			statisticsRequest.StatusCode = http.StatusNotFound
			http.NotFound(w, req)
		}
		log.Printf("Requested %s \n", statisticsRequest)

		if endpoint.statisticsChannel != nil {
			endpoint.statisticsChannel <- statisticsRequest
		}
	}
}
