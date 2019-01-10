package slackbot

import (
	"crypto/tls"
	"crypto/x509/pkix"
	"fmt"
	"io"
	"log"
	"net"
	"strings"
	"sync"
	"time"

	"golang.org/x/net/context"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/keepalive"
	"google.golang.org/grpc/peer"

	"github.com/Sirupsen/logrus"
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

// ErrorLoadingCert Is returned when some failure to read/parse/load cert is encountered
type ErrorLoadingCert struct {
	m string
}

// Error implements the error interface for this type
func (e ErrorLoadingCert) Error() string {
	return e.m
}

// clientCertKey is the context key where the client mTLS cert (pkix.Name) is stored
type clientCertKey struct{}

// SlackBot is the slack adaptor for riker. It implments the riker.Bot interface for bridging redshirts
// to chat platforms
type SlackBot struct {
	name string
	rtm  *slack.RTM
	api  *slack.Client

	log    *logrus.Logger
	users  sync.Map
	groups []slack.UserGroup

	// redshirts holds state of commands that map to client registrations
	redshirts  map[string]*redshirtRegistration
	channels   map[string]bool
	allowedOUs []string

	bindAddr string
	grpc     *grpc.Server

	*sync.RWMutex
}

// NewRedShirt implements the riker protobuf server
func (b *SlackBot) NewRedShirt(ctx context.Context, cap *botpb.Capability) (*botpb.Registration, error) {
	b.log.Info("Registering client ", cap)
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
		b.log.Debug("Received value")
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
func New(name, bindAddr, botKey, token, tlsFile, caFile string, allowedOUs []string, log *logrus.Logger) (riker.Bot, error) {
	if log == nil {
		log = logrus.New()
	}

	cert, err := certutils.LoadKeyCertFiles(tlsFile, tlsFile)
	if err != nil {
		return nil, ErrorLoadingCert{fmt.Sprintf("Could not load TLS cert '%s': %s", tlsFile, err.Error())}
	}

	caPool, err := certutils.LoadCACertFile(caFile)
	if err != nil {
		return nil, ErrorLoadingCert{fmt.Sprintf("Could not load CA cert '%s': %s", caFile, err.Error())}
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

	debugOption := slack.OptionDebug(log.Level == logrus.DebugLevel)

	b := &SlackBot{
		name: name,
		log:  log,

		bindAddr:   bindAddr,
		allowedOUs: allowedOUs,

		api: slack.New(token, debugOption),
		rtm: slack.New(botKey, debugOption).NewRTM(),

		channels:  make(map[string]bool, 100),
		redshirts: make(map[string]*redshirtRegistration, 10),
	}

	grpcServer := grpc.NewServer(
		grpc.KeepaliveParams(k),
		grpc.StreamInterceptor(grpc_auth.StreamServerInterceptor(b.authOU)),
		grpc.UnaryInterceptor(grpc_auth.UnaryServerInterceptor(b.authOU)),
		grpc.Creds(credentials.NewTLS(tlsConfig)),
	)

	b.grpc = grpcServer
	b.RWMutex = &sync.RWMutex{}

	botpb.RegisterRikerServer(grpcServer, b)
	return b, nil
}

// Run starts the bot
func (b *SlackBot) Run() {
	b.log.Info("starting slack RTM broker")
	// TODO: check nil maybe return errors, and do that in New instead
	go b.rtm.ManageConnection()

	l, err := net.Listen("tcp", b.bindAddr)
	if err != nil {
		b.log.Fatalf("failed to listen: %v", err)
	}
	b.log.Info("Starting GRPC server listening on " + b.bindAddr)
	go func() {
		log.Fatal(b.grpc.Serve(l))
	}()

	b.log.Info("Starting Slack Broker")
	b.startBroker()
}

func (b *SlackBot) isChan(ev *slack.MessageEvent) (bool, error) {
	// we need to detect a direct message from a channel message, and unfortunately slack doens't make that super awesome
	// TODO: might try simplifing to this huristic though might be just as brittal as this code is now
	// https://stackoverflow.com/questions/41111227/how-can-a-slack-bot-detect-a-direct-message-vs-a-message-in-a-channel
	b.Lock()
	defer b.Unlock()

	isChan, ok := b.channels[ev.Channel]
	if !ok {
		_, err := b.rtm.GetChannelInfo(ev.Channel)
		if err != nil && (err.Error() != "channel_not_found" && err.Error() != "method_not_supported_for_channel_type") {
			b.log.Debug("not dm not channel: ", err)
			return false, err
		}

		if err == nil {
			isChan = true
		}

		b.channels[ev.Channel] = isChan
	}
	return isChan, nil
}

func (b *SlackBot) startBroker() {
	var botID string

	for {
		msg := <-b.rtm.IncomingEvents
		switch ev := msg.Data.(type) {
		case *slack.ConnectedEvent:
			b.log.Debugf("Connected to slack %+v", ev)
			var err error
			botID = ev.Info.User.ID

			usersSlice, err := b.rtm.GetUsers()
			if err != nil {
				b.log.Fatal(err)
			}
			for _, u := range usersSlice {
				b.users.Store(u.ID, u)
			}

			b.groups, err = b.api.GetUserGroups()
			if err != nil {
				b.log.Fatal(err)
			}

			b.log.Debugf("Riker has connected to Slack! %+v", ev.Info.User)

		case *slack.TeamJoinEvent:
			b.log.Debugf("User Joined: %+v", ev.User)
			b.users.Store(ev.User.ID, ev.User)

		case *slack.MessageEvent:
			// spew.Dump(ev)
			// ignore messages from other bots or ourself
			// XXX: in order to avoid responding to the bot's own messages, it appears we need to check for an empty ev.User
			//log.Printf("botID: %s, User: %s", ev.BotID, ev.User)
			if ev.BotID != "" || ev.User == botID || ev.User == "" {
				continue
			}

			if ev.Type != "message" {
				continue
			}

			ll := b.log.WithField("uid", ev.User)
			ll.Debugf("Message recieved: %+v", ev)

			// for now strip this out and convert the text message into an array of words for easier parsing
			msgSlice := strings.Split(ev.Text, " ")

			// if not a chan then it is a DM
			isChan, err := b.isChan(ev)
			if err != nil {
				continue
			}
			ll = ll.WithField("isChan", isChan)

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
			ll = ll.WithField("command", msgSlice[1])

			if msgSlice[1] == "help" {
				// TODO: implement help using the registered Capability's and their name / description / usage
				msg := b.rtm.NewOutgoingMessage("no help for you", ev.Channel)
				go b.rtm.SendMessage(msg)
				continue
			}

			// match the message prefix to registered commands
			cmdName := msgSlice[1]
			ll.Debug("checking for command")
			b.RLock()
			rsReg, ok := b.redshirts[cmdName]
			b.RUnlock()
			if !ok {
				msg := b.rtm.NewOutgoingMessage("Sorry, that redshirt has not reported for duty. I can't complete the request", ev.Channel)
				go b.rtm.SendMessage(msg)
				continue
			}

			// Verify this person is allowed to run this command, cause unauthed commands are a violation of starfleet protocols.
			ll = ll.WithField("auth", false)
			ll.Info("Checking user authentication")

			auth := false
			for _, ua := range rsReg.capability.Auth.Users {
				auth = b.idHasEmail(ev.User, ua)
				if auth {
					ll = ll.WithField("auth", true)
					ll = ll.WithField("auth-email", ua)
					break
				}
			}

			for _, ga := range rsReg.capability.Auth.Groups {
				auth = auth || b.idInGroup(ev.User, ga)
				if auth {
					ll = ll.WithField("auth", true)
					ll = ll.WithField("auth-group", ga)
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

			ll.Debug("sending Command", msg)
			select {
			case rsReg.queue <- msg:
			default:
				ll.Warn("Couldn't send to internal slack message queue.")
				msg := b.rtm.NewOutgoingMessage("Sorry, redshirt supply is low. Couldn't complete your request.", ev.Channel)
				go b.rtm.SendMessage(msg)
				break
			}

		case *slack.RTMError:
			b.log.Warn("Slack RTM error: ", ev.Msg)

		case *slack.InvalidAuthEvent:
			b.log.Warnf("Invalid credentials: %s", ev)

		case *slack.ConnectionErrorEvent:
			b.log.WithError(ev.ErrorObj).Warn("Error connecting to slack")

		case *slack.AckErrorEvent:
			b.log.Warnf("connection error: %+v", msg)
			// TODO: exp backoff ?
			time.Sleep(10 * time.Second)

		default:
			// Ignore other events..
			b.log.Debug("Unhandled slack event: ", msg.Type)
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
		b.log.Debug("User ID: " + user.ID + " email: " + user.Profile.Email + " name: " + user.Name)
		if user.ID == id && user.Profile.Email == email {
			b.log.Debug("---------------- MATCHED -----------------")
			return true
		}
	}
	return false
}

func (b *SlackBot) idInGroup(id, group string) bool {
	for _, g := range b.groups {
		b.log.Debug("checking group " + group + " == " + g.Handle)
		if g.Handle == group {
			m, err := b.api.GetUserGroupMembers(g.ID)
			if err != nil {
				b.log.Warnf("failed GetUserGroupMembers(%s): %s", g.ID, err)
			}
			for _, u := range m {
				b.log.Debugf("Match?: %s == %s", u, id)
				if u == id {
					return true
				}
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

// AuthOU extracts the client mTLS cert from a context and verifies the client cert OU is in the list of
// allowedOUs. It returns an error if the client is not permitted.
func (b *SlackBot) authOU(ctx context.Context) (context.Context, error) {
	b.log.Debugf("allowed OUs %s", b.allowedOUs)
	tlsState, err := tlsConnStateFromContext(ctx)
	if err != nil {
		return nil, err
	}

	clientCert, err := getCertificateSubject(tlsState)
	if err != nil {
		return nil, err
	}
	b.log.Debugf("authOU: client CN=%s, OU=%s", clientCert.CommonName, clientCert.OrganizationalUnit)

	if b.intersectArrays(clientCert.OrganizationalUnit, b.allowedOUs) {
		return context.WithValue(ctx, clientCertKey{}, clientCert), nil
	}
	b.log.Debugf("authOU: client cert failed authentication. CN=%s, OU=%s", clientCert.CommonName, clientCert.OrganizationalUnit)
	return nil, grpc.Errorf(codes.PermissionDenied, "client cert OU '%s' is not allowed.", clientCert.OrganizationalUnit)
}

// intersectArrays returns true when there is at least one element in common between the two arrays
func (b *SlackBot) intersectArrays(orig, tgt []string) bool {
	for _, i := range orig {
		for _, x := range tgt {
			x = strings.TrimSpace(x)
			b.log.Debug(i + "==" + x)
			if i == x {
				return true
			}
		}
	}
	return false
}
