#!/bin/bash
BASE_DIR=`realpath $(dirname $0)`
cd $BASE_DIR

cd $BASE_DIR/workflow
./run_quick.sh > workflow.log 2>&1

echo "Finished workflow"

cd $BASE_DIR/singleop
./run_quick.sh > singleop.log 2>&1

echo "Finished singleop"

cd $BASE_DIR/logsize
./optimal/run_all.sh 2 > logsize.log 2>&1
