#!/bin/bash
BASE_DIR=`realpath $(dirname $0)`
ROOT_DIR=`realpath $BASE_DIR/../../..`

BOKI_MACHINE_IAM=boki-ae-experiments
HELPER_SCRIPT=$ROOT_DIR/scripts/exp_helper

RUN=$1

$HELPER_SCRIPT start-machines --base-dir=$BASE_DIR --instance-iam-role=$BOKI_MACHINE_IAM

cd $BASE_DIR
if ! [ -f "$BASE_DIR/machines.json" ]; then
    echo "[ERROR] machines not started, skipping $BASE_DIR"
    rm ":~"
    exit 1
fi

./run_all_runtime.sh $RUN
sleep 10
./run_all_storage.sh $RUN

$HELPER_SCRIPT stop-machines --base-dir=$BASE_DIR