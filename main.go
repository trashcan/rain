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

/* TODO:
 * Detect non-terminal output and remove color. https://github.com/ttacon/chalk/issues/4
 * More correctly support ssh options like port: user@port:22
 * Parse and filter ssh -v to see when we are connected for slow logins.
 * Recent history, maybe even rain last to login to last server.
 * Configuration to disable unwanted features.
 * Adding does not detect when it is overwriting an existing server.
 * Smarter detection for typos in friendly name.
 */

func main() {
	flag.Usage = usage
	//chalk.DetectTerminal()
	parseArgs()
}

func usage() {
	fmt.Println("☔  ./rain <command> [options]")
	fmt.Println()

	fmt.Println("Commands:")
	fmt.Println("  list")
	fmt.Println("  ssh <alias>")
	fmt.Println("  add [alias] [root@][hostname][:22]")
	fmt.Println("  note <alias>")
	fmt.Println("  search <alias|hostname|notes>")
	fmt.Println("  delete <alias>")
	fmt.Println("  help")
	fmt.Println()
	fmt.Println("Report bugs at http://github.com/trashcan/rain/issues.")
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

func handleStatus(m string) {
	fmt.Printf("☔\t%s%s%s\n", chalk.Green, m, chalk.Reset)
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
	case "edit":
		requireArgs("note", 2)
		cmdNote(args[1])
	case "help":
		flag.Usage()
	default:
		flag.Usage()
		handleError(fmt.Errorf("unknown subcommand: %s", os.Args[1]))
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
		fmt.Print("Hostname ([user]@<hostname>[:port]): ")
		scanner.Scan()
		hostname = scanner.Text()
	}

	newServer := Server{
		Alias:    alias,
		Hostname: hostname,
		Notes:    string(""),
	}

	dbw := DBWrapper{}
	err := dbw.AddServer(newServer)
	handleError(err)
	handleStatus(hostname + " added successfully.")
}

func cmdList() {
	dbw := DBWrapper{}
	servers, err := dbw.AllServers()
	handleError(err)
	renderServers(servers, "")
}

func cmdDelete(alias string) {
	dbw := DBWrapper{}
	err := dbw.DeleteServer(alias)
	handleError(err)
}

func cmdSearch(search string) {
	dbw := DBWrapper{}
	servers, err := dbw.ServerSearch(search)
	handleError(err)
	renderServers(servers, search)
}

func cmdNote(alias string) {
	dbw := DBWrapper{}
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
	// Create a tempfile
	file, err := ioutil.TempFile(os.TempDir(), "rain")
	handleError(err)
	// Delete it when done.
	defer os.Remove(file.Name())

	// Write the current notes into the file
	err = ioutil.WriteFile(file.Name(), []byte(notes), 0644)
	handleError(err)

	// Launch vim to edit the notes
	cmd := exec.Command("vim", []string{file.Name()}...)
	cmd.Stderr = os.Stderr
	cmd.Stdout = os.Stdout
	cmd.Stdin = os.Stdin
	err = cmd.Start()
	handleError(err)
	err = cmd.Wait()
	handleError(err)

	// Read the updated notes
	newNoteByte, err := ioutil.ReadFile(file.Name())
	handleError(err)

	return string(newNoteByte)
}

func cmdSSH(alias string) {
	dbw := DBWrapper{}
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
		handleStatus(fmt.Sprintf("Connecting to %s.", s.Hostname))
		s.Hit++
		dbw.UpdateServer(s)
	}
	s.ssh()
}

func renderServers(servers []Server, highlight string) {
	if len(servers) == 0 {
		handleError(errors.New("No servers found."))
	}

	var ts = &termtables.TableStyle{
		SkipBorder: true,
		BorderX:    "", BorderY: "", BorderI: "",
		PaddingLeft: 0, PaddingRight: 4,
		Width:     80,
		Alignment: termtables.AlignLeft,
	}

	t := termtables.CreateTable()
	t.Style = ts
	cb := chalk.Bold.TextStyle

	// TODO FIXME: These are adding a blank line above the headers.
	// t.AddHeaders("Alias", "Hostname", "Hits")
	t.AddRow(cb("Alias"), cb("Hostname"), cb("Hits"))

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
	handleStatus(fmt.Sprintf("Notes for %s:", s.Alias))
	fmt.Println(s.Notes)
}
