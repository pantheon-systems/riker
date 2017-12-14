package terminalbot

import (
	"fmt"
	"io"
	"log"
	"net"
	"sync"
	"time"

	"github.com/jroimartin/gocui"
	"github.com/pantheon-systems/riker/pkg/botpb"
	"github.com/pantheon-systems/riker/pkg/riker"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
	"google.golang.org/grpc/keepalive"
)

// holds info on a connected client (redshirt) so we can send it data
type redshirtRegistration struct {
	queue      chan *botpb.Message
	capability *botpb.Capability
}

// TerminalBot Is the terminal running bot
type TerminalBot struct {
	*sync.RWMutex

	listenAddr string
	redshirts  map[string]*redshirtRegistration

	grpc *grpc.Server
	gui  *gocui.Gui

	nickname string
	groups   []string
}

func New(addr string, nickname string, groups []string) *TerminalBot {
	k := keepalive.ServerParameters{
		Time:    3 * time.Second,
		Timeout: 15 * time.Second,
	}
	grpcServer := grpc.NewServer(grpc.KeepaliveParams(k))

	g, err := gocui.NewGui(gocui.OutputNormal)
	if err != nil {
		log.Panic(err)
	}

	t := TerminalBot{
		listenAddr: addr,
		gui:        g,
		grpc:       grpcServer,
		nickname:   nickname,
		groups:     groups,
		redshirts:  make(map[string]*redshirtRegistration, 10),
	}
	t.RWMutex = &sync.RWMutex{}

	g.SetManagerFunc(t.layoutUI)
	g.Highlight = true
	g.SelFgColor = gocui.ColorGreen

	if err := g.SetKeybinding("", gocui.KeyCtrlC, gocui.ModNone, quit); err != nil {
		log.Panicln(err)
	}

	botpb.RegisterRikerServer(grpcServer, t)
	return &t
}

// ChatSend implements the chat side of sending a message in riker for the TerminalBot
func (t *TerminalBot) ChatSend(msg, channel string) {
	log.Println("sending mssg")
	fmt.Printf("%s> %s\n", channel, msg)
}

// NewRedShirt registers this shirt with the bot
func (t TerminalBot) NewRedShirt(ctx context.Context, cap *botpb.Capability) (*botpb.Registration, error) {
	log.Println("Registering client ", cap)
	t.Lock()
	defer t.Unlock()
	resp := &botpb.Registration{
		Name:              cap.Name,
		CapabilityApplied: true,
	}

	// Check for registration. If it already exists apply the registrations capabilities to it
	if r, ok := t.redshirts[cap.Name]; ok {
		log.Println("alredy registered ", cap.Name)

		r.capability = cap
		return resp, nil
	}

	// New registration so we should cary on
	reg := redshirtRegistration{}
	reg.queue = make(chan *botpb.Message, riker.RedshirtBacklogSize)
	reg.capability = cap
	t.redshirts[cap.Name] = &reg

	return resp, nil
}

// SendStream is the call a client makes to send a stream message back to riker
func (t TerminalBot) SendStream(stream botpb.Riker_SendStreamServer) error {
	log.Println("SendStream established")
	for {
		in, err := stream.Recv()
		if err == io.EOF {
			return nil
		}
		if err != nil {
			return err
		}
		t.writeView(outputView, in.Channel+":: "+in.Payload)
	}
}

// CommandStream is the call a client makes to setup a Push stream from riker -> client
func (t TerminalBot) CommandStream(reg *botpb.Registration, stream botpb.Riker_CommandStreamServer) error {
	log.Println("CommandStream established")
	t.RLock()
	rs, ok := t.redshirts[reg.Name]
	t.RUnlock()
	if !ok {
		return riker.ErrorNotRegistered
	}

	for m := range rs.queue {
		err := stream.Send(m)
		if err != nil {
			select {
			case rs.queue <- m:
			default:
				go t.ChatSend("Communicator malfunction while talking to redshirt.", m.Channel)
			}
			break
		}
	}
	return nil
}

// NextCommand is the call a client makes to pull the next command from rikers command buffer
func (t TerminalBot) NextCommand(ctx context.Context, reg *botpb.Registration) (*botpb.Message, error) {
	log.Println("CommandStream established")

	t.RLock()
	rs, ok := t.redshirts[reg.Name]
	t.RUnlock()
	if !ok {
		return nil, riker.ErrorNotRegistered
	}

	select {
	case m := <-rs.queue:
		return m, nil
	case <-ctx.Done():
		return nil, ctx.Err()
	}

	return nil, nil
}

// Send is the call a client makes to send a message back to riker
func (t TerminalBot) Send(ctx context.Context, msg *botpb.Message) (*botpb.SendResponse, error) {
	log.Println("Client called send")
	t.writeView(outputView, "rs::"+msg.Channel+":: "+msg.Payload)
	return &botpb.SendResponse{Ok: true}, nil
}

func (t *TerminalBot) Run() {
	log.Println("Starting terminal event loop")
	go func() {
		log.Fatal(t.startLoop())
	}()

	log.Println("starting logger")
	for i := 0; i <= 5; i++ {
		if i == 5 {
			log.Fatal("couldn't get logger")
		}
		v, err := t.gui.View(logView)
		if err != nil {
			log.Println("couldn't get logview: ", err)
			time.Sleep(1 * time.Second)
			continue
		}
		log.SetOutput(v)
		log.SetFlags(log.Ltime)
		break
	}

	log.Println("started logger")
	log.Println("Starting terminal mode server on ", t.listenAddr)
	l, err := net.Listen("tcp", t.listenAddr)
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}
	go func() {
		log.Fatal(t.grpc.Serve(l))
	}()

	select {}
}
