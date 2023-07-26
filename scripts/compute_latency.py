#!/usr/bin/python3
import os
import json
import argparse

import numpy as np


def parse_async_results(file_path):
    results = []
    with open(file_path) as fin:
        for line in fin:
            results.append(json.loads(line.strip()))
    return results


def compute_latency(async_result_file_path, warmup_ratio=1.0 / 6, outlier_ratio=30, sleep_duration=0):
    # queueing_delays = []
    latencies = []
    results = parse_async_results(async_result_file_path)
    skip = int(len(results) * warmup_ratio)
    for entry in results[skip:]:
        recv_ts = entry["recvTs"]
        dispatch_ts = entry["dispatchTs"]
        finish_ts = entry["finishedTs"]
        # if dispatch_ts > recv_ts:
        #     queueing_delays.append((dispatch_ts - recv_ts) / 1000.0)
        latencies.append((finish_ts - dispatch_ts) / 1000.0 - sleep_duration)
    threshold = np.median(latencies) * outlier_ratio
    filtered = list(filter(lambda x: x < threshold, latencies))
    p50 = np.percentile(filtered, 50)
    p99 = np.percentile(filtered, 99)
    avg = np.mean(filtered)
    return p50, p99, avg


if __name__ == "__main__":
    parser = argparse.ArgumentParser()
    parser.add_argument("--async-result-file", type=str, default=None)
    parser.add_argument("--warmup-ratio", type=float, default=1.0 / 6)
    parser.add_argument("--outlier-factor", type=int, default=20)
    parser.add_argument("--sleep-duration", type=int, default=0) # in ms
    args = parser.parse_args()

    p50, p99, avg = compute_latency(
        args.async_result_file,
        warmup_ratio=args.warmup_ratio,
        outlier_ratio=args.outlier_factor,
        sleep_duration=args.sleep_duration,
    )
    print("p50 latency: %.2f ms" % p50)
    print("p99 latency: %.2f ms" % p99)
    print("avg latency: %.2f ms" % avg)
