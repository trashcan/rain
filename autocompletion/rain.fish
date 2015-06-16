# fish completion for rain
# https://github.com/trashcan/rain
# Please note, this is first-attempt at auto-complete support for fish and can/should be improved later.

function __fish_rain_needs_command
  set cmd (commandline -opc)
  if [ (count $cmd) -eq 1 -a $cmd[1] = 'rain' ]
    return 0
  end
  return 1
end

function __fish_rain_using_command
  set cmd (commandline -opc)
  if [ (count $cmd) -gt 1 ]
    if [ $argv[1] = $cmd[2] ]
      return 0
    end
  end
  return 1
end

function __fish_rain_list
  rain list | grep -v 'Hits' | awk '{ print $1 }'
end

function __fish_rain_sshhosts
    set rainhosts (__fish_rain_list | tr '\n' '|')
    cat ~/.ssh/config  | grep hostname | awk '{ print $2 }' | egrep -vi "$rainhosts ADSF" | uniq
end

# general options
complete -e -c rain
complete -f -c rain -n '__fish_rain_needs_command' -a help -d 'Display the manual of a rain command'
complete -f -c rain -n '__fish_rain_needs_command' -a list -d 'Display a list of stored server configs'
complete -f -c rain -n '__fish_rain_needs_command' -a ssh -d 'SSH to specified alias, and auto-reconnect on abnormal disconnections'
complete -f -c rain -n '__fish_rain_needs_command' -a add -d 'Add a new server config'
complete -f -c rain -n '__fish_rain_needs_command' -a delete -d 'Delete a server config'
complete -f -c rain -n '__fish_rain_needs_command' -a note -d 'View and edit notes for a server'
complete -f -c rain -n '__fish_rain_needs_command' -a search -d 'Search server aliases and notes'

# ssh
#complete -f -c rain -n '__fish_rain_using_command ssh' -a '(__fish_rain_list)' --description 'Alias'
# TODO: There's probably a cleaner way to do this:
for hostline in (rain list | grep -v 'Hits' | awk '{ print $1" "$2 }')
    complete -f -c rain -n '__fish_rain_using_command ssh' -a (echo $hostline | awk '{ print $1 }') --description (echo $hostline | awk '{ print $2 }')
    # TODO: Do we want to also autocomplete using hostnames from ~/.known_hosts like ssh does?
end

# note
# TODO: There's probably a cleaner way to do this:
for hostline in (rain list | grep -v 'Hits' | awk '{ print $1" "$2 }')
    complete -f -c rain -n '__fish_rain_using_command note' -a (echo $hostline | awk '{ print $1 }') --description (echo $hostline | awk '{ print $2 }')
end

# delete
# TODO: There's probably a cleaner way to do this:
for hostline in (rain list | grep -v 'Hits' | awk '{ print $1" "$2 }')
    complete -f -c rain -n '__fish_rain_using_command delete' -a (echo $hostline | awk '{ print $1 }') --description (echo $hostline | awk '{ print $2 }')
end

#   add [alias] [root@][hostname][:22]
# TODO: Figure out how to show the accepted arguments
complete -f -c rain -n '__fish_rain_using_command add' -a '(__fish_rain_sshhosts)' --description 'from .ssh/config'
