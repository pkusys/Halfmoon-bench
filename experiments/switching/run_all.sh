#!/bin/bash
BASE_DIR=`realpath $(dirname $0)`
ROOT_DIR=`realpath $BASE_DIR/../..`

BOKI_MACHINE_IAM=boki-ae-experiments
HELPER_SCRIPT=$ROOT_DIR/scripts/exp_helper

RUN=$1

CONCURRENCY=(16 64)
NUM_OPS=(5)
READ_RATIOS=("0.2,0.8")

# $HELPER_SCRIPT start-machines --base-dir=$BASE_DIR --instance-iam-role=$BOKI_MACHINE_IAM

for cc in ${CONCURRENCY[@]}; do
    for ops in ${NUM_OPS[@]}; do
        for rr in ${READ_RATIOS[@]}; do
            EXP_DIR=con${cc}_n${ops}_rr${rr}
            if [ -d "$BASE_DIR/results/${EXP_DIR}_$run" ]; then
                echo "finished $EXP_DIR"
                continue
            fi
            $HELPER_SCRIPT start-machines --base-dir=$BASE_DIR --instance-iam-role=$BOKI_MACHINE_IAM
            $BASE_DIR/run_once.sh $EXP_DIR $cc $ops $rr 2>&1 | tee $BASE_DIR/run.log 
            cp $BASE_DIR/docker-compose.yml $BASE_DIR/results/$EXP_DIR
            cp $BASE_DIR/docker-compose-generated.yml $BASE_DIR/results/$EXP_DIR
            cp $BASE_DIR/config.json $BASE_DIR/results/$EXP_DIR
            cp $BASE_DIR/nightcore_config.json $BASE_DIR/results/$EXP_DIR
            mv $BASE_DIR/results/$EXP_DIR $BASE_DIR/results/${EXP_DIR}_$RUN
            echo "finished $EXP_DIR"
            $HELPER_SCRIPT stop-machines --base-dir=$BASE_DIR
            sleep 60
        done
    done
done

# $HELPER_SCRIPT stop-machines --base-dir=$BASE_DIR