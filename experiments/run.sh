#!/bin/bash
set -u

BASE_DIR=`realpath $(dirname $0)`

cd $BASE_DIR/singleop
./run_quick.sh 1 >run.log 2>&1
echo "Finished singleop"

cd $BASE_DIR/workflow
./run_quick.sh 1 >run.log 2>&1
echo "Finished workflow"

cd $BASE_DIR/overhead
./run_quick.sh 1 >run.log 2>&1
echo "Finished overhead"

cd $BASE_DIR/switching
./run_all.sh 1 >run.log 2>&1
echo "Finished switching"
