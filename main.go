package main

import (
	"flag"
	"fmt"
	"github.com/pokidovea/mimicro/config"
	"os"
)

func checkConfig(configPath string) error {
	err := config.CheckConfig(configPath)

	if err == nil {
		fmt.Println("Config is valid")
		return nil
	} else {
		fmt.Printf("Config is not valid. See errors below: \n %s \n", err.Error())
		return err
	}
}

func main() {

	configPath := flag.String("config", "", "a path to configuration file")
	checkConf := flag.Bool("check", false, "validates passed config")
	flag.Parse()

	err := checkConfig(*configPath)

	if err != nil {
		os.Exit(1)
	}

	if *checkConf == true {
		os.Exit(0)
	}

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
