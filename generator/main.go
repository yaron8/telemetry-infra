package main

import (
	"log"

	"github.com/yaron8/telemetry-infra/generator/bootstrap"
)

func main() {
	bootstrap, err := bootstrap.NewBootstrap()
	if err != nil {
		log.Fatalf("Failed to create bootstrap: %v", err)
	}

	if err := bootstrap.Start(); err != nil {
		panic(err)
	}
}
