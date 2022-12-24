#! /usr/bin/env python3
import sys
import os

root = os.path.dirname(__file__)
logDir = os.path.join(root, sys.argv[1], "logs")
logFile = [f for f in os.listdir(logDir) if "gateway" in f][0]

# logFile = "boki_boki-gateway.1.9nufqxxl3kd2pl2mgt8qi0fcb.stderr"
print(logFile)
pending = {}
with open(os.path.join(logDir, logFile), "r") as f:
    for n, line in enumerate(f, 1):
        if "OnNewHttpFuncCall" in line:
            funcId = line.split(",")[0].strip().split("=")[-1]
            callId = line.split(",")[-1].strip().split("=")[-1]
            pending[callId] = f"new func {funcId}"
        if "dispatch sync call_id" in line:
            parsed = line.split("call_id", 1)[-1].strip().split()
            callId = parsed[0]
            engine = parsed[-1]
            assert callId in pending
            funcId = pending[callId].split(" ")[-1]
            pending[callId] = f"func {funcId} dispatched to {engine}"
        elif "OnFuncCallCompleted" in line:
            callId = line.split(",")[-1].strip().split("=")[-1]
            assert callId in pending
            pending.pop(callId)
for callId, status in pending.items():
    print(callId, status)
