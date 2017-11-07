package statistics

import (
	"log"
	"os"
	"os/signal"
)

type Request struct {
	ServerName  string `json:"server"`
	Url         string `json:"endpoint"`
	Method      string `json:"method"`
	StatusCode  int    `json:"status_code"`
	ContentType string `json:"content_type"`
}

type Collector struct {
	Chan     chan Request
	requests map[Request]int
}

func (collector *Collector) Add(request Request) {
	if collector.requests == nil {
		collector.requests = make(map[Request]int)
	}
	collector.requests[request]++
	log.Printf("Added %s (%d)\n", request, collector.requests[request])
}

func (collector Collector) Get(request Request) int {
	output := collector.requests[request]
	return output
}

func (collector *Collector) Run(done chan bool) {
	signalChannel := make(chan os.Signal, 1)
	defer close(signalChannel)
	signal.Notify(signalChannel, os.Interrupt)

	for {
		select {
		case request, ok := <-collector.Chan:
			if !ok {
				done <- true
				log.Printf("Statistics collector stop")
				return
			}
			collector.Add(request)
		case <-signalChannel:
			close(collector.Chan)
			done <- true
			log.Printf("Statistics collector stop")
			return
		}
	}
}
