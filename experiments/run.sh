#!/bin/bash
BASE_DIR=`realpath $(dirname $0)`
cd $BASE_DIR

cd $BASE_DIR/singleop
./run_quick.sh 1 >run.log 2>&1

echo "Finished singleop"

cd $BASE_DIR/workflow
./run_quick.sh 1 >run.log 2>&1

echo "Finished workflow"
