#!/bin/bash

set -exu

ROOT_DIR=`realpath $(dirname $0)/..`
# TAG=single-reorder
TAG=readlog

# Use BuildKit as docker builder
export DOCKER_BUILDKIT=1

function build_local {
    cd $ROOT_DIR/boki
    CXX=clang++ make -j $(nproc)
}

function build_release {
    docker build -t shengqipku/boki:$TAG -f $ROOT_DIR/dockerfiles/Dockerfile.boki-release $ROOT_DIR
}

function build_rwbench {
    cd $ROOT_DIR/workloads/rw/ && ./build.sh
    docker build -t shengqipku/boki-rwbench:$TAG -f $ROOT_DIR/dockerfiles/Dockerfile.rwbench $ROOT_DIR
}

function build_occbench {
    cd $ROOT_DIR/workloads/occ/ && ./build.sh
    docker build -t shengqipku/boki-occbench:$TAG -f $ROOT_DIR/dockerfiles/Dockerfile.occbench $ROOT_DIR
}

function build_boki-retwis {
    cd $ROOT_DIR/workloads/boki-retwis/ && ./build.sh
    docker build -t shengqipku/boki-retwisbench:$TAG -f $ROOT_DIR/dockerfiles/Dockerfile.boki-retwisbench $ROOT_DIR
}

function build_my-retwis {
    cd $ROOT_DIR/workloads/my-retwis/ && ./build.sh
    docker build -t shengqipku/my-retwisbench:$TAG -f $ROOT_DIR/dockerfiles/Dockerfile.my-retwisbench $ROOT_DIR
}

function update {
    # commit_dev
    build_local
    build_release
    # build_boki-retwis
    # build_my-retwis
    build_rwbench
}

function push {
    docker push shengqipku/boki:$TAG
    # docker push shengqipku/boki-retwisbench:$TAG
    # docker push shengqipku/my-retwisbench:$TAG
    docker push shengqipku/boki-rwbench:$TAG
}

case "$1" in
push)
    push
    ;;
update)
    update
    push
    ;;
esac
