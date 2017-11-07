package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/pokidovea/mimicro/config"
	"github.com/pokidovea/mimicro/statistics"
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

	doneBufferLength := len(serverCollection.Servers)

	if serverCollection.CollectStatistics {
		doneBufferLength++
	}
	statisticsChannel := make(chan statistics.Request)
	statisticsCollector := statistics.Collector{Chan: statisticsChannel}

	done := make(chan bool, doneBufferLength)

	if serverCollection.CollectStatistics {
		go statisticsCollector.Run(done)
	}
	for _, server := range serverCollection.Servers {
		go server.Serve(statisticsChannel, done)
	}

	<-done
}
