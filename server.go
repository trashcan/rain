package main

import (
	"encoding/json"
	"fmt"
	"os"
	"time"

	"github.com/boltdb/bolt"
)

// Server ...
type Server struct {
	Alias    string
	Hostname string
	Notes    string
	Tags     []string
	//TODO: Increment this. Also sort search results by it maybe.
	hit int
}

func (s Server) ssh() {
	if s.Notes != "" {
		fmt.Printf(s.Notes)
	}

	//s.hit++
	//s.save()

	success := s.sshStartProcess(s.Hostname)
	for success == false {
		fmt.Println("Unusual termination, reconnecting (or the last ran command did not return 0).")
		time.Sleep(1000 * time.Millisecond)
		success = s.sshStartProcess(s.Hostname)
	}

}

func (s Server) save() (err error) {
	db := getDB()
	db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte("servers"))
		if err != nil {
			panic(err)
		}

		encoded, err := json.Marshal(s)
		if err != nil {
			panic(err)
		}

		err = b.Put([]byte(s.Alias), encoded)
		return err
	})
	return
}

func (s Server) sshStartProcess(hostname string) (success bool) {
	cwd, err := os.Getwd()
	if err != nil {
		panic(err)
	}

	pa := os.ProcAttr{
		Files: []*os.File{os.Stdin, os.Stdout, os.Stderr},
		Dir:   cwd,
	}

	proc, err := os.StartProcess("/usr/bin/ssh", []string{"-v", s.Hostname}, &pa)
	if err != nil {
		panic(err)
	}

	state, err := proc.Wait()
	if err != nil {
		panic(err)
	}

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
