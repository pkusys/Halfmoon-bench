#!/usr/bin/python3
import os
import sys
import json
import parse
import copy


def summary(baseline, exp_name, run, log_mode=None):
    base_dir = os.path.join(
        os.path.dirname(os.path.realpath(__file__)), baseline, "results"
    )
    if log_mode is not None:
        run_dir = f"{exp_name}_{log_mode}_{run}"
    else:
        run_dir = f"{exp_name}_{run}"
    exp_dir = os.path.join(base_dir, run_dir)
    with open(os.path.join(exp_dir, "latency.txt")) as f:
        line_p50, line_p99 = f.read().strip().split("\n")
        p50 = parse.parse("p50 latency: {:f} ms", line_p50)[0]
        p99 = parse.parse("p99 latency: {:f} ms", line_p99)[0]
    if not os.path.exists(os.path.join(exp_dir, "logsize.txt")):
        return p50, p99, None
    with open(os.path.join(exp_dir, "logsize.txt")) as f:
        line = f.read().strip()
        logsize = parse.parse("time average: total={:d}B, {}, {}", line)[0]
    return p50, p99, logsize


read_ratios = [0.1, 0.3, 0.5, 0.7, 0.9]

if __name__ == "__main__":
    run = sys.argv[1]

    # baseline
    single_result = {
        "p50": [],
        "p99": [],
        "logsize": [],
    }
    full_result = {
        "baseline": copy.deepcopy(single_result),
        "boki": copy.deepcopy(single_result),
        "read-optimal": copy.deepcopy(single_result),
        "write-optimal": copy.deepcopy(single_result),
    }

    # baseline
    result = full_result["baseline"]
    for read_ratio in read_ratios:
        exp_name = f"ReadRatio{read_ratio}"
        p50, p99, _ = summary("beldi", exp_name, run)
        result["p50"].append(p50)
        result["p99"].append(p99)

    # boki
    result = full_result["boki"]
    for read_ratio in read_ratios:
        exp_name = f"ReadRatio{read_ratio}"
        p50, p99, logsize = summary("boki", exp_name, run)
        result["p50"].append(p50)
        result["p99"].append(p99)
        result["logsize"].append(logsize)

    # read-optimal
    result = full_result["read-optimal"]
    for read_ratio in read_ratios:
        exp_name = f"ReadRatio{read_ratio}"
        p50, p99, logsize = summary("optimal", exp_name, run, "write")
        result["p50"].append(p50)
        result["p99"].append(p99)
        result["logsize"].append(logsize)

    # write-optimal
    result = full_result["write-optimal"]
    for read_ratio in read_ratios:
        exp_name = f"ReadRatio{read_ratio}"
        p50, p99, logsize = summary("optimal", exp_name, run, "read")
        result["p50"].append(p50)
        result["p99"].append(p99)
        result["logsize"].append(logsize)

    with open("summary.json", "w") as f:
        json.dump(full_result, f, indent=4)
