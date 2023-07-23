#!/bin/bash
BASE_DIR=`realpath $(dirname $0)`

RUN=$1

cd $BASE_DIR
./baseline-hotel/run_all.sh $RUN
sleep 10
./boki-hotel/run_all.sh $RUN
sleep 10
./opt-hotel/run_all.sh $RUN
sleep 10
./baseline-movie/run_all.sh $RUN
sleep 10
./boki-movie/run_all.sh $RUN
sleep 10
./opt-movie/run_all.sh $RUN





# ./summary.py $RUN