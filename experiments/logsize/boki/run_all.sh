#!/bin/bash
BASE_DIR=`realpath $(dirname $0)`
ROOT_DIR=`realpath $BASE_DIR/../../..`

BOKI_MACHINE_IAM=boki-ae-experiments
HELPER_SCRIPT=$ROOT_DIR/scripts/exp_helper

RUN=$1

QPS=(100 300 400)
NUM_OPS=(10)
READ_RATIO=(0.1 0.3 0.5 0.7 0.9)
VALUE_SIZE=(256 1024)

$HELPER_SCRIPT start-machines --base-dir=$BASE_DIR --instance-iam-role=$BOKI_MACHINE_IAM

for qps in ${QPS[@]}; do
    for ops in ${NUM_OPS[@]}; do
        for rr in ${READ_RATIO[@]}; do
            for v in ${VALUE_SIZE[@]}; do
                if [ $qps -ne 100 ] && [ $v -ne 256 ]; then
                    continue
                fi
                if [ $qps -eq 100 ] && [ $v -eq 256 ]; then
                    continue
                fi
                EXP_DIR=ReadRatio${rr}_QPS${qps}_v${v}
                if [ -d "$BASE_DIR/results/${EXP_DIR}_$run" ]; then
                    echo "finished ReadRatio${rr}_QPS${qps}_v${v}"
                    continue
                fi
                # $HELPER_SCRIPT start-machines --base-dir=$BASE_DIR --instance-iam-role=$BOKI_MACHINE_IAM
                $BASE_DIR/run_once.sh $EXP_DIR $qps $ops $rr $v # 2>&1 | tee $BASE_DIR/run.log 
                cp $BASE_DIR/docker-compose.yml $BASE_DIR/results/$EXP_DIR
                cp $BASE_DIR/docker-compose-generated.yml $BASE_DIR/results/$EXP_DIR
                cp $BASE_DIR/config.json $BASE_DIR/results/$EXP_DIR
                cp $BASE_DIR/nightcore_config.json $BASE_DIR/results/$EXP_DIR
                mv $BASE_DIR/results/$EXP_DIR $BASE_DIR/results/${EXP_DIR}_$RUN
                echo "finished ReadRatio${rr}_QPS${qps}_v${v}"
                # $HELPER_SCRIPT stop-machines --base-dir=$BASE_DIR
                sleep 60
            done
        done
    done
done

$HELPER_SCRIPT stop-machines --base-dir=$BASE_DIR

