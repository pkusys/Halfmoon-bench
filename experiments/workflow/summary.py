#!/usr/bin/python3
import os
import sys
import json
import parse
import copy
import argparse
import numpy as np


def summary(baseline, exp_name, run, log_mode=None):
    base_dir = os.path.join(
        os.path.dirname(os.path.realpath(__file__)), baseline, "results"
    )
    if log_mode is not None:
        run_dir = f"{exp_name}_{log_mode}_{run}"
    else:
        run_dir = f"{exp_name}_{run}"
    exp_dir = os.path.join(base_dir, run_dir)
    result = {}
    with open(os.path.join(exp_dir, "latency.txt")) as f:
        for line in f.read().strip().split("\n"):
            op, p50, p99 = parse.parse("{} latency: p50={:f}ms, p99={:f}ms", line)
            result[op] = np.array((p50, p99))
    return result


if __name__ == "__main__":
    parser = argparse.ArgumentParser()
    parser.add_argument("runs", type=str, default=None, nargs="+")
    args = parser.parse_args()
    # baseline
    single_result = {
        "read": [],
        "write": [],
        "invoke": [],
        "runs": 0,
    }
    full_result = {
        "baseline": copy.deepcopy(single_result),
        "boki": copy.deepcopy(single_result),
        "read-optimal": copy.deepcopy(single_result),
        "write-optimal": copy.deepcopy(single_result),
    }

    exp_name = "QPS1"

    for run in args.runs:
        # baseline
        result = full_result["baseline"]
        latencies = summary("beldi", exp_name, run)
        for k, v in latencies.items():
            if result[k] == []:
                result[k] = v
                result["runs"] += 1
            else:
                result[k] += v
                result["runs"] += 1

        # boki
        result = full_result["boki"]
        latencies = summary("boki", exp_name, run)
        for k, v in latencies.items():
            if result[k] == []:
                result[k] = v
                result["runs"] = 1
            else:
                result[k] += v
                result["runs"] += 1

        # read-optimal
        result = full_result["read-optimal"]
        latencies = summary("optimal", exp_name, run, "write")
        for k, v in latencies.items():
            if result[k] == []:
                result[k] = v
                result["runs"] += 1
            else:
                result[k] += v
                result["runs"] += 1

        # write-optimal
        result = full_result["write-optimal"]
        latencies = summary("optimal", exp_name, run, "read")
        for k, v in latencies.items():
            if result[k] == []:
                result[k] = v
                result["runs"] += 1
            else:
                result[k] += v
                result["runs"] += 1

    for baseline, result in full_result.items():
        runs = result["runs"]
        del result["runs"]
        for op, lat in result.items():
            full_result[baseline][op] = lat / runs

    with open("summary.json", "w") as f:
        json.dump(full_result, f, indent=4)
