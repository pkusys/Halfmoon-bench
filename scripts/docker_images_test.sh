#!/bin/bash

set -exu

ROOT_DIR=`realpath $(dirname $0)/..`
# TAG=single-reorder
TAG=workflow

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

function build_beldibench {
    cd $ROOT_DIR/workloads/workflow
    cd ./beldi && make hotel-baseline && cd ..
    cd ./beldi && make media-baseline && cd ..
    cd ./beldi && make rw-baseline && cd ..
    cd ./beldi && make singleop-baseline && cd ..
    # cd ./beldi && make hotel && cd ..
    # cd ./beldi && make media && cd ..
    cd ./beldi && make singleop && cd ..
    # cd ./optimal-beldi && make hotel && cd ..
    # cd ./optimal-beldi && make media && cd ..
    cd ./optimal-beldi && make singleop && cd ..
    cd ./boki && make hotel && cd ..
    cd ./boki && make media && cd ..
    cd ./boki && make rw && cd ..
    cd ./boki && make singleop && cd ..
    cd ./optimal && make hotel && cd ..
    cd ./optimal && make media && cd ..
    cd ./optimal && make rw && cd ..
    cd ./optimal && make singleop && cd ..
    cd ./optimal && make switching && cd ..
    docker build -t shengqipku/boki-beldibench:$TAG -f $ROOT_DIR/dockerfiles/Dockerfile.my-beldibench $ROOT_DIR
}

function update {
    # commit_dev
    build_local
    build_release
    # build_boki-retwis
    # build_my-retwis
    # build_rwbench
    build_beldibench
}

function push {
    docker push shengqipku/boki:$TAG
    # docker push shengqipku/boki-retwisbench:$TAG
    # docker push shengqipku/my-retwisbench:$TAG
    docker push shengqipku/boki-beldibench:$TAG
    # docker push shengqipku/boki-rwbench:$TAG
}

case "$1" in
push)
    push
    ;;
build)
    build_local
    ;;
update)
    update
    push
    ;;
esac
