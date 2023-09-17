#!/bin/bash

_wombatt_completions() {
    local cur prev

    cur=${COMP_WORDS[COMP_CWORD]}
    prev=${COMP_WORDS[COMP_CWORD-1]}

    case ${COMP_CWORD} in
        1)
            COMPREPLY=($(compgen -W "monitor-inverters monitor-batteries inverter-query battery-info forward" -- "${COMP_WORDS[1]}"))
            ;;
        *)
            case ${prev} in
            "battery-info")
                COMPREPLY=($(compgen -W "-B --baud-rate -p --serial-port -i --battery-ids -t --read-timeout --battery-type" -- ${cur}))
                ;;
            "forward")
                COMPREPLY=($(compgen -W "-B --baud-rate --controller-port --subordinate-port" -- ${cur}))
                ;;
            "inverter-query")
                COMPREPLY=($(compgen -W "-p --serial-port -c --commands -t --read-timeout -B --baud-rate -T --device-type" -- ${cur}))
                ;;
            "modbus-read")
                COMPREPLY=($(compgen -W "--protocol -B --baud-rate -p --serial-port --id --start --count -t --read-timeout --battery-type" -- ${cur}))
                ;;
            "monitor-batteries")
                COMPREPLY=($(compgen -W "-B --baud-rate -p --serial-port -i --battery-ids -P --poll-interval -t --read-timeout --battery-type -w --web-port " -- ${cur}))
                ;;
            "monitor-inverters")
                COMPREPLY=($(compgen -W "-B --baud-rate -P --poll-interval -t --read-timeout -w --web-port -T --device-type" -- ${cur}))
                ;;
        esac
    esac

}

complete -F _wombatt_completions wombatt
