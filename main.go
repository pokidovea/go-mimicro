package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"sync"

	"github.com/equinox-io/equinox"
	"github.com/pokidovea/mimicro/mimicro"
)

const appID = "app_cub6zaUSQM5"
const appVersion = "0.1.1"

var publicKey = []byte(`
-----BEGIN ECDSA PUBLIC KEY-----
MHYwEAYHKoZIzj0CAQYFK4EEACIDYgAE+I1EKZgg9I9/jZYUSAafZmGtS2QKx/6m
qiuY5GqpQR4YnJxMe9vs/xJZiK+pjD+dSJgqbMTHNQlqdDdngxk7ncwJ7lyD6fEb
CLTkXUVQ2EIDOH6GSBIQZM1sY98lsJ8b
-----END ECDSA PUBLIC KEY-----
`)

func equinoxUpdate() error {
	var opts equinox.Options
	if err := opts.SetPublicKeyPEM(publicKey); err != nil {
		return err
	}

	// check for the update
	resp, err := equinox.Check(appID, opts)
	switch {
	case err == equinox.NotAvailableErr:
		fmt.Println("No update available, already at the latest version")
		return nil
	case err != nil:
		fmt.Println("Update failed:", err)
		return err
	}

	// fetch the update and apply it
	err = resp.Apply()
	if err != nil {
		return err
	}

	fmt.Printf("Updated to new version: %s\n", resp.ReleaseVersion)
	return nil
}

func checkConfig(configPath string) error {
	err := mimicro.CheckConfig(configPath)

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
	update := flag.Bool("update", false, "check for a new version and update")
	version := flag.Bool("version", false, "current version")

	flag.Parse()

	if *version {
		fmt.Println(appVersion)
		os.Exit(0)
	}

	if *update {
		err := equinoxUpdate()
		if err != nil {
			log.Printf(err.Error())
			os.Exit(1)
		}
		os.Exit(0)
	}

	err := checkConfig(*configPath)

	if err != nil {
		os.Exit(1)
	}

	if *checkConf == true {
		os.Exit(0)
	}

	serverCollection, err := mimicro.LoadConfig(*configPath)

	if err != nil {
		log.Printf(err.Error())
		os.Exit(1)
	}

	var wg sync.WaitGroup

	managementServer := mimicro.NewManagementServer(*managementPort, *collectStatistics)
	wg.Add(1)
	go managementServer.Serve(&wg)

	for _, server := range serverCollection.Servers {
		wg.Add(1)
		go server.Serve(managementServer.WriteRequestLog, &wg)
	}

	wg.Wait()
	log.Printf("Mimicro successfully down")
}
