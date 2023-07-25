#!/bin/bash
BASE_DIR=`realpath $(dirname $0)`

RUN=$1

cd $BASE_DIR
./boki/run_all_runtime.sh $RUN
sleep 10
./boki/run_all_storage.sh $RUN
sleep 10
./optimal/run_all_runtime.sh $RUN
sleep 10
./optimal/run_all_storage.sh $RUN