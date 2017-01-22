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
	RunCmd   string
	Username string
}

func (s Server) ssh() {
	/*if s.Notes != "" {
		renderNotes(s)
	}*/

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
	args := []string{"-p " + s.port, s.Hostname}
	if len(s.RunCmd) > 0 {
		// we have a command to run once we connect, so append it to the args we're sending ssh
		// specify -t to require a tty (just in case, for tmux/screen/etc)
		args = append(args, "-t", "--", s.RunCmd)
	}

	//handleDebug(fmt.Sprintf("running ssh %s", args))
	cmd := exec.Command("ssh", args...)
	cmd.Stderr = os.Stderr
	cmd.Stdout = os.Stdout
	cmd.Stdin = os.Stdin

	err := cmd.Start()
	handleError(err)

	err2 := cmd.Wait()
	handleError(err2)

	// 127 is command not found
	// 130 is ctrl+c
	// TODO: is there not a way to get the return code as an int?
	rs := cmd.ProcessState.String()
	if rs == "exit status 127" || rs == "exit status 130" {
		return true
	}
	return cmd.ProcessState.Success()
}
