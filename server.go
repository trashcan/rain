package main

import (
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
		fmt.Println("Unusual termination, reconnecting (or the last ran command did not return 0).")
		time.Sleep(3000 * time.Millisecond)
		success = s.sshStartProcess()
	}
}

func (s Server) sshStartProcess() (success bool) {
	cwd, err := os.Getwd()
	handleError(err)

	pa := os.ProcAttr{
		Files: []*os.File{os.Stdin, os.Stdout, os.Stderr},
		Dir:   cwd,
	}

	ssh, err := exec.LookPath("ssh")
	handleError(err)

	args := []string{"", "-p " + s.port, s.Hostname}
	proc, err := os.StartProcess(ssh, args, &pa)
	handleError(err)

	state, err := proc.Wait()
	handleError(err)

	// 127 is command not found
	// 130 is ctrl+c
	// TODO: is there not a way to get the return code as an int?
	if state.String() == "exit status 127" || state.String() == "exit status 130" {
		return true
	} else if !state.Success() {
		fmt.Printf("WARNING: %s. \n", state.String())
	}

	return state.Success()
}
