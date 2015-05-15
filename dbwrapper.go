package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"os/user"
	"strings"
	"time"

	"github.com/boltdb/bolt"
	"github.com/ttacon/chalk"
)

// newBoltDB ...
func newBoltDB() (db *bolt.DB) {
	usr, err := user.Current()
	handleError(err)

	dbpath := usr.HomeDir + "/.rain"
	err = os.MkdirAll(dbpath, 0755)
	handleError(err)

	fullpath := dbpath + "/bolt.db"
	db, err = bolt.Open(fullpath, 0600, &bolt.Options{Timeout: 1 * time.Second})
	handleError(err)

	err = db.Update(func(tx *bolt.Tx) error {
		_, err := tx.CreateBucketIfNotExists([]byte("servers"))
		handleError(err)
		return err
	})
	handleError(err)

	return db
}

// DBWrapper ...
type DBWrapper struct {
	db *bolt.DB
}

//NewDBWrapper ..
func NewDBWrapper() (dbw DBWrapper) {
	bdb := newBoltDB()
	dbw = DBWrapper{db: bdb}
	return
}

// DeleteServer ...
func (dbw DBWrapper) DeleteServer(alias string) (err error) {
	err = dbw.db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte("servers"))
		//TODO: check if key exists first
		err := b.Delete([]byte(alias))
		return err
	})
	handleError(err)
	return
}

// AddServer ...
func (dbw DBWrapper) AddServer(s Server) (err error) {
	err = dbw.db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte("servers"))
		handleError(err)

		encoded, err := json.Marshal(s)
		handleError(err)

		err = b.Put([]byte(s.Alias), encoded)
		return err
	})
	handleError(err)
	return
}

// ServerSearch ...
func (dbw DBWrapper) ServerSearch(search string) (servers []Server, err error) {
	err = dbw.db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte("servers"))
		cursor := b.Cursor()

		for alias, host := cursor.First(); alias != nil; alias, host = cursor.Next() {
			var s Server
			err := json.Unmarshal(host, &s)
			handleError(err)

			// TODO: move color into cmdSSH. breaks other searches
			if strings.Contains(string(alias), search) || strings.Contains(s.Hostname, search) || strings.Contains(s.Notes, search) {
				s.Alias = strings.Replace(string(alias), search, fmt.Sprintf("%s%s%s", chalk.Yellow, search, chalk.Reset), 1)
				s.Hostname = strings.Replace(s.Hostname, search, fmt.Sprintf("%s%s%s", chalk.Yellow, search, chalk.Reset), 1)
				servers = append(servers, s)
			}
		}
		return nil
	})
	return
}

// AllServers ...  TODO merge with search?
func (dbw DBWrapper) AllServers() (servers []Server, err error) {
	err = dbw.db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte("servers"))
		c := b.Cursor()

		for k, v := c.First(); k != nil; k, v = c.Next() {
			var s Server
			err := json.Unmarshal(v, &s)
			handleError(err)
			servers = append(servers, s)
		}
		return nil
	})
	handleError(err)
	return
}

// GetServer ...
func (dbw DBWrapper) GetServer(alias string) (s Server, err error) {
	err = dbw.db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte("servers"))

		encoded := b.Get([]byte(alias))
		if encoded == nil {
			return errors.New("alias " + alias + " not found.")
		}

		err := json.Unmarshal(encoded, &s)
		handleError(err)
		return nil
	})
	return
}

// UpdateServer ...
func (dbw DBWrapper) UpdateServer(s Server) (err error) {
	return dbw.AddServer(s)
}
