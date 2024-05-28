#!/bin/bash

_wombatt_completions() {
    local cur prev
    local common bi br bt dt mqtt p pi sp rto webs

    common="-h --help -v --version -l --log-level"

    bi="-i --battery-id"
    br="-B --baud-rate"
    bt="--bms-type"
    dt="-T --device-type"
    mqtt="--mqtt-server --mqtt-password --mqtt-topic-prefix --mqtt-user"
    p="--protocol"
    pi="-P --poll-interval"
    sp="-p --address"
    rto="-t --read-timeout"
    webs="-w --web-server-address"

    cur=${COMP_WORDS[COMP_CWORD]}
    prev=${COMP_WORDS[COMP_CWORD-1]}

    case ${COMP_CWORD} in
        1)
            COMPREPLY=($(compgen -W "battery-info forward inverter-query monitor-batteries monitor-inverters" -- "${COMP_WORDS[1]}"))
            ;;
        *)
            case ${prev} in
            "battery-info")
                COMPREPLY=($(compgen -W "$common $bi $br $bt $dt $p $rto $sp" -- ${cur}))
                ;;
            "forward")
                COMPREPLY=($(compgen -W "$common $br $bt --controller-port --subordinate-port" -- ${cur}))
                ;;
            "inverter-query")
                COMPREPLY=($(compgen -W "$common $br $dt $rto $sp -c --command" -- ${cur}))
                ;;
            "modbus-read")
                COMPREPLY=($(compgen -W "$common $br $dt $p $rto $sp --id --start --count" -- ${cur}))
                ;;
            "monitor-batteries")
                COMPREPLY=($(compgen -W "$common $bi $br $bt $dt $mqtt $p $pi $rto $sp $webs --mqtt-prefix" -- ${cur}))
                ;;
            "monitor-inverters")
                COMPREPLY=($(compgen -W "$common $dt $mqtt $pi $rto $webs" -- ${cur}))
                ;;
        esac
    esac

}

complete -F _wombatt_completions wombatt
