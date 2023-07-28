#!/bin/bash

set -exu

ROOT_DIR=`realpath $(dirname $0)/..`
# TAG=single-reorder
TAG=sosp-ae

# Use BuildKit as docker builder
export DOCKER_BUILDKIT=1

function build_halfmoon {
    docker build -t shengqipku/halfmoon:$TAG \
        -f $ROOT_DIR/dockerfiles/Dockerfile.Halfmoon \
        $ROOT_DIR/halfmoon
}

function build_halfmoon_bench {
    docker build -t shengqipku/halfmoon-bench:$TAG \
        -f $ROOT_DIR/dockerfiles/Dockerfile.Halfmoon-bench \
        $ROOT_DIR/workloads/workflow
}

function build {
    build_halfmoon
    build_halfmoon_bench
}

function push {
    docker push shengqipku/halfmoon:$TAG
    docker push shengqipku/halfmoon-bench:$TAG
}

case "$1" in
build)
    build
    ;;
push)
    push
    ;;
esac