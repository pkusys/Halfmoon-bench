#!/bin/bash
BASE_DIR=`realpath $(dirname $0)`
ROOT_DIR=`realpath $BASE_DIR/../../..`

BOKI_MACHINE_IAM=boki-ae-experiments
HELPER_SCRIPT=$ROOT_DIR/scripts/exp_helper

RUN=$1

QPS=(100)
NUM_OPS=(10)
READ_RATIO=(0 1)
VALUE_SIZE=(256)
FAIL_RATE=(0.1 0.2 0.3 0.4)

$HELPER_SCRIPT start-machines --base-dir=$BASE_DIR --instance-iam-role=$BOKI_MACHINE_IAM

if ! [ -f "$BASE_DIR/machines.json" ]; then
    echo "[ERROR] machines not started, skipping $BASE_DIR"
    rm ":~"
    exit 1
fi

for qps in ${QPS[@]}; do
    for ops in ${NUM_OPS[@]}; do
        for rr in ${READ_RATIO[@]}; do
            for f in ${FAIL_RATE[@]}; do
                for v in ${VALUE_SIZE[@]}; do
                    EXP_DIR=ReadRatio${rr}_QPS${qps}_v${v}_f${f}
                    if [ -d "$BASE_DIR/results/${EXP_DIR}_$RUN" ]; then
                        echo "finished $BASE_DIR/$EXP_DIR"
                        continue
                    fi
                    sleep 60
                    $BASE_DIR/run_once.sh $EXP_DIR $qps $ops $rr $v $f # 2>&1 | tee $BASE_DIR/run.log 
                    mv $BASE_DIR/results/$EXP_DIR $BASE_DIR/results/${EXP_DIR}_$RUN
                    echo "finished $BASE_DIR/$EXP_DIR"
                done
            done
        done
    done
done

$HELPER_SCRIPT stop-machines --base-dir=$BASE_DIR

