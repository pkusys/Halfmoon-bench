#!/bin/bash
BASE_DIR=`realpath $(dirname $0)`

RUN=$1

cd $BASE_DIR
./optimal/run_all_runtime.sh $RUN
sleep 10
./optimal/run_all_storage.sh $RUN
sleep 10
./optimal/run_all_runtime.sh $RUN
sleep 10
./optimal/run_all_storage.sh $RUN
# sleep 10
# ./beldi/run_all.sh $RUN # 2>&1 >/dev/null

# ./summary.py $RUN