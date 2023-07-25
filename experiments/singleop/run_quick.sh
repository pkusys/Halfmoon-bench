#!/bin/bash
BASE_DIR=`realpath $(dirname $0)`

RUN=$1

cd $BASE_DIR
./boki/run_all.sh $RUN
sleep 10
./optimal/run_all.sh $RUN