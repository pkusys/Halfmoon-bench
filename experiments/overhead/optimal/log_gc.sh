#!/bin/bash

set -xu

BASE_DIR=`realpath $(dirname $0)`
ROOT_DIR=`realpath $BASE_DIR/../../..`

cd $BASE_DIR

run=3

QPS=(100)
NUM_OPS=(10)
READ_RATIO=(0.1 0.3 0.5 0.7 0.9)
LOGMODE=("read" "write")
GC=(10000 60000) # in ms, 1s, 10s, 1min

NUM_KEYS=10000
VALUE_SIZE=(256 1024)

for qps in ${QPS[@]}; do
    for ops in ${NUM_OPS[@]}; do
        for mode in ${LOGMODE[@]}; do
            for v in ${VALUE_SIZE[@]}; do
                for gc in ${GC[@]}; do
                    echo QPS${qps}_${mode}_v${v}_${gc}_${run}
                    for rr in ${READ_RATIO[@]}; do
                        EXP_DIR=ReadRatio${rr}_QPS${qps}_${mode}_v${v}_${run}
                        $ROOT_DIR/scripts/compute_logsize.py --async-result-file results/$EXP_DIR/async_results \
                            --num-keys $NUM_KEYS --value-size $v --gc-interval $gc
                    done
                done
            done
        done
    done
done
