#!/bin/bash

set -xu

BASE_DIR=`realpath $(dirname $0)`
ROOT_DIR=`realpath $BASE_DIR/../../..`

cd $BASE_DIR

run=0

QPS=(100 200 300 400)
READ_RATIO=(0.1 0.3 0.5 0.7 0.9)
# LOGMODE=("read" "write")

VALUE_SIZE=256

for qps in ${QPS[@]}; do
    # for mode in ${LOGMODE[@]}; do
        echo QPS${qps}_${run}
        for rr in ${READ_RATIO[@]}; do
            EXP_DIR=ReadRatio${rr}_QPS${qps}_v${VALUE_SIZE}_${run}
            head -n 1 results/$EXP_DIR/latency.txt | cat
        done
    # done
done
