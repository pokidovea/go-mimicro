package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"sync"

	"github.com/pokidovea/mimicro/management"
	"github.com/pokidovea/mimicro/mockServer"
)

func checkConfig(configPath string) error {
	err := mockServer.CheckConfig(configPath)

	if err == nil {
		fmt.Println("Config is valid")
		return nil
	}

	fmt.Printf("Config is not valid. See errors below: \n %s \n", err.Error())
	return err
}

func main() {

	configPath := flag.String("config", "", "a path to configuration file")
	checkConf := flag.Bool("check", false, "validates passed config")
	managementPort := flag.Int("management-port", 4444, "port for the management server")
	collectStatistics := flag.Bool(
		"collect-statistics", false, "pass this flag if you want to collect statistics of requests",
	)

	flag.Parse()

	err := checkConfig(*configPath)

	if err != nil {
		os.Exit(1)
	}

	if *checkConf == true {
		os.Exit(0)
	}

	serverCollection, err := mockServer.Load(*configPath)

	if err != nil {
		log.Printf(err.Error())
		os.Exit(1)
	}

	var wg sync.WaitGroup

	managementServer := management.NewServer(*managementPort, *collectStatistics)
	wg.Add(1)
	go managementServer.Serve(&wg)

	for _, server := range serverCollection.Servers {
		wg.Add(1)
		go server.Serve(managementServer.WriteRequestLog, &wg)
	}

	wg.Wait()
	log.Printf("Mimicro successfully down")
}
