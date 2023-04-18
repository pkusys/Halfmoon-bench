#!/bin/bash
BASE_DIR=`realpath $(dirname $0)`

# RUN=$1

cd $BASE_DIR
./baseline-hotel/run_all.sh 3
sleep 10
./boki-hotel/run_all.sh 3
sleep 10
./opt-hotel/run_all.sh 3
sleep 10
./baseline-movie/run_all.sh 3
sleep 10
./boki-movie/run_all.sh 3
sleep 10
./opt-movie/run_all.sh 3
sleep 10
./beldi-movie/run_all.sh 3
sleep 10
./opt-beldi-movie/run_all.sh 3
sleep 10
# ./beldi-hotel/run_all.sh 1
# sleep 10





# ./summary.py $RUN