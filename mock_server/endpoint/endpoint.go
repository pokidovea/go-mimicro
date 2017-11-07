package endpoint

import (
	"net/http"

	"github.com/pokidovea/mimicro/mock_server/response"
	"github.com/pokidovea/mimicro/statistics"
)

type Endpoint struct {
	statisticsChannel chan statistics.Request
	serverName        string
	Url               string             `json:"url"`
	GET               *response.Response `json:"GET"`
	POST              *response.Response `json:"POST"`
	PATCH             *response.Response `json:"PATCH"`
	PUT               *response.Response `json:"PUT"`
	DELETE            *response.Response `json:"DELETE"`
}

func (endpoint *Endpoint) CollectStatistics(statisticsChannel chan statistics.Request, serverName string) {
	endpoint.statisticsChannel = statisticsChannel
	endpoint.serverName = serverName
}

func (endpoint Endpoint) GetHandler() func(w http.ResponseWriter, req *http.Request) {
	return func(w http.ResponseWriter, req *http.Request) {
		if endpoint.statisticsChannel != nil {
			endpoint.statisticsChannel <- statistics.Request{endpoint.serverName, endpoint.Url, req.Method}
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
