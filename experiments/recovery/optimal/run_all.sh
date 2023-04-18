#!/bin/bash
BASE_DIR=`realpath $(dirname $0)`
ROOT_DIR=`realpath $BASE_DIR/../../..`

BOKI_MACHINE_IAM=boki-ae-experiments
HELPER_SCRIPT=$ROOT_DIR/scripts/exp_helper

RUN=$1

# runtime overhead: vary QPS
# storage overhead: vary value size and gc interval
# QPS=(50 100 150)
QPS=(100)
NUM_OPS=(10)
# READ_RATIO=(0 1)
LOGMODE=("read" "write")
VALUE_SIZE=(256)
# FAIL_RATE=(0.1)
FAIL_RATE=(0.1 ` 0.2 0.3 0.4)

$HELPER_SCRIPT start-machines --base-dir=$BASE_DIR --instance-iam-role=$BOKI_MACHINE_IAM

for qps in ${QPS[@]}; do
    for ops in ${NUM_OPS[@]}; do
        for mode in ${LOGMODE[@]}; do
            for f in ${FAIL_RATE[@]}; do
                for v in ${VALUE_SIZE[@]}; do
                    if [ "$mode" == "read" ]; then
                        rr=0
                    else
                        rr=1
                    fi
                    EXP_DIR=ReadRatio${rr}_QPS${qps}_${mode}_v${v}_f${f}
                    if [ -d "$BASE_DIR/results/${EXP_DIR}_$run" ]; then
                        echo "finished ${EXP_DIR}"
                        continue
                    fi
                    # $HELPER_SCRIPT start-machines --base-dir=$BASE_DIR --instance-iam-role=$BOKI_MACHINE_IAM
                    $BASE_DIR/run_once.sh $EXP_DIR $qps $ops $rr $mode $v $f # 2>&1 | tee $BASE_DIR/run.log 
                    cp $BASE_DIR/docker-compose.yml $BASE_DIR/results/$EXP_DIR
                    cp $BASE_DIR/docker-compose-generated.yml $BASE_DIR/results/$EXP_DIR
                    cp $BASE_DIR/config.json $BASE_DIR/results/$EXP_DIR
                    cp $BASE_DIR/nightcore_config.json $BASE_DIR/results/$EXP_DIR
                    mv $BASE_DIR/results/$EXP_DIR $BASE_DIR/results/${EXP_DIR}_$RUN
                    echo "finished ${EXP_DIR}"
                    # $HELPER_SCRIPT stop-machines --base-dir=$BASE_DIR
                    sleep 60
                done
            done
        done
    done
done

$HELPER_SCRIPT stop-machines --base-dir=$BASE_DIR