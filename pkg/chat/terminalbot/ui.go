package terminalbot

import (
	"fmt"
	"time"

	"github.com/jroimartin/gocui"
)

const (
	titleView  = "titleView"
	inputView  = "inputView"
	outputView = "outputView"
)

func (t *TerminalBot) viewTitle(lMaxX int, lMaxY int) error {
	g := t.gui

	v, err := g.SetView(titleView, -1, -1, lMaxX, 1)
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
	fmt.Fprintln(v, "-={ riker terminal mode }=- [ctrl-c] to exit")

	return nil
}

func (t *TerminalBot) viewOutput(lMaxX int, lMaxY int) error {
	g := t.gui

	v, err := g.SetView(outputView, 0, 1, lMaxX-1, lMaxY-3)
	if err != nil {
		if err != gocui.ErrUnknownView {
			return err
		}
	}
	v.Frame = true
	v.Editable = false
	v.Autoscroll = true
	v.Wrap = true
	return nil
}

func (t *TerminalBot) viewInput(lMaxX int, lMaxY int) error {
	g := t.gui
	v, err := g.SetView(inputView, 0, lMaxY-3, lMaxX-1, lMaxY-1)
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
	return nil
}

func (t *TerminalBot) layoutUI() error {
	g := t.gui
	maxX, maxY := g.Size()

	t.viewTitle(maxX, maxY)
	t.viewOutput(maxX, maxY)
	t.viewInput(maxX, maxY)

	if _, err := g.SetCurrentView(inputView); err != nil {
		return nil
	}

	return err
}

func quit(g *gocui.Gui, v *gocui.View) error {
	return gocui.ErrQuit
}

func echoInput(g *gocui.Gui, iv *gocui.View) error {
	// We want to read the view’s buffer from the beginning.
	iv.Rewind()
	if iv.Buffer() != "" {
		writeView(g, outputView, iv.Buffer())
	}
	iv.Clear()
	iv.SetCursor(0, 0)
	return nil
}

func (t *TerminalBot) startLoop() error {
	defer t.gui.Close()
	err := g.SetKeybinding(inputView, gocui.KeyEnter, gocui.ModNone, echoInput)

	if err := t.gui.MainLoop(); err != nil && err != gocui.ErrQuit {
		return err
	}
	return nil
}

// writeView writes string to view
func writeView(g *gocui.Gui, name, text string) {
	t := time.Now()
	v, _ := g.View(name)
	fmt.Fprint(v, t.Format("15:04:05")+"> "+text)

}
