#!/bin/bash

_wombatt_completions() {
    local cur prev
    local common bi br bt dt mqtt p pi sp rto webs id start count regtype of off db sb par tout

    common="-h --help -v --version -l --log-level"

    # Common flags for serial ports
    db="-D --data-bits"
    sb="-S --stop-bits"
    par="-P --parity"
    tout="-t --timeout"

    # Flags for BatteryInfoCmd
    bi="-i --battery-id"
    br="-B --baud-rate"
    bt="--bms-type"
    dt="-T --device-type"
    p="--protocol"
    sp="-p --address"
    rto="-t --read-timeout"

    # Flags for ForwardCmd
    controller_port="--controller-port"
    subordinate_port="--subordinate-port"

    # Flags for InverterQueryCmd
    command="-c --command"

    # Flags for ModbusReadCmd
    mr_p="-R --protocol"
    mr_id="-i --id"
    id="--id"
    start="--start"
    count="--count"
    regtype="--register-type"
    of="-o --output-format"
    off="-O --output-format-file"

    # Flags for MonitorBatteriesCmd
    mqtt="--mqtt-broker --mqtt-password --mqtt-topic-prefix --mqtt-user"
    pi="-P --poll-interval"
    webs="-w --web-server-address"
    mqtt_prefix="--mqtt-prefix"

    

    

    cur=${COMP_WORDS[COMP_CWORD]}
    prev=${COMP_WORDS[COMP_CWORD-1]}

    case ${COMP_CWORD} in
        1)
            COMPREPLY=($(compgen -W "battery-info forward inverter-query modbus-read monitor-batteries monitor-inverters" -- "${COMP_WORDS[1]}"))
            ;;
        *)
            case ${prev} in
            "battery-info")
                COMPREPLY=($(compgen -W "$common $bi $br $bt $dt $p $rto $sp" -- ${cur}))
                ;;
            "forward")
                COMPREPLY=($(compgen -W "$common $br $dt $controller_port $subordinate_port" -- ${cur}))
                ;;
            "inverter-query")
                COMPREPLY=($(compgen -W "$common $br $dt $rto $sp $command $db $sb $par $mr_p $mr_id" -- ${cur}))
                ;;
            "modbus-read")
                COMPREPLY=($(compgen -W "$common $br $dt $mr_p $rto $sp $mr_id $start $count $regtype $of $off" -- ${cur}))
                ;;
            "monitor-batteries")
                COMPREPLY=($(compgen -W "$common $bi $br $bt $dt $mqtt $p $pi $rto $sp $webs $mqtt_prefix" -- ${cur}))
                ;;
            "monitor-inverters")
                COMPREPLY=($(compgen -W "$common $br $db $sb $par $dt $mqtt $pi $rto $webs $p $id" -- ${cur}))
                ;;
            
        esac
    esac

}

complete -F _wombatt_completions wombatt
