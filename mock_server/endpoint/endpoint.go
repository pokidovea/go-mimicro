package endpoint

import (
	"net/http"

	"github.com/pokidovea/mimicro/mock_server/response"
	"github.com/pokidovea/mimicro/stats"
)

type Endpoint struct {
	statChan chan stats.Request
	server   string
	Url      string             `json:"url"`
	GET      *response.Response `json:"GET"`
	POST     *response.Response `json:"POST"`
	PATCH    *response.Response `json:"PATCH"`
	PUT      *response.Response `json:"PUT"`
	DELETE   *response.Response `json:"DELETE"`
}

func (endpoint *Endpoint) CollectStats(statChan chan stats.Request, server string) {
	endpoint.statChan = statChan
	endpoint.server = server
}

func (endpoint Endpoint) GetHandler() func(w http.ResponseWriter, req *http.Request) {
	return func(w http.ResponseWriter, req *http.Request) {
		if endpoint.statChan != nil {
			endpoint.statChan <- stats.Request{endpoint.server, endpoint.Url, req.Method}
		}
		if req.Method == "GET" && endpoint.GET != nil {
			endpoint.GET.WriteResponse(w)
			return
		}
		if req.Method == "POST" && endpoint.POST != nil {
			endpoint.POST.WriteResponse(w)
			return
		}
		if req.Method == "PATCH" && endpoint.PATCH != nil {
			endpoint.PATCH.WriteResponse(w)
			return
		}
		if req.Method == "PUT" && endpoint.PUT != nil {
			endpoint.PUT.WriteResponse(w)
			return
		}
		if req.Method == "DELETE" && endpoint.DELETE != nil {
			endpoint.DELETE.WriteResponse(w)
			return
		}

		http.NotFound(w, req)
	}
}
