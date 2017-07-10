package main

import (
	"os"

	"log"

	"github.com/pantheon-systems/riker/pkg/riker"
)

func main() {
	botKey := os.Getenv("SLACK_TOKEN")
	if botKey == "" {
		log.Fatal("SLACK_TOKEN env var not set")
	}

	b := riker.New(botKey)
	b.Run()
}
