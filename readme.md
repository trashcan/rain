```
pat@mbp ~> rain help
Usage of rain:
    rain ssh <alias>: ssh to server by alias
    rain list: list all known servers
    rain add: add a new server
    rain delete <alias>: delete server
    rain note <alias>: edit the notes of an existing server by alias
    rain help: print this message

pat@mbp ~> rain add
Alias: managed
Hostname/IP: managed.codero.com
pat@mbp ~> rain list
╭─────────┬────────────────────╮
│ Alias   │ Hostname           │
├─────────┼────────────────────┤
│ managed │ managed.codero.com │
╰─────────┴────────────────────╯
pat@mbp ~> rain search codero
╭─────────┬────────────────────╮
│ Alias   │ Hostname           │
├─────────┼────────────────────┤
│ managed │ managed.codero.com │
╰─────────┴────────────────────╯
pat@mbp ~> rain ssh managed
Last login: Thu May 14 20:56:48 2015 from 10.48.15.174
[pat@managed ~]$ killall sshd
Connection to managed.codero.com closed by remote host.
Connection to managed.codero.com closed.
Unusual termination, reconnecting (or the last ran command did not return 0).
Last login: Thu May 14 21:15:51 2015 from 10.48.15.174
[pat@managed ~]$ exit
logout
Connection to managed.codero.com closed.
pat@mbp ~>
```
