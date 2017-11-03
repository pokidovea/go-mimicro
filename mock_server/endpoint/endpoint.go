package endpoint

import (
	"github.com/pokidovea/mimicro/mock_server/response"
	"net/http"
)

type Endpoint struct {
	Url  string            `json:"url"`
	GET  response.Response `json:"GET"`
	POST response.Response `json:"POST"`
}

func (endpoint Endpoint) GetHandler() func(w http.ResponseWriter, req *http.Request) {
	return func(w http.ResponseWriter, req *http.Request) {
		if req.Method == "GET" && endpoint.GET.Body != "" {
			endpoint.GET.WriteResponse(w)
			return
		}
		if req.Method == "POST" && endpoint.POST.Body != "" {
			endpoint.POST.WriteResponse(w)
			return
		}
		http.NotFound(w, req)
	}
}
