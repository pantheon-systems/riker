package main

import (
	"fmt"
	"os"

	"github.com/pantheon-systems/riker/cmd"
	"github.com/pantheon-systems/riker/pkg/healthz"
	"github.com/pantheon-systems/riker/pkg/helpers"
)

var healthzConfig healthz.Config

func init() {
	healthzConfig.BindPort = helpers.GetEnvAsInt("HEALTHZ_PORT", 8080)
	healthzConfig.BindAddr = helpers.GetEnv("HEALTHZ_ADDR", "0.0.0.0")
}

func main() {
	healthServer, err := healthz.New(healthzConfig)
	if err != nil {
		fmt.Println(err)
		os.Exit(-1)
	}
	go healthServer.StartHealthz()
	cmd.Execute()
}
