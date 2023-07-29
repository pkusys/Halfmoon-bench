#!/bin/bash
BASE_DIR=`realpath $(dirname $0)`
ROOT_DIR=`realpath $BASE_DIR/../../..`

BOKI_MACHINE_IAM=boki-ae-experiments
HELPER_SCRIPT=$ROOT_DIR/scripts/exp_helper

RUN=$1

# QPS=(50 100 150 200 250 300 350 400 450)
QPS=(100 200 300)

$HELPER_SCRIPT start-machines --base-dir=$BASE_DIR --instance-iam-role=$BOKI_MACHINE_IAM

for qps in ${QPS[@]}; do
    # EXP_DIR=QPS${qps}
    # $BASE_DIR/run_once.sh $EXP_DIR $qps # 2>&1 | tee run.log 
    # mv $BASE_DIR/results/$EXP_DIR $BASE_DIR/results/${EXP_DIR}_$RUN
    # echo "finished QPS${qps}"
    EXP_DIR=QPS${qps}
    if [ -d "$BASE_DIR/results/${EXP_DIR}_$RUN" ]; then
        echo "finished QPS${qps}"
        continue
    fi
    while true; do
        sleep 60
        $BASE_DIR/run_once.sh $EXP_DIR $qps # 2>&1 | tee run.log 
        if [ -s "$BASE_DIR/results/$EXP_DIR/async_results" ]; then
            mv $BASE_DIR/results/$EXP_DIR $BASE_DIR/results/${EXP_DIR}_$RUN
            echo "finished QPS${qps}"
            break
        else
            echo "retrying QPS${qps}"
            rm -rf $BASE_DIR/results/$EXP_DIR
        fi
    done
done

$HELPER_SCRIPT stop-machines --base-dir=$BASE_DIR
