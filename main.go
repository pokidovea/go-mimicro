package main

import (
	"flag"
	"github.com/pokidovea/mimicro/config"
)

func main() {

	configPath := flag.String("config", "", "a path to configuration file")
	flag.Parse()

	servers, err := config.Load(*configPath)

	if err != nil {
		panic(err)
	}
	servers.Servers[0].Serve()
}
