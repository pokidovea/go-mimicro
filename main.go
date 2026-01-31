package main

import (
	"flag"
	"fmt"
	"log"
	"os"

	"mimicro/internal/server"
	"mimicro/internal/spec"
)

func main() {
	var specPath string
	var port int

	flag.StringVar(&specPath, "spec", "", "Path or URL to OpenAPI spec file (yaml/json)")
	flag.IntVar(&port, "port", 8080, "Port to run the mock server on")
	flag.Parse()

	if specPath == "" {
		fmt.Println("Error: -spec flag is required")
		flag.Usage()
		os.Exit(1)
	}

	log.Printf("Loading OpenAPI spec from: %s", specPath)

	loader := spec.NewLoader()
	apiSpec, err := loader.Load(specPath)
	if err != nil {
		log.Fatalf("Failed to load OpenAPI spec: %v", err)
	}

	log.Printf("Successfully loaded OpenAPI spec: %s %s", apiSpec.Info.Title, apiSpec.Info.Version)
	log.Printf("Starting mock server on port %d", port)

	mockServer := server.New(apiSpec, port)
	if err := mockServer.Start(); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
