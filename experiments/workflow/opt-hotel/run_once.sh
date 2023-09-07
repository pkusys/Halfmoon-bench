#!/bin/bash

set -u

BASE_DIR=`realpath $(dirname $0)`
ROOT_DIR=`realpath $BASE_DIR/../../..`

BENCH_IMAGE=shengqipku/halfmoon-bench:sosp-ae

STACK=halfmoon

AWS_REGION=ap-southeast-1

EXP_DIR=$BASE_DIR/results/$1
QPS=$2
LOGMODE=$3

HELPER_SCRIPT=$ROOT_DIR/scripts/exp_helper
WRK_DIR=/usr/local/bin

MANAGER_HOST=`$HELPER_SCRIPT get-docker-manager-host --base-dir=$BASE_DIR`
CLIENT_HOST=`$HELPER_SCRIPT get-client-host --base-dir=$BASE_DIR`
ENTRY_HOST=`$HELPER_SCRIPT get-service-host --base-dir=$BASE_DIR --service=boki-gateway`
ALL_HOSTS=`$HELPER_SCRIPT get-all-server-hosts --base-dir=$BASE_DIR`

$HELPER_SCRIPT generate-docker-compose --base-dir=$BASE_DIR
scp -q $BASE_DIR/docker-compose.yml $MANAGER_HOST:~
scp -q $BASE_DIR/docker-compose-generated.yml $MANAGER_HOST:~

ssh -q $MANAGER_HOST -- docker stack rm $STACK

sleep 40

TABLE_PREFIX=$(head -c 64 /dev/urandom | tr -dc 'a-zA-Z0-9' | fold -w 8 | head -n 1)
TABLE_PREFIX="${TABLE_PREFIX}-"

ssh -q $CLIENT_HOST -- docker pull $BENCH_IMAGE

ssh -q $CLIENT_HOST -- docker run -v /tmp:/tmp \
    $BENCH_IMAGE \
    cp -r /optimal-bin/hotel /tmp/

ssh -q $CLIENT_HOST -- TABLE_PREFIX=$TABLE_PREFIX AWS_REGION=$AWS_REGION LoggingMode=$LOGMODE \
    /tmp/hotel/init create cayon
ssh -q $CLIENT_HOST -- TABLE_PREFIX=$TABLE_PREFIX AWS_REGION=$AWS_REGION LoggingMode=$LOGMODE \
    /tmp/hotel/init populate cayon

scp -q $ROOT_DIR/scripts/zk_setup.sh $MANAGER_HOST:/tmp/zk_setup.sh
ssh -q $MANAGER_HOST -- sudo mkdir -p /mnt/inmem/store

for host in $ALL_HOSTS; do
    scp -q $BASE_DIR/nightcore_config.json $host:/tmp/nightcore_config.json
done

ALL_ENGINE_HOSTS=`$HELPER_SCRIPT get-machine-with-label --base-dir=$BASE_DIR --machine-label=engine_node`
for HOST in $ALL_ENGINE_HOSTS; do
    scp -q $BASE_DIR/run_launcher $HOST:/tmp/run_launcher
    ssh -q $HOST -- sudo rm -rf /mnt/inmem/boki
    ssh -q $HOST -- sudo mkdir -p /mnt/inmem/boki
    ssh -q $HOST -- sudo mkdir -p /mnt/inmem/boki/output /mnt/inmem/boki/ipc
    ssh -q $HOST -- sudo cp /tmp/run_launcher /mnt/inmem/boki/run_launcher
    ssh -q $HOST -- sudo cp /tmp/nightcore_config.json /mnt/inmem/boki/func_config.json
done

ALL_STORAGE_HOSTS=`$HELPER_SCRIPT get-machine-with-label --base-dir=$BASE_DIR --machine-label=storage_node`
for HOST in $ALL_STORAGE_HOSTS; do
    ssh -q $HOST -- sudo rm -rf   /mnt/storage/logdata
    ssh -q $HOST -- sudo mkdir -p /mnt/storage/logdata
done

ssh -q $MANAGER_HOST -- TABLE_PREFIX=$TABLE_PREFIX LoggingMode=$LOGMODE docker stack deploy \
    -c ~/docker-compose-generated.yml -c ~/docker-compose.yml $STACK
sleep 100

for HOST in $ALL_ENGINE_HOSTS; do
    ENGINE_CONTAINER_ID=`$HELPER_SCRIPT get-container-id --base-dir=$BASE_DIR --service boki-engine --machine-host $HOST`
    echo 4096 | ssh -q $HOST -- sudo tee /sys/fs/cgroup/cpu,cpuacct/docker/$ENGINE_CONTAINER_ID/cpu.shares
done
sleep 10

rm -rf $EXP_DIR
mkdir -p $EXP_DIR

ssh -q $MANAGER_HOST -- cat /proc/cmdline >>$EXP_DIR/kernel_cmdline
ssh -q $MANAGER_HOST -- uname -a >>$EXP_DIR/kernel_version

scp -q $ROOT_DIR/workloads/workflow/optimal/benchmark/hotel/workload.lua $CLIENT_HOST:/tmp

ssh -q $CLIENT_HOST -- $WRK_DIR/wrk -t 2 -c 2 -d 30 -L -U \
    -s /tmp/workload.lua \
    http://$ENTRY_HOST:8080 -R $QPS >$EXP_DIR/wrk_warmup.log

sleep 10

ssh -q $CLIENT_HOST -- $WRK_DIR/wrk -t 2 -c 2 -d 150 -L -U \
    -s /tmp/workload.lua \
    http://$ENTRY_HOST:8080 -R $QPS 2>/dev/null >$EXP_DIR/wrk.log

sleep 10

scp -q $MANAGER_HOST:/mnt/inmem/store/async_results $EXP_DIR

if [ -s "$EXP_DIR/async_results" ]; then
    $ROOT_DIR/scripts/compute_latency.py --async-result-file $EXP_DIR/async_results >$EXP_DIR/latency.txt
fi

ssh -q $CLIENT_HOST -- TABLE_PREFIX=$TABLE_PREFIX AWS_REGION=$AWS_REGION LoggingMode=$LOGMODE \
    /tmp/hotel/init clean cayon

if [ ! -s "$EXP_DIR/async_results" ]; then
    $HELPER_SCRIPT collect-container-logs --base-dir=$BASE_DIR --log-path=$EXP_DIR
fi
# $HELPER_SCRIPT collect-container-logs --base-dir=$BASE_DIR --log-path=$EXP_DIR
