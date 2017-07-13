package main

import (
	"os"

	"log"

	"github.com/pantheon-systems/riker/pkg/riker"
)

func main() {
	botKey := os.Getenv("SLACK_BOT_TOKEN")
	if botKey == "" {
		log.Fatal("SLACK_BOT_TOKEN env var not set")
	}

	oauthToken := os.Getenv("SLACK_TOKEN")
	if oauthToken == "" {
		log.Fatal("SLACK_TOKEN env var not set")
	}

	b := riker.New(botKey, oauthToken)
	b.Run()
}
