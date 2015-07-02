Rain
====
Rain is a command line tool to store and categorize SSH hosts.

```
$ rain help
☔  ./rain <command> [options]

Commands:
  list
  ssh <alias>
  add [alias] [root@][hostname][:22]
  note <alias>
  search <alias|hostname|notes>
  delete <alias>
  help

Report bugs at http://github.com/trashcan/rain/issues.
```

Installation
============
Make sure you have your `$gopath` set. For bash you might do something like this:
```
$ mkdir -p ~/go/bin
$ echo "export GOPATH=~/go >> ~/.bashrc"
$ echo "export PATH=$PATH:$GOPATH/bin" >> ~/.bashrc
```

For fish:
```
$ mkdir -p ~/go/bin
$ echo "set -x GOPATH ~/go" >> ~/.config/fish/config.fish
$ echo "set -x PATH ~/go/bin $PATH" >> ~/.config/fish/config.sh"
```

Then just use `go get` to retrieve rain:
```
$ go get github.com/trashcan/rain
```

Usage
=====

Hosts are saved with `rain add`.
```
$ rain add router admin@192.168.1.1
☔	admin@192.168.1.1 added successfully.
```

After that, `rain ssh` will connect to the server. `router` is the friendly name that is passed to the ssh subcommand.

```
$ rain ssh router
☔	Connecting to admin@192.168.1.1.
Last login: Tue Jun  9 14:05:22 2015 from 192.168.1.131
```

You can also run ad-hoc commands on the remote server by appending the command or commands after the hostname.

```
$ rain ssh router echo hello world
☔	Connecting to admin@192.168.1.1.
hello world
Connection to admin@192.168.1.1 closed.
```

All of the stored hosts can be listed with `rain list`.

```
$ rain list
Alias              Hostname                     Hits
c7-employee        206.225.85.58:33             5
gitrex             gitrex.com                   56
rpi                pi@192.168.1.131:33          8
rpi2               192.168.1.46                 14
strider            10.48.6.50                   1
```

The list of servers can be searched with `rain search`.
```
$ rain search rpi
Alias    Hostname               Hits
rpi      pi@192.168.1.131:33    8
rpi2     192.168.1.46           14

$ rain search 192.168.1
Alias     Hostname               Hits
router    admin@192.168.1.1      1
rpi       pi@192.168.1.131:33    8
rpi2      192.168.1.46           14
```
Not visible here, but the matching substrings are highlighted with color.


Features
========

Automatic reconnection
----------------------
If an SSH connection drops or SSH returns an unusual code, rain will automatically reconnect.
```
$ rain ssh managed
☔	Connecting to pat@managed.mydomain.com:33.
Last login: Tue Jun  9 14:19:11 2015 from 10.16.1.75
[pat@managed ~]$ killall sshd
☔	Reconnecting. Press Ctrl+C to abort.
Last login: Tue Jun  9 14:20:08 2015 from 10.16.1.75
[pat@managed ~]$
```

Tab Completion
---------------
Press tab to autocomplete commands as well as hostnames.
```
$ rain ssh rpi<TAB>
rpi  (pi@192.168.1.131:33)  rpi2  (192.168.1.46)
```
This currently works with bash and fish shell.

Notes
-----
Notes can be added to each saved server. The syntax is `rain note friendlyname`. This will open vim with the contents of any existing notes. These notes are searched also when using `rain search`, so it's helpful to add keywords into the notes.
```
$ rain search gitlab
Alias     Hostname      Hits
gitrex    gitrex.com    57
```

These notes will be shown at each login as well.
```
$ rain ssh gitrex
☔	Connecting to gitrex.com.
☔	Notes for gitrex:
Gitrex Gitlab Server
☔	Connected.
[root@gitrex ~]#
```

Best search match connection
----------------------------
If you `rain ssh` to a server that isn't in local database, but the name provided matches exactly one server as a substring, rain will automatically connect to that server.
```
$ rain search yapsbuilder
Alias              Hostname      Hits
phx-yapsbuilder    10.48.6.24    3
$ rain ssh yaps
☔	Matched one result, going to 10.48.6.24.
Last login: Thu Jun  4 13:14:07 2015 from 10.16.1.96
[root@phx1-yapsbuilder1 ~]#
```

No match connection
-------------------
If no matches are found, rain will connect to the provided hostname. It will still auto-reconnect if the SSH connection drops.


Suggestions/Bugs
================
Feedback is very welcome. Please send any here: https://github.com/trashcan/rain/issues
