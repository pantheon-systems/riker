package riker

import (
	"crypto/tls"
	"crypto/x509/pkix"
	"io"
	"log"
	"net"
	"strings"
	"sync"

	"golang.org/x/net/context"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/peer"

	"github.com/davecgh/go-spew/spew"
	grpc_auth "github.com/grpc-ecosystem/go-grpc-middleware/auth"
	"github.com/nlopes/slack"
	"github.com/pantheon-systems/go-certauth/certutils"
	"github.com/pantheon-systems/riker/pkg/botpb"
)

// TODO: move this into config infrastructure. client certs must contain one of these OUs to connect
var allowedOUs = []string{
	"riker-redshirt",
	"riker-server",
	"titan",
}

// clientCertKey is the context key where the client mTLS cert (pkix.Name) is stored
type clientCertKey struct{}

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
	name string
	rtm  *slack.RTM
	api  *slack.Client

	users  sync.Map
	groups []slack.UserGroup

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

// NewRedShirt implements the riker protobuf server
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
				msg := b.rtm.NewOutgoingMessage("Communicator malfunction while talking to redshirt.", m.Channel)
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
	m.ThreadTimestamp = msg.ThreadTs
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
		msg.ThreadTimestamp = in.ThreadTs
		b.rtm.SendMessage(msg)
	}
}

// intersectArrays returns true when there is at least one element in common between the two arrays
func intersectArrays(orig, tgt []string) bool {
	for _, i := range orig {
		for _, x := range tgt {
			if i == x {
				return true
			}
		}
	}
	return false
}

// tlsConnStateFromContext extracts the client (peer) tls connectionState from a context
func tlsConnStateFromContext(ctx context.Context) (*tls.ConnectionState, error) {
	peer, ok := peer.FromContext(ctx)
	if !ok {
		return nil, grpc.Errorf(codes.PermissionDenied, "Permission denied: no peer info")
	}
	tlsInfo, ok := peer.AuthInfo.(credentials.TLSInfo)
	if !ok {
		return nil, grpc.Errorf(codes.PermissionDenied, "Permission denied: peer didn't not present valid peer certificate")
	}
	return &tlsInfo.State, nil
}

// getCertificateSubject extracts the subject from a verified client certificate
func getCertificateSubject(tlsState *tls.ConnectionState) (pkix.Name, error) {
	if tlsState == nil {
		return pkix.Name{}, grpc.Errorf(codes.PermissionDenied, "request is not using TLS")
	}
	if len(tlsState.PeerCertificates) == 0 {
		return pkix.Name{}, grpc.Errorf(codes.PermissionDenied, "no client certificates in request")
	}
	if len(tlsState.VerifiedChains) == 0 {
		return pkix.Name{}, grpc.Errorf(codes.PermissionDenied, "no verified chains for remote certificate")
	}

	return tlsState.VerifiedChains[0][0].Subject, nil
}

// authOU extracts the client mTLS cert from a context and verifies the client cert OU is in the list of
// allowedOUs. It returns an error if the client is not permitted.
func authOU(ctx context.Context) (context.Context, error) {
	tlsState, err := tlsConnStateFromContext(ctx)
	if err != nil {
		return nil, err
	}

	clientCert, err := getCertificateSubject(tlsState)
	if err != nil {
		return nil, err
	}
	log.Printf("authOU: client CN=%s, OU=%s", clientCert.CommonName, clientCert.OrganizationalUnit)

	if intersectArrays(clientCert.OrganizationalUnit, allowedOUs) {
		return context.WithValue(ctx, clientCertKey{}, clientCert), nil
	}
	log.Printf("authOU: client cert failed authentication. CN=%s, OU=%s", clientCert.CommonName, clientCert.OrganizationalUnit)
	return nil, grpc.Errorf(codes.PermissionDenied, "client cert OU '%s' is not allowed", clientCert.OrganizationalUnit)
}

// New is the constroctor for a bot
func New(botKey, token, tlsFile, caFile string) *Bot {
	cert, err := certutils.LoadKeyCertFiles(tlsFile, tlsFile)
	if err != nil {
		log.Fatalf("Could not load TLS cert '%s': %s", tlsFile, err.Error())
	}
	caPool, err := certutils.LoadCACertFile(caFile)
	if err != nil {
		log.Fatalf("Could not load CA cert '%s': %s", caFile, err.Error())
	}
	tlsConfig := certutils.NewTLSConfig(certutils.TLSConfigModern)
	tlsConfig.ClientCAs = caPool
	tlsConfig.ClientAuth = tls.RequireAndVerifyClientCert
	tlsConfig.Certificates = []tls.Certificate{cert}

	grpcServer := grpc.NewServer(
		grpc.StreamInterceptor(grpc_auth.StreamServerInterceptor(authOU)),
		grpc.UnaryInterceptor(grpc_auth.UnaryServerInterceptor(authOU)),
		grpc.Creds(credentials.NewTLS(tlsConfig)),
	)

	b := &Bot{
		rtm:       slack.New(botKey).NewRTM(),
		api:       slack.New(token),
		name:      "riker",
		grpc:      grpcServer,
		redshirts: make(map[string]*redShirtRegistration, 10),
		channels:  make(map[string]bool, 100),
	}
	b.RWMutex = &sync.RWMutex{}
	//b.rtm.SetDebug(true)

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
	go func() {
		log.Fatal(b.grpc.Serve(l))
	}()
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

			log.Printf("Bot has connected! %+v", ev.Info.User)

		case *slack.TeamJoinEvent:
			log.Printf("User Joined: %+v", ev.User)
			b.users.Store(ev.User.ID, ev.User)

		case *slack.MessageEvent:
			//	spew.Dump(ev)
			log.Printf("Message recieved: %+v", ev)
			// ignore messages from other bots or ourself
			// XXX: in order to avoid responding to the bot's own messages, it appears we need to check for an empty ev.User
			//log.Printf("botID: %s, User: %s", ev.BotID, ev.User)
			if ev.BotID != "" || ev.User == botID || ev.User == "" {
				continue
			}

			if ev.Type != "message" {
				continue
			}

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
				msg := b.rtm.NewOutgoingMessage("Sorry, that redshirt has not repoorted for duty. I can't complete the request", ev.Channel)
				go b.rtm.SendMessage(msg)
				continue
			}

			// Verify this person is allowed to run this command, cause unauthed commands are a violation of starfleet protocols.
			log.Println("Checking user authentication", ev.User)

			auth := false
			for _, ua := range rsReg.cap.Auth.Users {
				auth = b.idHasEmail(ev.User, ua)
				if auth {
					break
				}
			}

			for _, ga := range rsReg.cap.Auth.Groups {
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
				//Groups:     ev.Group, // TODO: implement
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
func (b *Bot) nicknameFromID(id string) string {
	if u, ok := b.users.Load(id); ok {
		return u.(slack.User).Name
	}
	return ""
}

func (b *Bot) idHasEmail(id, email string) bool {
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

func (b *Bot) idInGroup(id, group string) bool {
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
