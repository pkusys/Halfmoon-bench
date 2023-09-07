#!/bin/bash
BASE_DIR=`realpath $(dirname $0)`
ROOT_DIR=`realpath $BASE_DIR/../../..`

BOKI_MACHINE_IAM=boki-ae-experiments
HELPER_SCRIPT=$ROOT_DIR/scripts/exp_helper

RUN=$1

# QPS=(50 100 150 200 250 300 350 400 450)
QPS=(100 200 300)
LOGMODE=("read" "write")
MAX_RETRIES=3

$HELPER_SCRIPT start-machines --base-dir=$BASE_DIR --instance-iam-role $BOKI_MACHINE_IAM

if ! [ -f "$BASE_DIR/machines.json" ]; then
    echo "[ERROR] machines not started, skipping $BASE_DIR"
    rm ":~"
    exit 1
fi

for qps in ${QPS[@]}; do
    for mode in ${LOGMODE[@]}; do
        EXP_DIR=QPS${qps}_$mode
        if [ -d "$BASE_DIR/results/${EXP_DIR}_$RUN" ]; then
            echo "finished $BASE_DIR/$EXP_DIR"
            continue
        fi

        while [ $MAX_RETRIES -gt 0 ]; do
            sleep 60
            ((MAX_RETRIES--))
            $BASE_DIR/run_once.sh $EXP_DIR $qps # 2>&1 | tee run.log 
            if [ -s "$BASE_DIR/results/$EXP_DIR/async_results" ]; then
                mv $BASE_DIR/results/$EXP_DIR $BASE_DIR/results/${EXP_DIR}_$RUN
                echo "finished $BASE_DIR/$EXP_DIR"
                break
            else
                echo "retrying $BASE_DIR/$EXP_DIR"
                rm -rf $BASE_DIR/results/$EXP_DIR
            fi
        done
    done
done

$HELPER_SCRIPT stop-machines --base-dir=$BASE_DIR
