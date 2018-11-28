#!/bin/bash

for bin in $@; do
    case $bin in
    'tidb' )
        suit=chaos-tidb
        cases=( bank multi_bank )
        nemeses=( random_kill random_drop )
        ;;
    'rawkv' )
        suit=chaos-rawkv
        cases=( register )
        # TODO: add random_drop, chaos can not heal drop nemesis sometime.
        nemeses=( random_kill  )
        ;;
    'txnkv' )
        suit=chaos-txnkv
        cases=( register )
        nemeses=( random_kill  )
        ;;
    '--help' )
        HELP=1
        ;;
    *)
        echo "unknown option $1"
        exit 1
        ;;
    esac
    shift
done

if [ "$HELP" ]; then
    echo "usage: $0 [OPTION]"
    echo "  tidb                                           Chaos test TiDB"
    echo "  rawkv                                          Chaos test RawKV"
    echo "  txnkv                                          Chaos test TxnKV"
    echo "  --help                                         Display this message"
    exit 0
fi

mkdir -p var

for i in "${cases[@]}"
do
    for j in "${nemeses[@]}"
    do
        history_log=./var/history_"$suit"_"$i"_"$j".log
        echo "run $i with nemeses $j"
        ./bin/$suit \
            --case $i \
            --nemesis $j \
            --history $history_log \
            --request-count 200 \
            --round 10
    done
done
