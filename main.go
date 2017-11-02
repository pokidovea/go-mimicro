package main

import (
	"flag"
	"github.com/pokidovea/mimicro/config"
)

func main() {

	configPath := flag.String("config", "", "a path to configuration file")
	flag.Parse()

	serverCollection, err := config.Load(*configPath)

	if err != nil {
		panic(err)
	}

	done := make(chan bool, len(serverCollection.Servers))
	for _, server := range serverCollection.Servers {
		go server.Serve(done)
	}

	<-done
}
