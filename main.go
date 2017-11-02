package main

import (
	"github.com/pokidovea/mimicro/config"
)

func main() {

	servers, err := config.Load("/home/pokidovea/Dropbox/projects/go/src/github.com/pokidovea/mimicro/config/main.yml")

	if err != nil {
		panic(err)
	}
	servers.Servers[0].Serve()
}
