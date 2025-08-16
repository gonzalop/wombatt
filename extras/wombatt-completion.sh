#!/bin/bash

_wombatt_completions() {
    local cur prev
    local common bi br bt dt mqtt p pi sp rto webs id start count regtype of off db sb par tout

    common="-h --help -v --version -l --log-level"

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

    # Flags for MonitorInvertersCmd (inherits some from above)
    mi_br="-B --baud-rate"
    mi_db="--data-bits"
    mi_sb="--stop-bits"
    mi_par="--parity"
    mi_p="-R --protocol"
    mi_id="-i --id"

    # Flags for SolarkQueryCmd
    db="-D --data-bits"
    sb="-S --stop-bits"
    par="-P --parity"
    tout="-t --timeout"

    cur=${COMP_WORDS[COMP_CWORD]}
    prev=${COMP_WORDS[COMP_CWORD-1]}

    case ${COMP_CWORD} in
        1)
            COMPREPLY=($(compgen -W "battery-info forward inverter-query modbus-read monitor-batteries monitor-inverters solark-query" -- "${COMP_WORDS[1]}"))
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
                COMPREPLY=($(compgen -W "$common $br $dt $rto $sp $command" -- ${cur}))
                ;;
            "modbus-read")
                COMPREPLY=($(compgen -W "$common $br $dt $mr_p $rto $sp $mr_id $start $count $regtype $of $off" -- ${cur}))
                ;;
            "monitor-batteries")
                COMPREPLY=($(compgen -W "$common $bi $br $bt $dt $mqtt $p $pi $rto $sp $webs $mqtt_prefix" -- ${cur}))
                ;;
            "monitor-inverters")
                COMPREPLY=($(compgen -W "$common $mi_br $mi_db $mi_sb $mi_par $dt $mqtt $pi $rto $webs $mi_p $mi_id" -- ${cur}))
                ;;
            "solark-query")
                COMPREPLY=($(compgen -W "$common $sp $br $db $sb $par $tout $p $id" -- ${cur}))
                ;;
        esac
    esac

}

complete -F _wombatt_completions wombatt
