#!/bin/bash
BASE_DIR=`realpath $(dirname $0)`
ROOT_DIR=`realpath $BASE_DIR/../../..`

# RUN=$1

HELPER_SCRIPT=$ROOT_DIR/scripts/exp_helper

$HELPER_SCRIPT start-machines --base-dir=$BASE_DIR

Cons=(64 128 256)
Percs=("10,90" "30,70" "50,50" "70,30" "90,10")
Skews=(1.1 1.2 1.5)

# Cons=(256)
# Percs=("90,10")
# Skews=(1.5)

RUN=0

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

RUN=1

sed -i '/slog_engine_cache_prefetch/s/# -/-/g' $BASE_DIR/docker-compose.yml

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

sed -i '/slog_engine_cache_prefetch/s/-/# -/g' $BASE_DIR/docker-compose.yml

$HELPER_SCRIPT stop-machines --base-dir=$BASE_DIR
