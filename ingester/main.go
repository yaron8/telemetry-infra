package main

import (
	"fmt"

	"github.com/yaron8/telemetry-infra/ingester/bootstrap"
)

func main() {
	bootstrap, err := bootstrap.NewBootstrap()
	if err != nil {
		panic(fmt.Sprintf("Failed to create ingester bootstrap: %v", err))
	}

	if err := bootstrap.Start(); err != nil {
		panic(fmt.Sprintf("Failed to start ingester server: %v", err))
	}
}
