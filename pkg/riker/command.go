package riker

import (
	"github.com/pantheon-systems/riker/pkg/botpb"
)

// CommandAuth is a type for holding the authorization
type CommandAuth struct {
	Users  []string
	Groups []string
}

// Command structure contains a command command, description and handler
type Command struct {
	command     string
	description string
	sendStream  botpb.RedShirt_RegisterCommandServer
	auth        CommandAuth
}
