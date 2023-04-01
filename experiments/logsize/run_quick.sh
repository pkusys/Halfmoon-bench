#!/bin/bash
BASE_DIR=`realpath $(dirname $0)`

RUN=$1

cd $BASE_DIR
./beldi/run_all.sh $RUN 2>&1 >/dev/null &&
./boki/run_all.sh $RUN 2>&1 >/dev/null &&
./optimal/run_all.sh $RUN 2>&1 >/dev/null &&

wait

./summary.py $RUN