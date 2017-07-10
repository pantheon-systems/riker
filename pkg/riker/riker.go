package riker

import (
	"fmt"
	"io"
	"log"
	"net"
	"strings"
	"sync"

	"google.golang.org/grpc"

	"github.com/davecgh/go-spew/spew"
	"github.com/nlopes/slack"
	"github.com/pantheon-systems/riker/pkg/botpb"
)

// Bot is the bot
type Bot struct {
	rtm  *slack.RTM
	name string

	commands map[string]*Command
	grpc     *grpc.Server

	mutex *sync.Mutex
}

// ClientStream Implemnets the redshirt protobuf server
func (b *Bot) ClientStream(stream botpb.RedShirt_ClientStreamServer) error {
	log.Println("Started  client stream")
	for {
		in, err := stream.Recv()
		log.Println("Received value")
		if err == io.EOF {
			return nil
		}
		if err != nil {
			return err
		}
		spew.Dump(in)
	}
}

// RegisterCommand Implemnets the redshirt protobuf server
func (b *Bot) RegisterCommand(cmd *botpb.Command, stream botpb.RedShirt_RegisterCommandServer) error {
	spew.Dump(b.commands)

	log.Println("registering command")
	spew.Dump(cmd)

	c := &Command{
		command:     cmd.Name,
		description: cmd.Description,
		sendStream:  stream,
		auth: CommandAuth{
			Users:  cmd.Auth.Users,
			Groups: cmd.Auth.Groups,
		},
	}

	b.mutex.Lock()
	//	if _, ok := b.commands[cmd.Name]; ok {
	//		return errors.New("command already registered")
	//	}
	b.commands[cmd.Name] = c
	b.mutex.Unlock()

	spew.Dump(b.commands)
	// sleep forever so we can send commands
	select {}
}

// New is the constroctor for a bot
func New(key string) *Bot {
	grpcServer := grpc.NewServer()

	b := &Bot{
		rtm:      slack.New(key).NewRTM(),
		name:     "riker",
		grpc:     grpcServer,
		commands: make(map[string]*Command, 10),
		mutex:    &sync.Mutex{},
	}

	botpb.RegisterRedShirtServer(grpcServer, b)
	return b
}

// Run starts the bot
func (b *Bot) Run() {
	log.Println("starting slack RTM broker")
	go b.rtm.ManageConnection()

	l, err := net.Listen("tcp", "0.0.0.0:6000")
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}
	log.Println("Listening on tcp://0.0.0.0:6000")
	go b.grpc.Serve(l)
	b.startBroker()
}

/*
func (b Bot) findChannelByName(name string) *slack.Channel {
	for _, ch := range b.rtm.GetInfo().Channels {
		if ch.Name == name {
			return &ch
		}
	}
	return nil
}

func (b Bot) findChannelByID(id string) *slack.Channel {
	for _, ch := range b.rtm.GetInfo().Channels {
		if ch.ID == id {
			return &ch
		}
	}
	return nil
}
*/

func (b *Bot) startBroker() {
	var botID string

	for {
		select {
		case msg := <-b.rtm.IncomingEvents:
			switch ev := msg.Data.(type) {
			case *slack.ConnectedEvent:
				botID = ev.Info.User.ID
				log.Printf("Bot has connected!\n %+v", ev.Info.User)

			case *slack.TeamJoinEvent:
				log.Printf("User Joined: %+v", ev.User)

			case *slack.MessageEvent:
				log.Printf("Message recieved: %+v", ev)
				// ignore messages from other bots or ourself
				if ev.BotID != "" || ev.User == botID {
					continue
				}

				if ev.Text == "help" {
					b.rtm.SendMessage(b.rtm.NewOutgoingMessage("hi I am still being worked on", ev.Channel))
					continue
				}

				for name, cmd := range b.commands {
					// match the message prefix to registered commands
					log.Println("checking for command ", name)
					if strings.HasPrefix(ev.Text, name) {
						log.Println("sending Command")
						err := cmd.sendStream.Send(&botpb.ChatMessage{
							Channel: ev.Channel,
							Message: ev.Text,
						})

						if err != nil {
							log.Println("got error sending command: ", err)
						}
					}

				}

			case *slack.RTMError:
				fmt.Printf("Error: %s\n", ev.Error())

			case *slack.InvalidAuthEvent:
				log.Fatal("Invalid credentials")

			default:
				// Ignore other events..
				//fmt.Println("Unhandled event: ", msg.Type)
			}
		}
	}
}
