package main

import (
	"bufio"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"os/user"
	"strings"
	"time"

	"github.com/apcera/termtables"
	"github.com/boltdb/bolt"
	"github.com/ttacon/chalk"
)

// TODO: getDB should only be called once
// TODO: case insensitive search
// TODO: chalk sucks and doesn't check if the output is a terminal

func main() {
	termtables.EnableUTF8PerLocale()
	flag.Usage = usage
	parseArgs()
}

func usage() {
	fmt.Println("Usage:")
	fmt.Printf("\t%s ssh <alias>: ssh to server by alias\n", os.Args[0])
	fmt.Printf("\t%s list: list all known servers\n", os.Args[0])
	fmt.Printf("\t%s add [alias] [hostname]: add a new server\n", os.Args[0])
	fmt.Printf("\t%s delete <alias>: delete server\n", os.Args[0])
	fmt.Printf("\t%s note <alias>: edit the notes of an existing server by alias\n", os.Args[0])
	fmt.Printf("\t%s help: print this message\n\n", os.Args[0])
}

func handleError(m error) {
	if m != nil {
		fmt.Printf("%s%s%s", chalk.Red, m.Error(), chalk.Reset)
		os.Exit(1)
	}
}

func parseArgs() {
	requireArgs(os.Args[0], 1)
	args := os.Args[1:]

	switch args[0] {

	case "ssh":
		requireArgs("search", 2)
		cmdSSH(args[1])
	case "list":
		cmdList()
	case "add":
		cmdAdd()
	case "delete":
		requireArgs("delete", 2)
		cmdDelete(args[1])
	case "search":
		requireArgs("search", 2)
		cmdSearch(args[1])
	case "note":
		requireArgs("search", 2)
		cmdNote(args[1])
	case "help":
		flag.Usage()
	default:
		fmt.Println("Unknown subcommand:", os.Args[1])
		flag.Usage()
	}
}

func requireArgs(cmd string, count int) {
	args := os.Args[1:]

	if len(args) < count {
		flag.Usage()
		handleError(fmt.Errorf("%s requires more argument(s).\n", cmd))
	}
}

func getDB() (db *bolt.DB) {
	usr, err := user.Current()
	handleError(err)

	dbpath := usr.HomeDir + "/.rain"
	err = os.MkdirAll(dbpath, 0755)
	handleError(err)

	fullpath := dbpath + "/bolt.db"
	db, err = bolt.Open(fullpath, 0600, &bolt.Options{Timeout: 1 * time.Second})
	handleError(err)

	db.Update(func(tx *bolt.Tx) error {
		_, err := tx.CreateBucketIfNotExists([]byte("servers"))
		handleError(err)
		return err
	})

	return db
}

func cmdAdd() {
	var alias, hostname string
	if len(os.Args) == 4 {
		alias = os.Args[2]
		hostname = os.Args[3]
	} else {
		scanner := bufio.NewScanner(os.Stdin)
		fmt.Print("Alias: ")
		scanner.Scan()
		alias = scanner.Text()
		fmt.Print("Hostname/IP: ")
		scanner.Scan()
		hostname = scanner.Text()
	}

	newServer := &Server{
		Alias:    alias,
		Hostname: hostname,
		Notes:    string(""),
	}

	newServer.save()
}

func cmdList() {
	table := termtables.CreateTable()
	table.AddHeaders("Alias", "Hostname")

	db := getDB()

	var rows []interface{}

	err := db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte("servers"))
		c := b.Cursor()

		for k, v := c.First(); k != nil; k, v = c.Next() {
			rows = append(rows, "test")
			var s Server
			err := json.Unmarshal(v, &s)
			handleError(err)
			table.AddRow(s.Alias, s.Hostname)
		}

		fmt.Printf(table.Render())
		return nil
	})
	handleError(err)
}

func cmdDelete(alias string) {
	db := getDB()
	err := db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte("servers"))
		//TODO: check if key exists first
		err := b.Delete([]byte(alias))
		return err
	})
	handleError(err)
}

func cmdSearch(search string) {
	db := getDB()

	db.View(func(tx *bolt.Tx) error {
		table := termtables.CreateTable()
		table.AddHeaders("Alias", "Hostname")

		b := tx.Bucket([]byte("servers"))
		cursor := b.Cursor()
		match := false
		for alias, host := cursor.First(); alias != nil; alias, host = cursor.Next() {
			var s Server
			err := json.Unmarshal(host, &s)
			handleError(err)

			if strings.Contains(string(alias), search) || strings.Contains(s.Hostname, search) || strings.Contains(s.Notes, search) {
				s.Alias = strings.Replace(string(alias), search, fmt.Sprintf("%s%s%s", chalk.Yellow, search, chalk.Reset), 1)
				s.Hostname = strings.Replace(s.Hostname, search, fmt.Sprintf("%s%s%s", chalk.Yellow, search, chalk.Reset), 1)
				table.AddRow(s.Alias, s.Hostname)
				match = true
			}

		}
		if match {
			fmt.Println(table.Render())
		} else {
			fmt.Println("No matches.")
		}
		return nil
	})
}

func cmdNote(alias string) {
	db := getDB()

	db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte("servers"))
		j := b.Get([]byte(alias))
		if j == nil {
			return errors.New("alias " + alias + " not found.")
		}

		var s Server
		err := json.Unmarshal(j, &s)
		if err != nil {
			panic(err)
		}

		newNote := openEditor(s.Notes)

		if s.Notes != string(newNote) {
			s.Notes = newNote
			encoded, err := json.Marshal(s)
			handleError(err)

			err = b.Put([]byte(alias), encoded)
			handleError(err)

			// Can't nest this
			//s.save()
		}

		return nil
	})

}

func openEditor(notes string) (newNote string) {
	file, err := ioutil.TempFile(os.TempDir(), "rain")
	if err != nil {
		panic(err)
	}
	defer os.Remove(file.Name())

	err = ioutil.WriteFile(file.Name(), []byte(notes), 0644)
	handleError(err)

	cwd, err := os.Getwd()
	handleError(err)

	pa := os.ProcAttr{
		Files: []*os.File{os.Stdin, os.Stdout, os.Stderr},
		Dir:   cwd,
	}

	// TODO: use path
	proc, err := os.StartProcess("/usr/local/bin/vim", []string{"", file.Name()}, &pa)
	handleError(err)

	_, err = proc.Wait()
	handleError(err)

	newNoteByte, err := ioutil.ReadFile(file.Name())
	handleError(err)

	return string(newNoteByte)
}

func cmdSSH(hostname string) {
	db := getDB()
	err := db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte("servers"))
		host := b.Get([]byte(hostname))

		var s Server
		if host != nil {
			err := json.Unmarshal(host, &s)
			handleError(err)
		} else {
			s.Hostname = hostname
		}

		s.ssh()
		return nil
	})
	handleError(err)
}
