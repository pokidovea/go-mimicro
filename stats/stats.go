package stats

import (
	"log"
	"os"
	"os/signal"
)

type Request struct {
	Server   string `json:"server"`
	Endpoint string `json:"endpoint"`
	Method   string `json:"method"`
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

func (collector *Collector) Run() {
	for request := range collector.Chan {
		collector.Add(request)
	}
	log.Printf("Run exit")
}

func (collector Collector) HandleExit(done chan bool) {
	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, os.Interrupt)
	go func() {
		for _ = range signalChan {
			close(collector.Chan)
			done <- true
		}
	}()
}
