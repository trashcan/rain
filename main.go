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

// TODO: chalk doesn't check if the output supports color
// TODO: (more correctly) support ssh options like port: user@port:22
// TODO: parse and filter ssh -v to see when we are connected for slow logins.

func main() {
	flag.Usage = usage
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
	fmt.Println("  delete <alias>")
	fmt.Println("  help")
	fmt.Println()
	fmt.Println("Report bugs at http://gitrex.com/patl/rain/issues.")
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
		fmt.Print("Hostname ([user]@<hostname>:[port]): ")
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
		fmt.Printf("☔\tConnecting to %s.\n", s.Hostname)
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
	// TODO FIXME: These are adding a blank line above the headers.
	// t.AddHeaders("Alias", "Hostname", "Hits")
	t.AddRow("Alias", "Hostname", "Hits")
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
