#!/bin/bash
BASE_DIR=`realpath $(dirname $0)`
ROOT_DIR=`realpath $BASE_DIR/../../..`

# pre+sync pre+sync+!setall pre+nosync 

RUN=$1

HELPER_SCRIPT=$ROOT_DIR/scripts/exp_helper

$HELPER_SCRIPT start-machines --base-dir=$BASE_DIR

# Cons=(64 128 256)
# Percs=("10,90" "20,80" "40,60" "60,40" "80,20" "90,10")
# Skews=(1.1 1.2 1.5 2)

Cons=(64)
Percs=("10,90")
Skews=(1.1)

for c in ${Cons[@]}; do
    for p in ${Percs[@]}; do
        for s in ${Skews[@]}; do
            EXP_DIR=$BASE_DIR/results/con${c}_rw${p}_skew${s}
            $BASE_DIR/run_once.sh $c $p $s $EXP_DIR
            cp $BASE_DIR/docker-compose.yml $EXP_DIR/
            mv $EXP_DIR ${EXP_DIR}_$RUN
            echo "finished con${c}_rw${p}_skew${s}"
        done
    done
done

# $HELPER_SCRIPT stop-machines --base-dir=$BASE_DIR
