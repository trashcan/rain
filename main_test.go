package main

import (
	"fmt"
	"testing"
)

func TestDB(t *testing.T) {
	s := Server{Alias: "myalias", Hostname: "myhostname", Notes: "mynotes"}
	dbw := DBWrapper{}

	fmt.Println("Testing adding a server.")
	err := dbw.AddServer(s)
	if err != nil {
		t.Errorf("Failed to add a server to DB. err: %s\n", err.Error())
	}

	fmt.Println("Testing retrieving a server.")
	r, err := dbw.GetServer(s.Alias)
	if err != nil {
		t.Errorf("Failed to get server from DB. err: %s\n", err.Error())
	}
	if !compareServers(s, r) {
		t.Errorf("Retrieved wrong content from DB.\n")
	}

	fmt.Println("Testing searching for a server.")
	as, err := dbw.ServerSearch(s.Alias)
	if err != nil {
		t.Errorf("Failed to get server from DB. err: %s\n", err.Error())
	}
	if len(as) != 1 {
		t.Errorf("Incorrect number of search results for alias search: %d\n", len(as))
	}
	if !compareServers(s, as[0]) {
		t.Errorf("Retrieved wrong content from DB searching by alias.\n")
	}

	fmt.Println("Testing updating a server.")
	updated := Server{Alias: s.Alias, Hostname: "updated-hostname", Notes: "update-notes"}
	err = dbw.UpdateServer(updated)
	if err != nil {
		t.Errorf("Failed to update server. err: %s\n", err.Error())
	}
	updated2, err := dbw.GetServer(s.Alias)
	if err != nil {
		t.Errorf("Failed to get updated server.\n")
	}
	if !compareServers(updated, updated2) {
		t.Errorf("Updated server did not retrieve with the same values.\n")
	}

	fmt.Println("Testing deleting a server.")
	err = dbw.DeleteServer(s.Alias)
	if err != nil {
		t.Errorf("Failed to remove a server from DB. err: %s\n", err.Error())
	}
}

func compareServers(a, b Server) bool {
	return a.Alias == b.Alias && a.Hostname == b.Hostname && a.Notes == b.Notes
}
