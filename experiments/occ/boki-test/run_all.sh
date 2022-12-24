#!/bin/bash
BASE_DIR=`realpath $(dirname $0)`
ROOT_DIR=`realpath $BASE_DIR/../../..`

RUN=$1

HELPER_SCRIPT=$ROOT_DIR/scripts/exp_helper

$HELPER_SCRIPT start-machines --base-dir=$BASE_DIR

# Cons=(64 128 256)
# Percs=("10,90" "20,80" "40,60" "60,40" "80,20" "90,10")
# Skews=(1.1 1.2 1.5 2)

Cons=(256)
RKeys=(8)
RWRatios=(0.125)
Skews=(1.1)

for c in ${Cons[@]}; do
    for rk in ${RKeys[@]}; do
        for rw in ${RWRatios[@]}; do
            for s in ${Skews[@]}; do
                EXP_DIR=$BASE_DIR/results/con${c}_rk${rk}_rw${rw}_skew${s}
                rm -rf $EXP_DIR
                mkdir -p $EXP_DIR
                $BASE_DIR/run_once.sh $c $rk $rw $s $EXP_DIR 2>&1 | tee $EXP_DIR/run.log
                cp $BASE_DIR/docker-compose.yml $EXP_DIR/
                mv $EXP_DIR ${EXP_DIR}_$RUN
                echo "finished con${c}_rk${rk}_rw${rw}_skew${s}"
            done
        done
    done
done

# $HELPER_SCRIPT stop-machines --base-dir=$BASE_DIR
