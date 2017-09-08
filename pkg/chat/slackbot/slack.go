package slackbot

import (
	"crypto/tls"
	"io"
	"log"
	"net"
	"strings"
	"sync"
	"time"

	"golang.org/x/net/context"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/keepalive"

	"github.com/davecgh/go-spew/spew"
	grpc_auth "github.com/grpc-ecosystem/go-grpc-middleware/auth"
	"github.com/nlopes/slack"
	"github.com/pantheon-systems/go-certauth/certutils"
	"github.com/pantheon-systems/riker/pkg/botpb"
	"github.com/pantheon-systems/riker/pkg/riker"
)

// holds info on a connected client (redshirt) so we can send it data
type redshirtRegistration struct {
	queue      chan *botpb.Message
	capability *botpb.Capability
}

// SlackBot is the slack adaptor for riker. It implments the riker.Bot interface for bridging redshirts
// to chat platforms
type SlackBot struct {
	name string
	rtm  *slack.RTM
	api  *slack.Client

	users  sync.Map
	groups []slack.UserGroup

	// redshirts holds state of commands that map to client registrations
	redshirts map[string]*redshirtRegistration
	channels  map[string]bool

	grpc *grpc.Server
	*sync.RWMutex
}

// NewRedShirt implements the riker protobuf server
func (b *SlackBot) NewRedShirt(ctx context.Context, cap *botpb.Capability) (*botpb.Registration, error) {
	log.Println("Registering client ", cap)
	b.Lock()
	defer b.Unlock()
	resp := &botpb.Registration{
		Name:              cap.Name,
		CapabilityApplied: true,
	}

	// Check for registration. If it already exists apply the registrations capabilities to it
	if r, ok := b.redshirts[cap.Name]; ok {
		r.capability = cap
		return resp, nil
	}

	// New registration so we should cary on
	reg := redshirtRegistration{}
	reg.queue = make(chan *botpb.Message, riker.RedshirtBacklogSize)
	reg.capability = cap
	b.redshirts[cap.Name] = &reg

	return resp, nil
}

// NextCommand is the call a client makes to pull the next command from rikers command buffer
func (b *SlackBot) NextCommand(ctx context.Context, reg *botpb.Registration) (*botpb.Message, error) {
	// TODO: CommandStream and this method do the same registration checking, could be refactored if we do it more than this.
	b.RLock()
	rs, ok := b.redshirts[reg.Name]
	b.RUnlock()
	if !ok {
		return nil, riker.ErrorNotRegistered
	}

	select {
	case m := <-rs.queue:
		return m, nil
	case <-ctx.Done():
		return nil, ctx.Err()
	}
}

// CommandStream is the call a client makes to setup a Push stream from riker -> client
func (b *SlackBot) CommandStream(reg *botpb.Registration, stream botpb.Riker_CommandStreamServer) error {
	b.RLock()
	rs, ok := b.redshirts[reg.Name]
	b.RUnlock()
	if !ok {
		return riker.ErrorNotRegistered
	}

	for m := range rs.queue {
		err := stream.Send(m)
		if err != nil {
			select {
			case rs.queue <- m:
			default:
				go b.ChatSend("Communicator malfunction while talking to redshirt.", m.Channel)
			}
			break
		}
	}
	return nil
}

func (b *SlackBot) ChatSend(msg, channel string) {
	m := b.rtm.NewOutgoingMessage("Communicator malfunction while talking to redshirt.", channel)
	b.rtm.SendMessage(m)
}

// Send is the call a client makes to send a message back to riker
func (b *SlackBot) Send(ctx context.Context, msg *botpb.Message) (*botpb.SendResponse, error) {
	m := b.rtm.NewOutgoingMessage(msg.Payload, msg.Channel)
	m.ThreadTimestamp = msg.ThreadTs
	b.rtm.SendMessage(m)
	return &botpb.SendResponse{Ok: true}, nil
}

// SendStream is the call a client makes to send a stream message back to riker
func (b *SlackBot) SendStream(stream botpb.Riker_SendStreamServer) error {
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
		msg.ThreadTimestamp = in.ThreadTs
		b.rtm.SendMessage(msg)
	}
}

// New is the constroctor for a bot
func New(botKey, token, tlsFile, caFile string) riker.Bot {
	cert, err := certutils.LoadKeyCertFiles(tlsFile, tlsFile)
	if err != nil {
		log.Fatalf("Could not load TLS cert '%s': %s", tlsFile, err.Error())
	}
	caPool, err := certutils.LoadCACertFile(caFile)
	if err != nil {
		log.Fatalf("Could not load CA cert '%s': %s", caFile, err.Error())
	}
	// TODO: use CertReloader from certutils
	tlsConfig := certutils.NewTLSConfig(certutils.TLSConfigModern)
	tlsConfig.ClientCAs = caPool
	tlsConfig.ClientAuth = tls.RequireAndVerifyClientCert
	tlsConfig.Certificates = []tls.Certificate{cert}

	k := keepalive.ServerParameters{
		// After a duration of this time if the server doesn't see any activity it pings the client to see if the transport is still alive.
		// The grpc default value is 2 hours.
		Time: 3 * time.Second,

		// After having pinged for keepalive check, the server waits for a duration of Timeout and if no activity is seen even after that
		// the connection is closed.
		// The grpc default value is 20 seconds.
		Timeout: 15 * time.Second,
	}

	grpcServer := grpc.NewServer(
		grpc.KeepaliveParams(k),
		grpc.StreamInterceptor(grpc_auth.StreamServerInterceptor(riker.AuthOU)),
		grpc.UnaryInterceptor(grpc_auth.UnaryServerInterceptor(riker.AuthOU)),
		grpc.Creds(credentials.NewTLS(tlsConfig)),
	)

	b := &SlackBot{
		rtm:       slack.New(botKey).NewRTM(),
		api:       slack.New(token),
		name:      "riker",
		grpc:      grpcServer,
		redshirts: make(map[string]*redshirtRegistration, 10),
		channels:  make(map[string]bool, 100),
	}
	b.RWMutex = &sync.RWMutex{}
	//b.rtm.SetDebug(true)

	botpb.RegisterRikerServer(grpcServer, b)
	return b
}

