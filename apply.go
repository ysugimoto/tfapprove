package main

import (
	"bufio"
	"fmt"
	"io"
	"log"
	"os"
	"regexp"
	"strings"
	"time"

	"os/exec"

	"github.com/google/uuid"
	"golang.org/x/net/websocket"
)

// Removale shell escape sequence
var trimColorRegex = regexp.MustCompile(`\033\[[0-9]+m`)

// Following string is spcific point of terraform apply command output.
const (
	// EnterValueMessage is trap point to wait for inputting "yes" of "no" from terraform
	EnterValueMessage = "Enter a value:"
	// PlanStart is trap point to start collecting plan result
	PlanStart = "Terraform will perform the following actions:"
	// PlanEnd is trap point to end collecting plan result
	PlanEnd = "Plan:"
	// yes is shortcut command to input "yes"
	yes = "yes\n"
	// no is shortcut command to input "no"
	no = "no\n"
)

// Wrap "terraform apply" command function
// Pipe stdout, stderr, stdin of terraform apply process.
func wrapTerraformApply(c *Config) error {
	cmd := exec.Command(c.Command.TerraformCommandPath, c.args...)
	cmd.Stderr = os.Stderr

	// Capture terraform output
	sop, err := cmd.StdoutPipe()
	if err != nil {
		return err
	}

	// Wrap stdin to trap user input.
	// tfapprove supresses that the user input "yes" or "no" directly.
	sip, err := cmd.StdinPipe()
	if err != nil {
		return err
	}

	if err := cmd.Start(); err != nil {
		return err
	}

	r := bufio.NewReader(sop)
	applyChan := make(chan bool, 1)
	var planData string
	var isPlanning bool
	var delimiter byte = ':'

	// Collect plan data and wait for approval
	go func() {
		for {
			out, _ := r.ReadString(delimiter)
			if strings.Contains(out, PlanStart) {
				isPlanning = true
				_, _ = io.WriteString(os.Stdout, out)
				continue
			} else if strings.Contains(out, PlanEnd) {
				if index := strings.Index(out, PlanEnd); index > 0 {
					planData += out[0:index]
				}
				isPlanning = false
			}
			if strings.Contains(out, EnterValueMessage) {
				spl := strings.Split(out, "\n")
				_, _ = io.WriteString(os.Stdout, strings.Join(spl[0:len(spl)-6], "\n"))
				_, _ = io.WriteString(os.Stdout, "\n\ntfapprove prevents confirmation input.\nWating for approval...\n")
				planData = trimColorRegex.ReplaceAllString(planData, "")
				action, err := waitForApproval(c, strings.TrimSpace(planData))
				if err != nil {
					if err == io.EOF {
						log.Printf("[TFApprove] %s\n", "Connection Closed")
					} else {
						log.Printf("[TFApprove] %s\n", err)
					}
					applyChan <- false
				}
				delimiter = '\n'
				applyChan <- action
				continue
			}

			if isPlanning {
				planData += out
			}
			_, _ = io.WriteString(os.Stdout, out)
		}
	}()

	// Wait approval result and pass "yes" or "no" to the terraform process
	go func() {
		ok := <-applyChan
		if ok {
			log.Println("Apply this plan")
			_, _ = io.WriteString(sip, yes)
			sip.Close()
		} else {
			log.Println("Cancel this plan")
			_, _ = io.WriteString(sip, no)
			_ = cmd.Process.Kill()
			os.Exit(1)
		}
	}()

	return cmd.Wait()
}

// Connect to aggregate server and check the member approved or rejected.
func waitForApproval(c *Config, plan string) (bool, error) {
	sessionId := uuid.New().String()
	dc, err := websocket.NewConfig(fmt.Sprintf("%s/%s", server, sessionId), server)
	if err != nil {
		return false, err
	}
	dc.Header.Add("TFApprove-Api-Key", c.Server.ApiKey)
	conn, err := websocket.DialConfig(dc)
	if err != nil {
		return false, fmt.Errorf("Failed to connect aggregate server")
	}
	defer conn.Close()

	// Send handshake
	if err := websocket.JSON.Send(conn, Handshake{
		Plan:    plan,
		Channel: c.Approve.SlackChannel,
	}); err != nil {
		return false, err
	}

	timeout := time.After(time.Duration(c.Approve.WaitTimeout) * time.Minute)
	ticker := time.NewTicker(10 * time.Second)
	action := make(chan bool)
	errCh := make(chan error)
	defer ticker.Stop()

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
					fmt.Fprint(os.Stdout, " Continue to apply!\n")
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

	for {
		select {
		case <-timeout:
			log.Println("[TFApprove] Wait timeout, cancel apply")
			_ = websocket.JSON.Send(conn, Action{
				Type: "timeout",
			})
			return false, nil
		case result := <-action:
			_ = websocket.JSON.Send(conn, Action{
				Type: "done",
			})
			return result, nil
		case err := <-errCh:
			return false, err
		case <-ticker.C:
			if err := websocket.JSON.Send(conn, Action{
				Type: "ping",
			}); err != nil {
				return false, err
			}
		}
	}
}
