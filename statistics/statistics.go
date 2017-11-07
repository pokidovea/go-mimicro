package statistics

import (
	"log"
	"os"
	"os/signal"
	"sync"
)

type Request struct {
	ServerName string `json:"server"`
	Url        string `json:"endpoint"`
	Method     string `json:"method"`
	StatusCode int    `json:"status_code"`
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

func (collector *Collector) Run(wg *sync.WaitGroup) {
	signalChannel := make(chan os.Signal, 1)
	signal.Notify(signalChannel, os.Interrupt)

	defer close(signalChannel)
	defer signal.Stop(signalChannel)
	defer log.Printf("Statistics collector stopped")
	defer wg.Done()

	for {
		select {
		case request, ok := <-collector.Chan:
			if !ok {
				return
			}
			collector.Add(request)
		case <-signalChannel:
			return
		}
	}
}
