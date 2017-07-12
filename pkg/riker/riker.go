package riker

import (
	"fmt"
	"io"
	"log"
	"net"
	"strings"
	"sync"

	"golang.org/x/net/context"

	"google.golang.org/grpc"

	"github.com/davecgh/go-spew/spew"
	"github.com/nlopes/slack"
	"github.com/pantheon-systems/riker/pkg/botpb"
)

type rikerError string

func (e rikerError) Error() string {
	return string(e)
}

// ErrorAlreadyRegistered is returned when a registration attempt is made for an
// already existing command and the Capability deosn't specify forced registration
const ErrorAlreadyRegistered = rikerError("Already registered")

// ErrorNotRegistered is returned when a client making a request request is not registered.
const ErrorNotRegistered = rikerError("Not registered")

// Bot is the bot
type Bot struct {
	rtm  *slack.RTM
	name string

	// redshirts holds state of commands that map to client registrations
	redshirts map[string]*redShirtRegistration
	*sync.RWMutex

	channels map[string]bool
	grpc     *grpc.Server
}

// holds info on a connected client so we can send it data
type redShirtRegistration struct {
	queue chan *botpb.Message
	cap   *botpb.Capability
}

// NewRedShirt Implemnets the riker protobuf server
func (b *Bot) NewRedShirt(ctx context.Context, cap *botpb.Capability) (*botpb.Registration, error) {
	log.Println("registered client ", cap)

	// map is not goroutine safe
	b.Lock()
	defer b.Unlock()
	resp := &botpb.Registration{
		Name:              cap.Name,
		CapabilityApplied: true,
	}

	spew.Dump(b.redshirts)
	// we want to register commands if they are requesting a force registration, othewise it should error
	if reg, ok := b.redshirts[cap.Name]; ok {
		spew.Dump(b.redshirts)
		if !cap.ForcedRegistration {
			resp.CapabilityApplied = false
			return resp, nil
		}

		// this should signal all existing write pumps to clients from the commandStream to shutdown and return.
		close(reg.queue)
		reg.queue = make(chan *botpb.Message, int32(cap.BufferSize))

		// we want newest clients registering with force to apply their capailities
		// so that it makes it easier for them to upgrade themselves.
		reg.cap = cap
		return resp, nil
	}

	// Happy path? comand isn't registered
	reg := redShirtRegistration{}
	reg.cap = cap
	reg.queue = make(chan *botpb.Message, int32(cap.BufferSize))
	b.redshirts[cap.Name] = &reg

	return resp, nil
}

// NextCommand is the call a client makes to pull the next command from rikers command buffer
func (b *Bot) NextCommand(ctx context.Context, reg *botpb.Registration) (*botpb.Message, error) {
	// TODO: CommandStream and this method do the same registration checking, could be refactored if we do it more than this.
	b.RLock()
	rs, ok := b.redshirts[reg.Name]
	b.RUnlock()
	if !ok {
		return nil, ErrorNotRegistered
	}

	select {
	case m := <-rs.queue:
		return m, nil
	case <-ctx.Done():
		return nil, ctx.Err()
	}
}

// CommandStream is the call a client makes to setup a Push stream from riker -> client
func (b *Bot) CommandStream(reg *botpb.Registration, stream botpb.Riker_CommandStreamServer) error {
	b.RLock()
	rs, ok := b.redshirts[reg.Name]
	b.RUnlock()
	if !ok {
		return ErrorNotRegistered
	}

	for m := range rs.queue {
		err := stream.Send(m)
		if err != nil {
			select {
			case rs.queue <- m:
			default:
				msg := b.rtm.NewOutgoingMessage("Comunicator malfunction while talking to redshirt.", m.Channel)
				go b.rtm.SendMessage(msg)
			}
			break
		}
	}
	return nil
}

