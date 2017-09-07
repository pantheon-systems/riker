package riker

import (
	"log"
	"net"
	"time"

	"github.com/pantheon-systems/riker/pkg/botpb"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
	"google.golang.org/grpc/keepalive"
)

type TerminalBot struct {
	listenAddr string
	grpc       *grpc.Server
}

func NewTerminal(addr string) Bot {
	k := keepalive.ServerParameters{
		Time:    3 * time.Second,
		Timeout: 15 * time.Second,
	}

	return &TerminalBot{
		listenAddr: addr,
		grpc:       grpc.NewServer(grpc.KeepaliveParams(k)),
	}
}

func (t *TerminalBot) CommandStream(reg *botpb.Registration, stream botpb.Riker_CommandStreamServer) error {
	return nil
}

func (t *TerminalBot) NewRedShirt(ctx context.Context, cap *botpb.Capability) (*botpb.Registration, error) {
	return nil, nil
}

func (t *TerminalBot) NextCommand(ctx context.Context, reg *botpb.Registration) (*botpb.Message, error) {
	return nil, nil
}

func (t *TerminalBot) Send(ctx context.Context, msg *botpb.Message) (*botpb.SendResponse, error) {
	return nil, nil
}

func (t *TerminalBot) SendStream(stream botpb.Riker_SendStreamServer) error {
	return nil
}

func (t *TerminalBot) Run() {
	l, err := net.Listen("tcp", t.listenAddr)
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}
	log.Fatal(t.grpc.Serve(l))
}
