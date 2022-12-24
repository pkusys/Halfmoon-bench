#! /bin/bash
set -exu

BASE_DIR=`realpath $(dirname $0)`
ROOT_DIR=`realpath $BASE_DIR/..`

cd $ROOT_DIR/experiments/retwis/boki

./run_all.sh $1

cd $ROOT_DIR/experiments/retwis/reorder

./run_all.sh $1