// Send is the call a client makes to send a message back to riker
func (b *Bot) Send(ctx context.Context, msg *botpb.Message) (*botpb.SendResponse, error) {
	m := b.rtm.NewOutgoingMessage(msg.Payload, msg.Channel)
	b.rtm.SendMessage(m)
	return &botpb.SendResponse{Ok: true}, nil
}

// SendStream is the call a client makes to send a stream message back to riker
func (b *Bot) SendStream(stream botpb.Riker_SendStreamServer) error {
	// pump for messages to slack
	for {
		in, err := stream.Recv()
		log.Println("Received value")
		if err == io.EOF {
			return nil
		}
		if err != nil {
			return err
		}
		msg := b.rtm.NewOutgoingMessage(in.Payload, in.Channel)
		b.rtm.SendMessage(msg)
	}

	return nil
}

// New is the constroctor for a bot
func New(key string) *Bot {
	grpcServer := grpc.NewServer()

	b := &Bot{
		rtm:       slack.New(key).NewRTM(),
		name:      "riker",
		grpc:      grpcServer,
		redshirts: make(map[string]*redShirtRegistration, 10),
		channels:  make(map[string]bool, 100),
	}
	b.RWMutex = &sync.RWMutex{}

	botpb.RegisterRikerServer(grpcServer, b)
	return b
}

// Run starts the bot
func (b *Bot) Run() {
	log.Println("starting slack RTM broker")
	// TODO: check nil maybe return errors, and do that in New instead
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
				spew.Dump(ev)
				log.Printf("Message recieved: %+v", ev)
				// ignore messages from other bots or ourself
				if ev.BotID != "" || ev.User == botID {
					continue
				}

				if ev.Type != "message" {
					continue
				}

				// for now strip this out
				msgSlice := strings.Split(ev.Text, " ")

				//  we need to detect  a direct message from a channel message, and unfortunately slack doens't make that super awesome
				b.Lock()
				isChan, ok := b.channels[ev.Channel]
				if !ok {
					_, err := b.rtm.GetChannelInfo(ev.Channel)
					if err != nil && err.Error() != "channel_not_found" {
						log.Println("not dm not channel: ", err)
						continue
					}

					if err == nil {
						isChan = true
					}

					b.channels[ev.Channel] = isChan
				}
				b.Unlock()

				botString := "<@" + botID + ">"
				if msgSlice[0] != botString {
					if isChan {
						continue
					}
					// normalize the msgSlice, total garbage.. but CBF
					msgSlice = append([]string{botString}, msgSlice...)
				}

				// ignore when someone addresses us without a command
				if len(msgSlice) < 2 {
					continue
				}

				if msgSlice[1] == "help" {
					msg := b.rtm.NewOutgoingMessage("no help for you", ev.Channel)
					go b.rtm.SendMessage(msg)
					continue
				}

				// match the message prefix to registered commands
				cmdName := msgSlice[1]
				log.Println("checking for command ", cmdName)
				b.RLock()
				rsReg, ok := b.redshirts[cmdName]
				b.RUnlock()
				if !ok {
					msg := b.rtm.NewOutgoingMessage("Sorry, that redshirt has not repoorted for duty. I can't complete the request", ev.Channel)
					go b.rtm.SendMessage(msg)
					continue
				}

				msg := &botpb.Message{
					Channel:   ev.Channel,
					Timestamp: ev.Timestamp,
					ThreadTs:  ev.ThreadTimestamp,
					Payload:   ev.Text,
				}

				log.Println("sending Command", msg)
				select {
				case rsReg.queue <- msg:
				default:
					log.Println("Couldn't send to internal slack message queue.")
					msg := b.rtm.NewOutgoingMessage("Sorry, redshirt supply is low. Couldn't complete your request.", ev.Channel)
					go b.rtm.SendMessage(msg)
					break
				}

			case *slack.RTMError:
				fmt.Printf("Error: %s\n", ev.Error())

			case *slack.InvalidAuthEvent:
				log.Fatal("Invalid credentials")

			default:
				// Ignore other events..
				fmt.Println("Unhandled event: ", msg.Type)
			}
		}
	}
}
