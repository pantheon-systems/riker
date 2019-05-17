package riker

import (
	"github.com/pantheon-systems/riker/pkg/botpb"
)

type rikerError string

func (e rikerError) Error() string {
	return string(e)
}

const (
	// ErrorAlreadyRegistered is returned when a registration attempt is made for an
	// already existing command and the Capability deosn't specify forced registration
	ErrorAlreadyRegistered = rikerError("Already registered")

	// ErrorNotRegistered is returned when a client making a request request is not registered.
	ErrorNotRegistered = rikerError("Not registered")

	// ErrorInternalServer is a Generic failure on the server side while trying to process a client (usually a registration)
	ErrorInternalServer = rikerError("Unable to complete request")

	// RedshirtBacklogSize is the redshirt message queue buffer size. how many messages we will backlog for a registerd command
	RedshirtBacklogSize = uint32(20)
)

type Bot interface {
	HealthZ() error
	Run()
	ChatSend(msg, channel string)
	botpb.RikerServer
}
