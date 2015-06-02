package main

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"time"
)

// Server ...
type Server struct {
	Alias    string
	Hostname string
	Notes    string
	Tags     []string
	Hit      int
	port     string
}

func (s Server) ssh() {
	if s.Notes != "" {
		renderNotes(s)
	}

	// TODO: this parsing should probably be done at time of server add.
	s.port = "22"
	if strings.Contains(s.Hostname, ":") {
		p := strings.Split(s.Hostname, ":")
		s.Hostname = p[0]
		s.port = p[1]
	}

	success := s.sshStartProcess()
	for success == false {
		handleWarning(fmt.Errorf("Reconnecting. Press Ctrl+C to abort."))
		time.Sleep(3000 * time.Millisecond)
		success = s.sshStartProcess()
	}
}

func (s Server) sshStartProcess() (success bool) {
	args := []string{"-p " + s.port, s.Hostname, "-v"}

	cmd := exec.Command("ssh", args...)
	cmd.Stdout = os.Stdout
	cmd.Stdin = os.Stdin

	cmdReader, err := cmd.StderrPipe()
	handleError(err)

	// First attempt at monitoring connection status.
	scanner := bufio.NewScanner(cmdReader)
	go func() {
		connected := false
		loggedin := false
		for scanner.Scan() {
			stderr := scanner.Text()
			if !connected && strings.Contains(stderr, "Connection established") {
				connected = true
				handleStatus("Connected.")
			} else if !loggedin && strings.Contains(stderr, "Entering interactive session") {
				loggedin = true
			} else if loggedin && !strings.Contains(stderr, "debug") {
				cmd.Stderr = os.Stderr
			}
		}
	}()

	err = cmd.Start()
	handleError(err)

	err = cmd.Wait()
	handleError(err)

	// 127 is command not found
	// 130 is ctrl+c
	// TODO: is there not a way to get the return code as an int?
	rs := cmd.ProcessState.String()
	if rs == "exit status 127" || rs == "exit status 130" {
		return true
	}
	return cmd.ProcessState.Success()
}
