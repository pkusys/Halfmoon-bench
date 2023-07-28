#!/bin/bash
BASE_DIR=`realpath $(dirname $0)`

RUN=$1

cd $BASE_DIR
./baseline/run_all.sh $RUN
sleep 10
./boki/run_all.sh $RUN
sleep 10
./optimal/run_all.sh $RUN

./plot_singleop.py --qps 15 $RUN