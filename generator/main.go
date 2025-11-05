package main

import "github.com/yaron8/telemetry-infra/generator/bootstrap"

func main() {
	bootstrap := bootstrap.NewBootstrap()

	if err := bootstrap.StartServer(); err != nil {
		panic(err)
	}
}
