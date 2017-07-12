package main

import (
	"context"
	"io"
	"log"

	"github.com/davecgh/go-spew/spew"
	"github.com/pantheon-systems/riker/pkg/botpb"
	"google.golang.org/grpc"
)

func main() {

	conn, err := grpc.Dial("localhost:6000", grpc.WithInsecure())
	if err != nil {
		log.Fatal(err)
	}
	defer conn.Close()

	client := botpb.NewRikerClient(conn)
	cap := &botpb.Capability{
		Name:        "echo",
		Usage:       "send text we will send it back",
		Description: "This command will echo what you say back to you",
		Auth: &botpb.CommandAuth{
			Users:  []string{"jesse@pantheon.io"},
			Groups: []string{"engineering"},
		},
	}

	reg, err := client.NewRedShirt(context.Background(), cap)
	if err != nil {
		log.Fatal("wtf this shouln't fail: ", err.Error())
	}

	if !reg.CapabilityApplied {
		log.Println("someone already registered this command, but that's fine with me.")
	}

	stream, err := client.CommandStream(context.Background(), reg)
	if err != nil {
		log.Fatal("error talking to server", err)
	}

	for {
		msg, err := stream.Recv()
		if err == io.EOF {
			continue
		}
		if err != nil {
			log.Fatalf(" error %+v = %v", client, err)
		}

		log.Printf("Got message: %+v\n", msg)
		resp, err := client.Send(context.Background(), msg)
		if err != nil {
			log.Println("Error sending: ", err)
		}

		spew.Dump(resp)

	}
}