// Run starts the bot
func (b *SlackBot) Run() {
	log.Println("starting slack RTM broker")
	// TODO: check nil maybe return errors, and do that in New instead
	go b.rtm.ManageConnection()

	l, err := net.Listen("tcp", "0.0.0.0:6000")
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}
	log.Println("Listening on tcp://0.0.0.0:6000")
	go func() {
		log.Fatal(b.grpc.Serve(l))
	}()
	b.startBroker()
}

func (b *SlackBot) startBroker() {
	var botID string

	for {
		msg := <-b.rtm.IncomingEvents
		switch ev := msg.Data.(type) {
		case *slack.ConnectedEvent:
			var err error
			botID = ev.Info.User.ID

			usersSlice, err := b.rtm.GetUsers()
			if err != nil {
				log.Fatal(err)
			}
			for _, u := range usersSlice {
				b.users.Store(u.ID, u)
			}

			b.groups, err = b.api.GetUserGroups()
			if err != nil {
				log.Fatal(err)
			}

			log.Printf("Riker has connected to Slack! %+v", ev.Info.User)

		case *slack.TeamJoinEvent:
			log.Printf("User Joined: %+v", ev.User)
			b.users.Store(ev.User.ID, ev.User)

		case *slack.MessageEvent:
			//	spew.Dump(ev)
			// ignore messages from other bots or ourself
			// XXX: in order to avoid responding to the bot's own messages, it appears we need to check for an empty ev.User
			//log.Printf("botID: %s, User: %s", ev.BotID, ev.User)
			if ev.BotID != "" || ev.User == botID || ev.User == "" {
				continue
			}

			if ev.Type != "message" {
				continue
			}

			log.Printf("Message recieved: %+v", ev)

			// for now strip this out and convert the text message into an array of words for easier parsing
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

			// fix the message so that it is the same no matter if the message came via DM or a channel
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
				// TODO: implement help using the registered Capability's and their name / description / usage
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
				msg := b.rtm.NewOutgoingMessage("Sorry, that redshirt has not reported for duty. I can't complete the request", ev.Channel)
				go b.rtm.SendMessage(msg)
				continue
			}

			// Verify this person is allowed to run this command, cause unauthed commands are a violation of starfleet protocols.
			log.Println("Checking user authentication", ev.User)

			auth := false
			for _, ua := range rsReg.capability.Auth.Users {
				auth = b.idHasEmail(ev.User, ua)
				if auth {
					break
				}
			}

			for _, ga := range rsReg.capability.Auth.Groups {
				auth = auth || b.idInGroup(ev.User, ga)
				if auth {
					break
				}
			}

			if !auth {
				msg := b.rtm.NewOutgoingMessage("Computer reports: 'ACCESS DENIED'", ev.Channel)
				go b.rtm.SendMessage(msg)
				continue
			}

			// if this was a direct message we disabled threaded responses even if the redshirt requested
			// threads. This is because Slack will duplicate the response in the Thread view and the direct message view.
			threadTs := ev.ThreadTimestamp
			if !isChan {
				threadTs = ""
			}
			msg := &botpb.Message{
				Channel:   ev.Channel,
				Timestamp: ev.Timestamp,
				ThreadTs:  threadTs,

				Payload:  ev.Text,
				Nickname: b.nicknameFromID(ev.User),
				//Groups:     ev.Group, // TODO: implement sending a list of the user's groups to the redshirt if it wants to make more complicated authz decisions
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
			log.Printf("Error: %s", ev.Error())

		case *slack.InvalidAuthEvent:
			log.Fatalf("Invalid credentials: %s", ev)

		default:
			// Ignore other events..
			//				fmt.Println("Unhandled event: ", msg.Type)
		}
	}
}

// We want a way to return the slaslack user ID
func (b *SlackBot) nicknameFromID(id string) string {
	if u, ok := b.users.Load(id); ok {
		return u.(slack.User).Name
	}
	return ""
}

func (b *SlackBot) idHasEmail(id, email string) bool {
	if u, ok := b.users.Load(id); ok {
		user := u.(slack.User)
		log.Println("User ID: " + user.ID + " email: " + user.Profile.Email + " name: " + user.Name)
		if user.ID == id && user.Profile.Email == email {
			log.Println("---------------- MATCHED -----------------")
			return true
		}
	}
	return false
}

func (b *SlackBot) idInGroup(id, group string) bool {
	for _, g := range b.groups {
		log.Printf("checking group " + group + " == " + g.Handle)
		if g.Handle == group {
			m, err := b.api.GetUserGroupMembers(g.ID)
			spew.Dump(err)
			spew.Dump(m)
			for _, u := range m {
				log.Printf("ZOMG: %s == %s", u, id)
				if u == id {
					return true
				}
			}
		}

	}
	return false
}
