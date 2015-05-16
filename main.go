package main

import (
	"bufio"
	"errors"
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"strings"

	"github.com/apcera/termtables"
	"github.com/ttacon/chalk"
)

// TODO: chalk sucks and doesn't check if the output is a terminal
// TODO: support ssh options like user, port

func main() {
	termtables.EnableUTF8PerLocale()
	flag.Usage = usage
	parseArgs()
}

func usage() {
	fmt.Println("☔")
	fmt.Printf("%s ssh <alias>: ssh to server by alias\n", os.Args[0])
	fmt.Printf("%s list: list all known servers\n", os.Args[0])
	fmt.Printf("%s add [alias] [hostname]: add a new server\n", os.Args[0])
	fmt.Printf("%s delete <alias>: delete server\n", os.Args[0])
	fmt.Printf("%s note <alias>: edit the notes of an existing server by alias\n", os.Args[0])
	fmt.Printf("%s help: print this message\n\n", os.Args[0])
}

func handleError(m error) {
	if m != nil {
		fmt.Printf("☔\t%s%s%s\n", chalk.Red, m.Error(), chalk.Reset)
		os.Exit(1)
	}
}

func handleWarning(m error) {
	if m != nil {
		fmt.Printf("☔\t%s%s%s\n", chalk.Yellow, m.Error(), chalk.Reset)
	}
}

func parseArgs() {
	requireArgs(os.Args[0], 1)
	args := os.Args[1:]

	switch args[0] {

	case "ssh":
		requireArgs("ssh", 2)
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
		requireArgs("note", 2)
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

	newServer := Server{
		Alias:    alias,
		Hostname: hostname,
		Notes:    string(""),
	}

	dbw := NewDBWrapper()
	err := dbw.AddServer(newServer)
	handleError(err)
}

func cmdList() {
	dbw := NewDBWrapper()
	servers, err := dbw.AllServers()
	handleError(err)
	renderServers(servers, "")
}

func cmdDelete(alias string) {
	dbw := NewDBWrapper()
	err := dbw.DeleteServer(alias)
	handleError(err)
}

func cmdSearch(search string) {
	dbw := NewDBWrapper()
	servers, err := dbw.ServerSearch(search)
	handleError(err)
	renderServers(servers, search)
}

func cmdNote(alias string) {
	dbw := NewDBWrapper()
	s, err := dbw.GetServer(alias)
	handleError(err)

	newNote := openEditor(s.Notes)
	if s.Notes != string(newNote) {
		s.Notes = newNote
		err = dbw.UpdateServer(s)
		handleError(err)
	}
}

func openEditor(notes string) (newNote string) {
	file, err := ioutil.TempFile(os.TempDir(), "rain")
	handleError(err)
	defer os.Remove(file.Name())

	err = ioutil.WriteFile(file.Name(), []byte(notes), 0644)
	handleError(err)

	cwd, err := os.Getwd()
	handleError(err)

	pa := os.ProcAttr{
		Files: []*os.File{os.Stdin, os.Stdout, os.Stderr},
		Dir:   cwd,
	}
	vim, err := exec.LookPath("vim")
	handleError(err)

	proc, err := os.StartProcess(vim, []string{"", file.Name()}, &pa)
	handleError(err)

	_, err = proc.Wait()
	handleError(err)

	newNoteByte, err := ioutil.ReadFile(file.Name())
	handleError(err)

	return string(newNoteByte)
}

func cmdSSH(alias string) {
	dbw := NewDBWrapper()
	s, err := dbw.GetServer(alias)
	if err != nil {
		search, _ := dbw.ServerSearch(alias)
		if len(search) == 0 {
			// Just ssh to the provided hostname.
			handleWarning(err)
			s = Server{Hostname: alias}
		} else if len(search) == 1 {
			// If there's one search result, ssh to it.
			handleWarning(errors.New("Matched one result, going to " + search[0].Hostname + "."))
			s = search[0]
			s.Hit++
			dbw.UpdateServer(s)
		} else {
			// Otherwise, list the search results and quit.
			renderServers(search, alias)
			return
		}
	} else {
		s.Hit++
		dbw.UpdateServer(s)
	}
	s.ssh()
}

func renderServers(servers []Server, highlight string) {
	if len(servers) == 0 {
		handleError(errors.New("No servers found."))
	}

	t := termtables.CreateTable()
	t.AddHeaders("Alias", "Hostname", "Hits")
	for _, s := range servers {
		if highlight != "" {
			s.Alias = strings.Replace(s.Alias, highlight, fmt.Sprintf("%s%s%s", chalk.Green, highlight, chalk.Reset), 1)
			s.Hostname = strings.Replace(s.Hostname, highlight, fmt.Sprintf("%s%s%s", chalk.Green, highlight, chalk.Reset), 1)
		}
		t.AddRow(s.Alias, s.Hostname, s.Hit)
	}
	fmt.Printf(t.Render())
}

func renderNotes(s Server) {
	fmt.Printf("☔\t%sNotes for %s%s\n\n", chalk.Green, s.Alias, chalk.Reset)
	fmt.Println(s.Notes)
}
