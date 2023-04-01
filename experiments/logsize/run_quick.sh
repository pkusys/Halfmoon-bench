#!/bin/bash
BASE_DIR=`realpath $(dirname $0)`

RUN=$1

cd $BASE_DIR
./optimal/run_all.sh $RUN
sleep 10
./boki/run_all.sh $RUN
sleep 10
./beldi/run_all.sh $RUN # 2>&1 >/dev/null

./summary.py $RUN