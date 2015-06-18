_rain() 
{
    local cur prev opts base
    COMPREPLY=()
    cur="${COMP_WORDS[COMP_CWORD]}"
    prev="${COMP_WORDS[COMP_CWORD-1]}"

    opts="list ssh add note search delete help"

    case "${prev}" in
        help)
            COMPREPLY=""
            return 0
            ;;
        list)
            COMPREPLY=""
            return 0
            ;;
        add)
            COMPREPLY=""
            return 0
            ;;
        note)
            
            return 0
            ;;
        ssh)
            local servers=$(for x in `rain list | awk '{print $1}' | tail -n+2`; do echo ${x} ; done )
            COMPREPLY=( $(compgen -W "${servers}" ${cur}) )
            return 0
            ;;
        search)
            return 0
            ;;
        delete)
            return 0
            ;;
        help)
            return 0
            ;;
    esac

   COMPREPLY=($(compgen -W "${opts}" -- ${cur}))  
   return 0
}
complete -F _rain rain
