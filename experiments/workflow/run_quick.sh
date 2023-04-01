#!/bin/bash
BASE_DIR=`realpath $(dirname $0)`

RUN=$1

cd $BASE_DIR
./beldi-hotel/run_all.sh $RUN # 2>&1 >/dev/null
sleep 10
./beldi-movie/run_all.sh $RUN # 2>&1 >/dev/null
sleep 10
./boki-hotel/run_all.sh $RUN
sleep 10
./boki-movie/run_all.sh $RUN
sleep 10
./opt-hotel/run_all.sh $RUN
sleep 10
./opt-movie/run_all.sh $RUN

# ./summary.py $RUN