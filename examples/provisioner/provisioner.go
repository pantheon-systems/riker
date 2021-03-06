package main

import (
	"bufio"
	"bytes"
	"context"
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
	"regexp"
	"strings"

	"github.com/davecgh/go-spew/spew"
	"github.com/pantheon-systems/riker/pkg/botpb"
	"github.com/pantheon-systems/riker/pkg/redshirt"
)

var okMatch = regexp.MustCompile(`OK:`)
var client botpb.RikerClient

func main() {

	// load TLS shit
	tlsFile := os.Getenv("REDSHIRT_TLS_CERT")
	if tlsFile == "" {
		log.Fatal("REDSHIRT_TLS_CERT env var not set")
	}
	caFile := os.Getenv("REDSHIRT_CA_FILE")
	if caFile == "" {
		log.Fatal("REDSHIRT_CA_FILE env var not set")
	}

	// connect to riker
	conn, err := redshirt.NewTLSConnection("localhost:6000", tlsFile, caFile)
	if err != nil {
		log.Fatal(err)
	}
	defer func() {
		err := conn.Close()
		if err != nil {
			log.Panic(err)
		}
	}()

	client = botpb.NewRikerClient(conn)
	cap := &botpb.Capability{
		Name:        "gce-provision",
		Usage:       "Manage provisioning of gce instances",
		Description: "Provision and delete gce instances",
		Auth: &botpb.CommandAuth{
			Users: []string{
				"jesse@getpantheon.com",
				"joe@getpantheon.com",
			},
			//			Groups: []string{"infra"},
		},
	}

	reg, err := client.NewRedShirt(context.Background(), cap)
	if err != nil {
		log.Fatal("Failed creating the redshirt: ", err.Error())
	}

	if reg.CapabilityApplied {
		log.Println("Rejoice we are the first instance to register 'gce-provision'.")
	} else {
		log.Println("Starting up as another 'gce-provision' minion.")
	}

	stream, err := client.CommandStream(context.Background(), reg)
	if err != nil {
		log.Fatal("Error talking to riker: ", err)
	}

	for {
		msg, err := stream.Recv()
		if err == io.EOF {
			continue
		}
		if err != nil {
			log.Fatalf("error reading message from riker: %+v = %v", client, err)
		}

		log.Printf("Got message: %+v\n", msg)
		fields := strings.Fields(msg.Payload)

		// buld up the command and the args from the passed in cmd
		args := fields[1:]
		if fields[0] == "<@" {
			args = fields[2:]
		}

		reply := &botpb.Message{
			Channel:   msg.Channel,
			Timestamp: msg.Timestamp,
			ThreadTs:  msg.Timestamp,
		}

		spew.Dump(reply)

		c := exec.Cmd{
			Path: "./test",
			// Path: "./provision",
			Args: args,
		}

		go runCmd(reply, c)
	}
}

func runCmd(reply *botpb.Message, c exec.Cmd) {

	stdout, _ := c.StdoutPipe()
	stderr, _ := c.StderrPipe()

	stderrCopy := &bytes.Buffer{}
	tee := io.TeeReader(stderr, stderrCopy)
	combined := io.MultiReader(stdout, tee)

	err := c.Start()
	if err != nil {
		log.Println("Failed to start command: ", err.Error())
		reply.Payload = "Failed to start command: ```" + err.Error() + "```"
		sendMsg(reply)
		return
	}

	reply.Payload = "Starting provision with args: ```" + strings.Join(c.Args, " ") + "```"
	sendMsg(reply)

	lines := make([]string, 200)
	// recieve line
	// set timer
	// if we get data before timer expires
	// reset timer
	// if timer expires if buffer >= max msg size, flush
	r := bufio.NewReader(combined)
	for {
		line, _, err := r.ReadLine()
		if err != nil {
			if err == io.EOF {
				reply.Payload = "```" + strings.Join(lines, "") + "```"
				sendMsg(reply)
				break
			}

			reply.Payload = "Error reading output: ```" + err.Error() + "```\n While Processing: ```" + string(line) + "```"
			sendMsg(reply)
			break
		}

		fmt.Print(string(line))
		lines = append(lines, string(line))

		if len(lines) >= 5 {
			reply.Payload = "```" + strings.Join(lines, "") + "```"
			sendMsg(reply)
			lines = lines[:0]
		}
		//		if okMatch.Match(line) {
		//		}
	}

	if err = c.Wait(); err != nil {
		if _, ok := err.(*exec.ExitError); ok {
			// The program has exited with an exit code != 0
			reply.Payload = "Command exit with status code > 0:\n ```" + stderrCopy.String() + "```"
			sendMsg(reply)
			reply.ThreadTs = ""
			reply.Timestamp = ""
			sendMsg(reply)
		}
	}
}

func sendMsg(msg *botpb.Message) {
	resp, err := client.Send(context.Background(), msg)
	if err != nil {
		log.Println("Error sending: ", err)
	}
	log.Println("Sent!!! ", resp)
}
