#!/bin/bash
BASE_DIR=`realpath $(dirname $0)`
ROOT_DIR=`realpath $BASE_DIR/../../..`

BOKI_MACHINE_IAM=boki-ae-experiments
HELPER_SCRIPT=$ROOT_DIR/scripts/exp_helper

RUN=$1

QPS=(100 200 300 400 500 600 700 800 900 1000 1100)

$HELPER_SCRIPT start-machines --base-dir=$BASE_DIR --instance-iam-role=$BOKI_MACHINE_IAM

if ! [ -f "$BASE_DIR/machines.json" ]; then
    echo "[ERROR] machines not started, skipping $BASE_DIR"
    rm ":~"
    exit 1
fi

for qps in ${QPS[@]}; do
    EXP_DIR=QPS${qps}
    if [ -d "$BASE_DIR/results/${EXP_DIR}_$RUN" ]; then
        echo "finished $BASE_DIR/$EXP_DIR"
        continue
    fi

    sleep 60
    $BASE_DIR/run_once.sh $EXP_DIR $qps # 2>&1 | tee run.log
    if [ -s "$BASE_DIR/results/$EXP_DIR/async_results" ]; then
        mv $BASE_DIR/results/$EXP_DIR $BASE_DIR/results/${EXP_DIR}_$RUN
        echo "finished $BASE_DIR/$EXP_DIR"
        # break
    else
        echo "retry $BASE_DIR/$EXP_DIR"
        # rm -rf $BASE_DIR/results/$EXP_DIR
    fi
done

$HELPER_SCRIPT stop-machines --base-dir=$BASE_DIR
