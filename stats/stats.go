package stats

type Request struct {
	Server   string `json:"server"`
	Endpoint string `json:"endpoint"`
	Method   string `json:"method"`
}

type Stats struct {
	Requests []Request `json:"requests"`
}

func (stats *Stats) Add(request Request) {
	stats.Requests = append(stats.Requests, request)
}

func (stats Stats) Get(request Request) int {
	output := 0
	for _, req := range stats.Requests {
		if req == request {
			output++
		}
	}
	return output
}
