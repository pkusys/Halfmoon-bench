#!/bin/bash
BASE_DIR=`realpath $(dirname $0)`
ROOT_DIR=`realpath $BASE_DIR/../../..`

BOKI_MACHINE_IAM=boki-ae-experiments
HELPER_SCRIPT=$ROOT_DIR/scripts/exp_helper

RUN=$1

QPS=(20)
LOGMODE=("read" "write")

$HELPER_SCRIPT start-machines --base-dir=$BASE_DIR --instance-iam-role=$BOKI_MACHINE_IAM

for qps in ${QPS[@]}; do
    for mode in ${LOGMODE[@]}; do
        EXP_DIR=QPS${qps}_$mode
        $BASE_DIR/run_once.sh $EXP_DIR $qps $mode # 2>&1 | tee $BASE_DIR/run.log 
        mv $BASE_DIR/results/$EXP_DIR $BASE_DIR/results/${EXP_DIR}_$RUN
        echo "finished QPS${qps}_$mode"
    done
done

$HELPER_SCRIPT stop-machines --base-dir=$BASE_DIR
