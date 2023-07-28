#!/usr/bin/python3
import os
import parse
import argparse
import numpy as np
from itertools import product
import matplotlib.pyplot as plt


def summary(baseline, exp_name, gc_interval, run, log_mode=None):
    base_dir = os.path.join(os.path.dirname(os.path.realpath(__file__)), baseline, "results")
    if log_mode is not None:
        run_dir = f"{exp_name}_{log_mode}_{run}"
    else:
        run_dir = f"{exp_name}_{run}"
    exp_dir = os.path.join(base_dir, run_dir)
    with open(os.path.join(exp_dir, f"storage_gc{gc_interval}.txt")) as f:
        line = f.read().strip().split("\n")[-1]
        storage = parse.parse("time average: total={:f}KB, {}, {}", line)[0]
    return storage

def plot(boki, hm_read, hm_write, read_ratios, figname):
    font_size = 30
    legend_size = 30
    legend_length = 2
    bbox_to_anchor = (0.5, 1.075)
    markersize = 16
    linewidth = 5
    # markers = ["^", "o", "d"]
    # colors = ["red", "lightsalmon", "lightcoral"]

    plt.rc("font", **{"size": font_size})
    plt.rc("pdf", fonttype=42)
    ylabel = "Storage (MB)"
    fig = plt.figure(figsize=(10, 6))
    ax = plt.gca()
    ax.get_yaxis().set_tick_params(direction="in")
    ax.get_xaxis().set_tick_params(direction="in", pad=8)
    ax.grid(True)
    ax.set_xlabel("Read ratio", labelpad=8)
    ax.set_xticks(read_ratios)
    ax.set_ylabel(ylabel, labelpad=8, fontsize=30)

    boki_curve = ax.plot(
        read_ratios,
        boki,
        linestyle=":",
        label="Boki",
        marker="^",
        color="red",
        markersize=markersize,
        linewidth=linewidth,
    )
    hm_read_curve = ax.plot(
        read_ratios,
        hm_read,
        linestyle="-",
        label="Halfmoon-read",
        marker="o",
        color="lightcoral",
        markersize=markersize,
        linewidth=linewidth,
    )
    hm_write_curve = ax.plot(
        read_ratios,
        hm_write,
        linestyle="-",
        label="Halfmoon-write",
        marker="d",
        color="lightsalmon",
        markersize=markersize,
        linewidth=linewidth,
    )
    fig.legend(
        handlelength=legend_length,
        ncol=3,
        loc="upper center",
        bbox_to_anchor=bbox_to_anchor,
        frameon=True,
        prop={"size": legend_size},
    )
    # factors = boki / np.minimum(hm_read, hm_write)
    # print(
    #     f"{np.min(factors)}-{np.max(factors)}", f"{1-1/np.min(factors)}-{1-1/np.max(factors)}"
    # )
    fig_dir = os.path.join(os.path.dirname(os.path.realpath(__file__)), "figures")
    fig_path = os.path.join(fig_dir, figname)
    os.makedirs(os.path.dirname(fig_path), exist_ok=True)
    plt.savefig(
        os.path.join(fig_dir, figname), bbox_inches="tight", transparent=False
    )

if __name__ == "__main__":
    parser = argparse.ArgumentParser()
    parser.add_argument("--read-ratios", nargs='+', type=float, default=[0.1, 0.5, 0.9])
    parser.add_argument("--qps", type=int, default=100)
    parser.add_argument("--value-size", nargs="+", type=int, default=[256])
    parser.add_argument("--gc-intervals", nargs="+", type=int, default=[10000])
    parser.add_argument("run", metavar="run", type=int, default=0)
    args = parser.parse_args()
    run = args.run

    for v, gc in product(args.value_size, args.gc_intervals):
        boki = []
        hm_read = []
        hm_write = []
        for rr in args.read_ratios:
            exp_name = f"ReadRatio{rr}_QPS{args.qps}_v{v}"
            storage = summary("boki", exp_name, gc, run)
            boki.append(storage)
            storage = summary("optimal", exp_name, gc, run, "write")
            hm_read.append(storage)
            storage = summary("optimal", exp_name, gc, run, "read")
            hm_write.append(storage)
        plot(boki, hm_read, hm_write, args.read_ratios, f"{run}/storage_overhead_V{v}_GC{gc}.png")
