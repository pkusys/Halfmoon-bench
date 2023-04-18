#!/bin/bash
BASE_DIR=`realpath $(dirname $0)`
ROOT_DIR=`realpath $BASE_DIR/../../..`

BOKI_MACHINE_IAM=boki-ae-experiments
HELPER_SCRIPT=$ROOT_DIR/scripts/exp_helper

RUN=$1

# QPS=(100 200 300 400 500 600 700)
QPS=(1000 1100)

$HELPER_SCRIPT start-machines --base-dir=$BASE_DIR --instance-iam-role=$BOKI_MACHINE_IAM

for qps in ${QPS[@]}; do
    EXP_DIR=QPS${qps}
    $BASE_DIR/run_once.sh $EXP_DIR $qps # 2>&1 | tee run.log 
    cp $BASE_DIR/docker-compose.yml $BASE_DIR/results/$EXP_DIR
    cp $BASE_DIR/docker-compose-generated.yml $BASE_DIR/results/$EXP_DIR
    cp $BASE_DIR/config.json $BASE_DIR/results/$EXP_DIR
    cp $BASE_DIR/nightcore_config.json $BASE_DIR/results/$EXP_DIR
    mv $BASE_DIR/results/$EXP_DIR $BASE_DIR/results/${EXP_DIR}_$RUN
    echo "finished QPS${qps}"
done

$HELPER_SCRIPT stop-machines --base-dir=$BASE_DIR
