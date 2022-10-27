package main

import (
	"bufio"
	"fmt"
	"io"
	"log"
	"os"
	"strings"
	"time"
	"regexp"

	"os/exec"

	"github.com/google/uuid"
	"golang.org/x/net/websocket"
)

var trimColorRegex = regexp.MustCompile(`\033\[[0-9]+m`)

const (
	EnterValueMessage = "Enter a value:"
	PlanStart         = "Terraform will perform the following actions:"
	PlanEnd           = "Plan:"
	yes               = "yes\n"
	no                = "no\n"
)

func wrapTerraformApply(c *Config) error {
	cmd := exec.Command(c.Command.TerraformCommandPath, c.args...)
	cmd.Stderr = os.Stderr

	sop, err := cmd.StdoutPipe()
	if err != nil {
		return err
	}

	sip, err := cmd.StdinPipe()
	if err != nil {
		return err
	}

	if err := cmd.Start(); err != nil {
		return err
	}

	r := bufio.NewReader(sop)
	applyChan := make(chan bool)
	var planData string
	var isPlanning bool
	var delimiter byte = ':'

	go func() {
		for {
			out, _ := r.ReadString(delimiter)
			if strings.Contains(out, PlanStart) {
				isPlanning = true
				io.WriteString(os.Stdout, out)
				continue
			} else if strings.Contains(out, PlanEnd) {
				if index := strings.Index(out, PlanEnd); index > 0 {
					planData += out[0:index]
				}
				isPlanning = false
			}
			if strings.Contains(out, "Enter a value:") {
				spl := strings.Split(out, "\n")
				io.WriteString(os.Stdout, strings.Join(spl[0:len(spl)-6], "\n"))
				io.WriteString(os.Stdout, "\n\ntfapprove prevents confirmation input.\nWating for approval...")
				go func() {
					planData = trimColorRegex.ReplaceAllString(planData, "")
					if err := waitForApproval(applyChan, c, planData); err != nil {
						log.Printf("[TFApprove] %s\n", err)
					}
				}()
				delimiter = '\n'
				continue
			}

			if isPlanning {
				planData += out
			}
			io.WriteString(os.Stdout, out)
		}
	}()

	go func() {
		ok := <-applyChan
		if ok {
			io.WriteString(sip, yes)
			return
		}
		io.WriteString(sip, no)
	}()

	return cmd.Wait()
}

func waitForApproval(ac chan bool, c *Config, plan string) error {
	sessionId := uuid.New().String()
	dc, err := websocket.NewConfig(fmt.Sprintf("%s/wait/%s", c.Server.Url, sessionId), c.Server.Url)
	if err != nil {
		ac <- false
		return err
	}
	dc.Header.Add("TFApprove-Api-Key", c.Server.ApiKey)
	conn, err := websocket.DialConfig(dc)
	if err != nil {
		ac <- false
		return fmt.Errorf("Failed to connect step server")
	}
	defer conn.Close()

	if err := websocket.JSON.Send(conn, Handshake{
		Plan:    plan,
		Channel: c.Approve.SlackChannel,
	}); err != nil {
		ac <- false
		return err
	}

	timeout := time.After(time.Duration(c.Approve.WaitTimeout) + time.Minute)
	action := make(chan bool)
	errCh := make(chan error)

	go func() {
		var approvals int
		for {
			var msg Message
			if err := websocket.JSON.Receive(conn, &msg); err != nil {
				errCh <- err
				return
			}
			switch msg.Action {
			case "approve":
				fmt.Fprintf(os.Stdout, "%s approved your plan.", msg.User)
				approvals++
				if approvals >= c.Approve.NeedApprovers {
					fmt.Fprint(os.Stdout, "Continue to apply!\n")
					action <- true
					return
				} else {
					fmt.Fprint(os.Stdout, "\n")
				}
			case "reject":
				fmt.Fprintf(os.Stdout, "%s rejected your plan.\n", msg.User)
				action <- false
				return
			}
		}
	}()

	select {
	case <-timeout:
		log.Println("[TFApprove] Wait timeout, cancel apply")
		ac <- false
		return nil
	case result := <-action:
		ac <- result
		return nil
	case err := <-errCh:
		return err
	}
}
