package mimicro

import (
	"net/http"
)

type HttpHandler = func(w http.ResponseWriter, req *http.Request)

// Endpoint represents an URL, which accepts one or several types of requests
type Endpoint struct {
	URL    string    `json:"url"`
	GET    *Response `json:"GET"`
	POST   *Response `json:"POST"`
	PATCH  *Response `json:"PATCH"`
	PUT    *Response `json:"PUT"`
	DELETE *Response `json:"DELETE"`
}

// GetHandler returns a function to register it as a http handler
func (endpoint Endpoint) GetHandler(logWriter RequestLogWriter, serverName string) HttpHandler {
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

		if response != nil {
			logWriter(serverName, req.URL.String(), req.Method, response.StatusCode)
			response.WriteResponse(w, req)
		} else {
			logWriter(serverName, req.URL.String(), req.Method, http.StatusNotFound)
			http.NotFound(w, req)
		}
	}
}
