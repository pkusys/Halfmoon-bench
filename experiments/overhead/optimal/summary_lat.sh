#!/bin/bash

set -xu

BASE_DIR=`realpath $(dirname $0)`
ROOT_DIR=`realpath $BASE_DIR/../../..`

cd $BASE_DIR

run=2

QPS=(100 200 300 400)
READ_RATIO=(0.1 0.3 0.5 0.7 0.9)
LOGMODE=("read" "write")

VALUE_SIZE=256

for qps in ${QPS[@]}; do
    for mode in ${LOGMODE[@]}; do
        echo QPS${qps}_${mode}_${run}
        for rr in ${READ_RATIO[@]}; do
            EXP_DIR=ReadRatio${rr}_QPS${qps}_${mode}_${run}
            if [ $qps -eq 400 ]; then
                EXP_DIR=ReadRatio${rr}_QPS${qps}_${mode}_v256_${run}
            fi
            head -n 1 results/$EXP_DIR/latency.txt | cat
        done
    done
done
