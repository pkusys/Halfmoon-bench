#!/bin/bash
BASE_DIR=`realpath $(dirname $0)`
ROOT_DIR=`realpath $BASE_DIR/../../..`

RUN=$1

HELPER_SCRIPT=$ROOT_DIR/scripts/exp_helper

# $HELPER_SCRIPT start-machines --base-dir=$BASE_DIR

# Cons=(64 128 256)
# Percs=("10,90" "20,80" "40,60" "60,40" "80,20" "90,10")
# Skews=(1.1 1.2 1.5 2)

concurrency=(384)
skew=(0.75)
notifyUsers=(4)
Percs=("10,10,30,50")

for c in ${concurrency[@]}; do
    for s in ${skew[@]}; do
        for n in ${notifyUsers[@]}; do
            for p in ${Percs[@]}; do
                $HELPER_SCRIPT start-machines --base-dir=$BASE_DIR
                EXP_DIR=$BASE_DIR/results/con${c}_skew${s}_n${n}_p${p}
                rm -rf $EXP_DIR
                mkdir -p $EXP_DIR
                $BASE_DIR/run_once.sh $EXP_DIR $c $s $n $p 2>&1 | tee $EXP_DIR/run.log
                cp $BASE_DIR/docker-compose.yml $EXP_DIR/
                cp $BASE_DIR/docker-compose-generated.yml $EXP_DIR/
                cp $BASE_DIR/config.json $EXP_DIR/
                cp $BASE_DIR/nightcore_config.json $EXP_DIR/
                mv $EXP_DIR ${EXP_DIR}_$RUN
                echo "finished con${c}_skew${s}_n${n}_p${p}"
                $HELPER_SCRIPT stop-machines --base-dir=$BASE_DIR
            done
        done
    done
done

# $HELPER_SCRIPT stop-machines --base-dir=$BASE_DIR
