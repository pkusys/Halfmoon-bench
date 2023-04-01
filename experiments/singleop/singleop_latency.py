#!/usr/bin/python3
import base64
import json
import argparse

import numpy as np


def parse_json(file_path):
    results = []
    with open(file_path) as fin:
        for line in fin:
            results.append(json.loads(line.strip()))
    return results


def compute_latencies(json_results, warmup_ratio=1.0 / 6, outlier_ratio=30):
    refined_results = []
    skip = int(len(json_results) * warmup_ratio)
    for entry in json_results[skip:]:
        start = entry["dispatchTs"]
        end = entry["finishedTs"]
        latency = end - start
        output_base64 = entry["output"]
        output_raw = base64.b64decode(output_base64)
        output_json = json.loads(output_raw)
        latency_read = output_json["Output"]["Read"]
        latency_write = output_json["Output"]["Write"]
        latency_invoke = output_json["Output"]["Invoke"]
        refined_results.append(
            (
                latency,
                latency_read,
                latency_write,
                latency_invoke,
            )
        )
    threshold = np.median([r[0] for r in refined_results]) * outlier_ratio
    filtered = list(filter(lambda x: x[0] < threshold, refined_results))
    # filtered.sort(key=lambda x: x[1])
    for i, op_name in enumerate(["read", "write", "invoke"], 1):
        p50 = np.percentile([x[i] for x in filtered], 50) / 1000.0
        p99 = np.percentile([x[i] for x in filtered], 99) / 1000.0
        print(f"{op_name} latency: p50={p50:.2f}ms p99={p99:.2f}ms")
    # return filtered


if __name__ == "__main__":
    parser = argparse.ArgumentParser()
    parser.add_argument("--async-result-file", type=str, default=None)
    parser.add_argument("--warmup-ratio", type=float, default=1.0 / 6)
    parser.add_argument("--outlier-factor", type=int, default=30)
    args = parser.parse_args()

    json_results = parse_json(args.async_result_file)
    compute_latencies(
        json_results, warmup_ratio=args.warmup_ratio, outlier_ratio=args.outlier_factor
    )
