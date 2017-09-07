package main

import (
	"os"

	"log"

	"github.com/pantheon-systems/riker/pkg/riker"
)

func main() {
	cli := os.Getenv("RIKER_TERMINAL_MODE")
	if cli != "" {
		b := riker.NewTerminal("localhost:6000")
		b.Run()
	}

	botKey := os.Getenv("SLACK_BOT_TOKEN")
	if botKey == "" {
		log.Fatal("SLACK_BOT_TOKEN env var not set")
	}

	oauthToken := os.Getenv("SLACK_TOKEN")
	if oauthToken == "" {
		log.Fatal("SLACK_TOKEN env var not set")
	}

	tlsFile := os.Getenv("RIKER_TLS_CERT")
	if tlsFile == "" {
		log.Fatal("RIKER_TLS_CERT env var not set")
	}

	caFile := os.Getenv("RIKER_CA_FILE")
	if caFile == "" {
		log.Fatal("RIKER_CA_FILE env var not set")
	}

	b := riker.NewSlackBot(botKey, oauthToken, tlsFile, caFile)
	b.Run()
}
