#!/bin/bash

cases=( bank multi_bank )
nemeses=( random_kill random_drop )
verifiers=( tidb_bank tidb_bank_tso )

mkdir -p var

for i in "${cases[@]}"
do 
    for j in "${nemeses[@]}"
    do 
        history_log=./var/history_"$i"_"$j".log
        echo "run $i with nemeses $j"
        ./bin/chaos-tidb --case $i --nemesis $j --history $history_log --request-count 200

        for k in "${verifiers[@]}"
        do 
            echo "use $k to check history" $history_log
            ./bin/chaos-verifier --history $history_log --names $k
            ret=$?
            if [ $ret -ne 0 ]; then
                exit $ret
            fi
        done 
    done 
done 