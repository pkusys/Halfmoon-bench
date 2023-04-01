#!/bin/bash
BASE_DIR=`realpath $(dirname $0)`
ROOT_DIR=`realpath $BASE_DIR/../../..`


HELPER_SCRIPT=$ROOT_DIR/scripts/exp_helper

RUN=$1

Concurrency=(256)
LOGMODE=("none")
ReadKeys=(8)
WriteKeys=(8)

$HELPER_SCRIPT start-machines --base-dir=$BASE_DIR

for c in ${Concurrency[@]}; do
    for m in ${LOGMODE[@]}; do
        for rk in ${ReadKeys[@]}; do
            for wk in ${WriteKeys[@]}; do
                EXP_DIR=$BASE_DIR/results/c${c}_${m}_r${rk}_w${wk}
                rm -rf $EXP_DIR
                mkdir -p $EXP_DIR
                $BASE_DIR/run_once.sh $EXP_DIR $c $m $rk $wk 2>&1 | tee run.log 
                cp $BASE_DIR/docker-compose.yml $EXP_DIR/
                cp $BASE_DIR/docker-compose-generated.yml $EXP_DIR/
                cp $BASE_DIR/config.json $EXP_DIR/
                cp $BASE_DIR/nightcore_config.json $EXP_DIR/
                mv $EXP_DIR ${EXP_DIR}_$RUN
                echo "finished con${c}_skew${s}_rk${rk}_wk${wk}"
            done
        done
    done
done

$HELPER_SCRIPT stop-machines --base-dir=$BASE_DIR
