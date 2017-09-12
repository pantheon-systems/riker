package terminalbot

import (
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/jroimartin/gocui"
	"github.com/pantheon-systems/riker/pkg/botpb"
)

const (
	titleView  = "titleView"
	inputView  = "inputView"
	outputView = "outputView"
	logView    = "logView"

	botString = "@riker"
)

func (t *TerminalBot) viewTitle(lMaxX int, lMaxY int) error {
	v, err := t.gui.SetView(titleView, -1, -1, lMaxX, 1)
	if err != nil {
		if err != gocui.ErrUnknownView {
			return err
		}
	}

	// Settings
	v.Frame = false
	v.BgColor = gocui.ColorGreen
	v.FgColor = gocui.ColorBlack

	// Content
	fmt.Fprintln(v, "-={ riker }=- [ctrl-c] to exit")

	return nil
}

func (t *TerminalBot) viewLog(lMaxX int, lMaxY int) error {
	v, err := t.gui.SetView(logView, 0, 1, lMaxX-1, 9)
	if err != nil {
		if err != gocui.ErrUnknownView {
			return err
		}
	}

	v.Title = "Riker Log"
	v.Frame = true
	v.Editable = false
	v.Autoscroll = true
	v.Wrap = true

	// update the logger
	go func() {
		for {
			select {
			case <-time.After(500 * time.Millisecond):
				t.gui.Update(func(g *gocui.Gui) error { return nil })
			}
		}
	}()

	return nil
}

func (t *TerminalBot) viewOutput(lMaxX int, lMaxY int) error {
	v, err := t.gui.SetView(outputView, 0, 10, lMaxX-1, lMaxY-4)
	if err != nil {
		if err != gocui.ErrUnknownView {
			return err
		}
	}
	v.Title = "Chat output"
	v.Frame = true
	v.Editable = false
	v.Autoscroll = true
	v.Wrap = true
	return nil
}

func (t *TerminalBot) viewInput(lMaxX int, lMaxY int) error {
	v, err := t.gui.SetView(inputView, 0, lMaxY-3, lMaxX-1, lMaxY-1)
	if err != nil {
		if err != gocui.ErrUnknownView {
			return err
		}
	}
	v.Frame = true
	v.Editable = true

	v.Autoscroll = true
	v.Editable = true
	v.Wrap = false

	v.Title = "Input use @riker to send commands"
	return nil
}

func (t *TerminalBot) layoutUI(*gocui.Gui) error {
	maxX, maxY := t.gui.Size()

	t.viewTitle(maxX, maxY)
	// TODO: make log an overlay ? or zoomable / scrollable
	t.viewLog(maxX, maxY)
	t.viewOutput(maxX, maxY)
	t.viewInput(maxX, maxY)

	t.gui.SetCurrentView(inputView)

	return nil
}

func quit(g *gocui.Gui, v *gocui.View) error {
	return gocui.ErrQuit
}

func (t *TerminalBot) echoInput(g *gocui.Gui, iv *gocui.View) error {
	// We want to read the view’s buffer from the beginning.
	iv.Rewind()
	if iv.Buffer() != "" {
		t.writeView(outputView, iv.Buffer())
	}
	iv.Clear()
	iv.SetCursor(0, 0)
	return nil
}

func (t *TerminalBot) startLoop() error {
	defer t.gui.Close()
	if err := t.gui.SetKeybinding(inputView, gocui.KeyEnter, gocui.ModNone, t.echoInput); err != nil {
		return err
	}

	if err := t.gui.MainLoop(); err != nil && err != gocui.ErrQuit {
		return err
	}
	return nil
}

// writeView writes string to view
func (t *TerminalBot) writeView(name, text string) {
	ts := time.Now()
	v, _ := t.gui.View(name)
	fmt.Fprint(v, ts.Format("15:04:05")+"> "+text)

	if name != outputView {
		return
	}

	// for now strip this out and convert the text message into an array of words for easier parsing
	msgSlice := strings.Split(text, " ")
	botString := "@riker"
	if msgSlice[0] != botString {
		log.Println("riker not being addressed")
		return
	}

	// ignore when someone addresses us without a command
	if len(msgSlice) < 2 {
		log.Println("command is too short, skipping")
		return
	}

	if msgSlice[1] == "help" {
		// TODO: implement help using the registered Capability's and their name / description / usage
		t.writeView(outputView, "No help")
		return
	}

	// match the message prefix to registered commands
	cmdName := msgSlice[1]
	t.RLock()
	rsReg, ok := t.redshirts[cmdName]
	t.RUnlock()
	if !ok {
		t.writeView(outputView, "Sorry, that redshirt has not reported for duty. I can't complete the request")
		return
	}

	msg := &botpb.Message{
		Channel:   "term",
		Timestamp: ts.String(),
		ThreadTs:  ts.String(),

		Payload:  text,
		Nickname: "user",
	}

	log.Printf("sending msg to redshirt: %v\n", msg)

	go func() {
		select {
		case rsReg.queue <- msg:
		default:
			log.Println("Redshirt send queue full")
			break
		}
	}()
}
