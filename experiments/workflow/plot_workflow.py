#!/usr/bin/python3
import os
import parse
import argparse
import numpy as np
import matplotlib.pyplot as plt


def summary(baseline, exp_name, run):
    base_dir = os.path.join(os.path.dirname(os.path.realpath(__file__)), baseline, "results")
    run_dir = f"{exp_name}_{run}"
    exp_dir = os.path.join(base_dir, run_dir)
    with open(os.path.join(exp_dir, "latency.txt")) as f:
        lines = f.read().strip().split("\n")
        if len(lines) < 3:
            print(exp_dir)
            return
        line_p50, line_p99, line_avg = lines
        p50 = parse.parse("p50 latency: {:f} ms", line_p50)[0]
        p99 = parse.parse("p99 latency: {:f} ms", line_p99)[0]
        avg = parse.parse("avg latency: {:f} ms", line_avg)[0]
    return p50, p99, avg


def plot(data, tputs, figname):
    font_size = 20
    markersize = 10
    linewidth = 3
    plt.rc("font", **{"size": font_size})
    metrics = ["p50", "p99"]  # avg
    ylabels = ["Median latency (ms)", "99% latency (ms)"]
    fig, axs = plt.subplots(nrows=2, ncols=1, figsize=(10, 8))
    for ax in axs.flat:
        ax.get_yaxis().set_tick_params(direction="in")
        ax.get_xaxis().set_tick_params(direction="in", pad=8)
        ax.grid(True)
    axs[1].set_xlabel("Throughput (requests/s)", labelpad=8)
    for i, ax in enumerate(axs):
        ax.set_ylabel(ylabels[i], labelpad=8, fontsize=font_size)
    plt.subplots_adjust(hspace=0.18)

    ######################################## hotel
    labels = ["Boki", "HM-read", "HM-write", "Unsafe"]
    markers = ["^", "o", "d", "s"]
    colors = ["red", "lightsalmon", "lightcoral", "royalblue"]
    xticks = tputs

    # plot
    col = axs
    handles = None
    for i, metric in enumerate(metrics):
        curves = []
        for j, baseline in enumerate(labels):
            (curve,) = col[i].plot(
                xticks,
                data[baseline][metric],
                label=labels[j],
                marker=markers[j],
                markersize=markersize,
                markeredgecolor="k",
                color=colors[j],
                linestyle="--",
                linewidth=linewidth,
            )
            curves.append(curve)
        col[i].set_xticks(xticks)
        if handles is None:
            handles = curves
    # col[0].legend(handles=curves, handlelength=legend_length, ncol=len(baselines), loc='upper center', bbox_to_anchor=bbox_to_anchor, frameon=True, prop={'size':legend_size})
    # col[0].set_ylim(0,23)
    # col[1].set_ylim(0,35)

    ############################################ legend
    # label_order = ["Boki",  "Halfmoon-read", "Unsafe"]
    # handles = [global_curves[label] for label in label_order]
    legend_size = 20
    legend_length = 2
    bbox_to_anchor = (0.5, 1.0)
    fig.legend(
        handles=handles,
        labels=labels,
        handlelength=legend_length,
        ncol=len(labels),
        loc="upper center",
        bbox_to_anchor=bbox_to_anchor,
        frameon=True,
        prop={"size": legend_size},
    )
    fig_dir = os.path.join(os.path.dirname(os.path.realpath(__file__)), "figures")
    fig_path = os.path.join(fig_dir, figname)
    os.makedirs(os.path.dirname(fig_path), exist_ok=True)
    plt.savefig(os.path.join(fig_dir, figname), bbox_inches="tight", transparent=False)


if __name__ == "__main__":
    parser = argparse.ArgumentParser()
    parser.add_argument("--qps", nargs="+", type=int, default=[100, 200, 300])
    parser.add_argument("exp", metavar="experiment", type=str)
    parser.add_argument("run", metavar="run", type=int, default=0)
    args = parser.parse_args()
    run = args.run

    results = {}
    baselines = [
        ("baseline", None, "Unsafe"),
        ("boki", None, "Boki"),
        ("opt", "write", "HM-read"),
        ("opt", "read", "HM-write"),
    ]
    for baseline, logmode, name in baselines:
        result = {"p50": [], "p99": []}
        for qps in args.qps:
            exp_name = f"QPS{qps}"
            if logmode is not None:
                exp_name += f"_{logmode}"
            p50, p99, _ = summary(f"{baseline}-{args.exp}", exp_name, run)
            result["p50"].append(p50)
            result["p99"].append(p99)
        results[name] = result

    for baseline in results:
        print(f"'{baseline}' : {results[baseline]}")

    # plot(results, args.qps, f"{run}/{args.exp}.png")
