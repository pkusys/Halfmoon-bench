#!/bin/bash
BASE_DIR=`realpath $(dirname $0)`

RUN=$1

cd $BASE_DIR
./opt-hotel/run_all.sh $RUN
sleep 10
./opt-movie/run_all.sh $RUN
sleep 10
./baseline-hotel/run_all.sh $RUN
sleep 10
./baseline-movie/run_all.sh $RUN
sleep 10
./boki-hotel/run_all.sh $RUN
sleep 10
./boki-movie/run_all.sh $RUN

./plot_workflow.py --qps 100 300 500 700 -- hotel $RUN

./plot_workflow.py --qps 100 200 300 -- movie $RUN





# ./summary.py $RUN