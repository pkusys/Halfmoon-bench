#!/bin/bash

set -exu

ROOT_DIR=`realpath $(dirname $0)/..`
TAG=test-prefetch

# Use BuildKit as docker builder
export DOCKER_BUILDKIT=1

# function build_boki {
#     docker build -t shengqipku/boki-dev \
#         -f $ROOT_DIR/dockerfiles/Dockerfile.boki-dev \
#         $ROOT_DIR/boki
# }

# function commit_dev {
#     docker commit boki-dev shengqipku/boki-dev
# }

function build_local {
    cd $ROOT_DIR/boki
    CXX=clang++ make -j $(nproc)
}

function build_release {
    # cd $ROOT_DIR/dockerfiles
    docker build -t shengqipku/boki:$TAG -f $ROOT_DIR/dockerfiles/Dockerfile.boki-release $ROOT_DIR
    # docker build -t shengqipku/boki:$TAG \
    #     -f $ROOT_DIR/dockerfiles/Dockerfile.boki-release \
    #     $ROOT_DIR/boki
}

function build_rwbench {
    # cd $ROOT_DIR/dockerfiles
    cd $ROOT_DIR/workloads/rw/ && ./build.sh
    docker build -t shengqipku/boki-rwbench:$TAG -f $ROOT_DIR/dockerfiles/Dockerfile.rwbench $ROOT_DIR
    # docker build -t shengqipku/boki-rwbench:test-prefetch \
    #     -f $ROOT_DIR/dockerfiles/Dockerfile.rwbench $ROOT_DIR
        # $ROOT_DIR/workloads/rw
}

function update {
    # commit_dev
    build_local
    build_release
    build_rwbench
}

# function build {
#     # build_boki
#     build_rwbench
# }

function push {
    docker push shengqipku/boki:$TAG
    docker push shengqipku/boki-rwbench:$TAG
}

case "$1" in
build)
    build
    ;;
push)
    push
    ;;
update)
    update
    push
    ;;
esac
