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


def filter_json(json_results, warmup_ratio=1.0 / 6, outlier_ratio=30):
    refined_results = []
    skip = int(len(json_results) * warmup_ratio)
    for entry in json_results[skip:]:
        start = entry["dispatchTs"]
        end = entry["finishedTs"]
        latency = end - start
        output_base64 = entry["output"]
        output_raw = base64.b64decode(output_base64)
        output_json = json.loads(output_raw)
        logsize = output_json["LogSize"]
        writeset = []
        if output_json["Output"] is not None:
            writeset = output_json["Output"]
        refined_results.append((latency, start, end, logsize, writeset))
    threshold = np.median([r[0] for r in refined_results]) * outlier_ratio
    filtered = list(filter(lambda x: x[0] < threshold, refined_results))
    filtered.sort(key=lambda x: x[1])
    return filtered


def time_average_logsize(filtered_results):
    logsize_sum = 0
    start = np.min([r[1] for r in filtered_results])
    end = np.max([r[2] for r in filtered_results])
    duration = (end - start) / 1000.0
    for latency, _, _, logsize in filtered_results:
        logsize_sum += logsize * latency / 1000.0
    return logsize_sum / duration


def build_write_history(filtered_results):
    write_history = {}
    for _, start, _, _, writeset in filtered_results:
        for key in writeset:
            if key not in write_history:
                write_history[key] = []
            write_history[key].append(start)
    end = np.max([r[2] for r in filtered_results])
    for key in write_history:
        write_history[key].append(end)
    return write_history


def time_average_writesize(filtered_results, write_history, value_size):
    writesize_sum = 0
    # global_end = np.max([r[2] for r in filtered_results])
    for key in write_history:
        for i in range(len(write_history[key]) - 1):
            start = write_history[key][i]
            next_write = write_history[key][i + 1]
            readers = filter(
                lambda x: x[1] >= start and x[1] <= next_write, filtered_results
            )
            end = np.max([r[2] for r in readers])
            duration = (end - start) / 1000.0
            writesize_sum += value_size * duration

    start = np.min([r[1] for r in filtered_results])
    end = np.max([r[2] for r in filtered_results])
    duration = (end - start) / 1000.0
    return writesize_sum / duration


if __name__ == "__main__":
    parser = argparse.ArgumentParser()
    parser.add_argument("--async-result-file", type=str, default=None)
    parser.add_argument("--warmup-ratio", type=float, default=1.0 / 6)
    parser.add_argument("--outlier-factor", type=int, default=30)
    parser.add_argument("--value-size", type=int, default=256)
    args = parser.parse_args()

    json_results = parse_json(args.async_result_file)
    filtered_results = filter_json(json_results, args.warmup_ratio, args.outlier_factor)
    log_size = time_average_logsize(filtered_results) / 1024
    write_history = build_write_history(filtered_results)
    write_size = (
        time_average_writesize(filtered_results, write_history, args.value_size) / 1024
    )
    print(
        "time average: total=%dB, log=%dB, write=%dB"
        % (log_size + write_size, log_size, write_size)
    )
