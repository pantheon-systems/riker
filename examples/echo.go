package main

import (
	"context"
	"io"
	"log"

	"github.com/pantheon-systems/riker/pkg/botpb"
	"google.golang.org/grpc"
)

func main() {

	conn, err := grpc.Dial("localhost:6000", grpc.WithInsecure())
	if err != nil {
		log.Fatal(err)
	}
	defer conn.Close()

	client := botpb.NewRedShirtClient(conn)

	cmd := &botpb.Command{
		Name:        "echo",
		Usage:       "send text we will send it back",
		Description: "This command will echo what you say back to you",
		Auth: &botpb.CommandAuth{
			Users:  []string{"jesse@pantheon.io"},
			Groups: []string{"engineering"},
		},
	}

	stream, err := client.RegisterCommand(context.Background(), cmd)
	if err != nil {
		log.Fatal("error talking to server", err)
	}

	for {
		msg, err := stream.Recv()
		if err == io.EOF {
			break
		}
		if err != nil {
			log.Fatalf("%v.ListFeatures(_) = _, %v", client, err)
		}
		log.Println("got msg", msg)
	}
}
