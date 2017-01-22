package main

import (
	"encoding/json"
	"errors"
	"os"
	"os/user"
	"strings"
	"time"

	"github.com/boltdb/bolt"
)

// newBoltDB ...
func (dbw DBWrapper) connect() (db *bolt.DB) {
	usr, err := user.Current()
	handleError(err)

	dbpath := usr.HomeDir + "/.rain"
	err = os.MkdirAll(dbpath, 0755)
	handleError(err)

	fullpath := dbpath + "/bolt.db"
	db, err = bolt.Open(fullpath, 0600, &bolt.Options{Timeout: 1 * time.Second})
	handleError(err)

	err = db.Update(func(tx *bolt.Tx) error {
		_, err2 := tx.CreateBucketIfNotExists([]byte("servers"))
		return err2
	})

	handleError(err)

	return db
}

// DBWrapper ...
type DBWrapper struct {
}

// DeleteServer ...
func (dbw DBWrapper) DeleteServer(alias string) (err error) {
	bdb := dbw.connect()
	defer bdb.Close()

	err = bdb.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte("servers"))
		//TODO: check if key exists first
		err2 := b.Delete([]byte(alias))
		handleError(err2)
		return err
	})
	handleError(err)
	return
}

// AddServer ...
func (dbw DBWrapper) AddServer(s Server) (err error) {
	bdb := dbw.connect()
	defer bdb.Close()

	err = bdb.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte("servers"))
		handleError(err)

		encoded, err2 := json.Marshal(s)
		handleError(err2)

		err2 = b.Put([]byte(s.Alias), encoded)
		return err2
	})

	handleError(err)
	return
}

// ServerSearch ...
func (dbw DBWrapper) ServerSearch(search string) (servers []Server, err error) {
	bdb := dbw.connect()
	defer bdb.Close()

	err = bdb.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte("servers"))
		cursor := b.Cursor()

		for alias, host := cursor.First(); alias != nil; alias, host = cursor.Next() {
			var s Server
			err := json.Unmarshal(host, &s)
			handleError(err)

			searchL := strings.ToLower(search)
			aliasL := strings.ToLower(string(alias))
			hostL := strings.ToLower(s.Hostname)
			//notesL := strings.ToLower(s.Notes)

			if strings.Contains(aliasL, searchL) || strings.Contains(hostL, searchL) {
				servers = append(servers, s)
			}
		}
		return nil
	})
	return
}

// AllServers ...  TODO merge with search?
func (dbw DBWrapper) AllServers() (servers []Server, err error) {
	bdb := dbw.connect()
	defer bdb.Close()

	err = bdb.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte("servers"))
		c := b.Cursor()

		for k, v := c.First(); k != nil; k, v = c.Next() {
			var s Server
			err2 := json.Unmarshal(v, &s)
			handleError(err2)
			servers = append(servers, s)
		}
		return err
	})
	handleError(err)
	return
}

// GetServer ...
func (dbw DBWrapper) GetServer(alias string) (s Server, err error) {
	bdb := dbw.connect()
	defer bdb.Close()

	err = bdb.View(func(tx *bolt.Tx) error {
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
