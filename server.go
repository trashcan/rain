package main

import (
	"fmt"
	"os"
	"os/exec"
	"time"
)

// Server ...
type Server struct {
	Alias    string
	Hostname string
	Notes    string
	Tags     []string
	Hit      int
}

func (s Server) ssh() {
	if s.Notes != "" {
		renderNotes(s)
	}

	success := s.sshStartProcess(s.Hostname)
	for success == false {
		fmt.Println("Unusual termination, reconnecting (or the last ran command did not return 0).")
		time.Sleep(1000 * time.Millisecond)
		success = s.sshStartProcess(s.Hostname)
	}
}

func (s Server) sshStartProcess(hostname string) (success bool) {
	cwd, err := os.Getwd()
	handleError(err)

	pa := os.ProcAttr{
		Files: []*os.File{os.Stdin, os.Stdout, os.Stderr},
		Dir:   cwd,
	}

	ssh, err := exec.LookPath("ssh")
	handleError(err)

	proc, err := os.StartProcess(ssh, []string{"-v", s.Hostname}, &pa)
	handleError(err)

	state, err := proc.Wait()
	handleError(err)

	// 127 is command not found
	// 130 is ctrl+c
	// TODO: is there not a way to get the return code as an int?
	if state.String() == "exit status 127" || state.String() == "exit status 130" {
		return true
	} else if !state.Success() {
		fmt.Printf("WARNING: %s. If this should not cause a reconnect email plathem@gmail.com.\n", state.String())
	}

	return state.Success()
}
