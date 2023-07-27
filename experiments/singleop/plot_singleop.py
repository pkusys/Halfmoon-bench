#!/usr/bin/python3
import os
import parse
import argparse
import numpy as np
import matplotlib.pyplot as plt


def summary(baseline, exp_name, run, log_mode=None):
    base_dir = os.path.join(os.path.dirname(os.path.realpath(__file__)), baseline, "results")
    if log_mode is not None:
        run_dir = f"{exp_name}_{log_mode}_{run}"
    else:
        run_dir = f"{exp_name}_{run}"
    exp_dir = os.path.join(base_dir, run_dir)
    with open(os.path.join(exp_dir, "latency.txt")) as f:
        lines = f.read().strip().split("\n")
        read_p50, read_p99 = parse.parse("read latency: p50={:f}ms p99={:f}ms", lines[0])
        write_p50, write_p99 = parse.parse("write latency: p50={:f}ms p99={:f}ms", lines[1])
    return read_p50, read_p99, write_p50, write_p99

def plot(read_p50, read_p99, write_p50, write_p99, figname):
    font_size = 18
    plt.rc('font',**{'size': font_size})
    fig_size = (10, 4)
    fig, ax = plt.subplots(figsize=fig_size, ncols=2)
    width=0.4
    xticks=np.arange(0, 4, 1)+1
    xlabels = ["Raw","Boki","HM-R.","HM-W."]
    colors = ["royalblue","red","lightsalmon","lightcoral"]
    
    # read
    ax[0].bar(xticks, read_p50[:4], color=colors, width=width, align='center')
    for i in range(len(xticks)):
        ax[0].errorbar(xticks[i], read_p50[i], yerr=read_p99[i]-read_p50[i], fmt='o', color=colors[i], ecolor=colors[i], capsize=5, zorder=0, markersize=5)
    # write
    ax[1].bar(xticks, write_p50[:4], color=colors, width=width, align='center')
    for i in range(len(xticks)):
        ax[1].errorbar(xticks[i], write_p50[i], yerr=write_p99[i]-write_p50[i], fmt='o', color=colors[i], ecolor=colors[i], capsize=5, zorder=0, markersize=5)
    
    # ticks
    ylabel="Latency (ms)"
    ax[0].set_ylabel(ylabel, labelpad=8)
    for i in range(len(ax)):
        ax[i].spines['top'].set_visible(False)
        ax[i].spines['right'].set_visible(False)
        ax[i].yaxis.set_ticks_position('left')
        ax[i].set_ylim(bottom=0)
        # ax.set_yticks(y_ticks)
        ax[i].get_yaxis().set_tick_params(direction='in', pad=5)
        ax[i].set_xticks(xticks)
        # ax[i].set_xticklabels(xlabels, rotation=45, ha="right")
        ax[i].set_xticklabels(xlabels)
        ax[i].xaxis.set_ticks_position('bottom')
    fig.subplots_adjust(wspace=0.2)
    
    fig_dir = os.path.join(os.path.dirname(os.path.realpath(__file__)), "figures")
    fig_path = os.path.join(fig_dir, figname)
    os.makedirs(os.path.dirname(fig_path), exist_ok=True)
    plt.savefig(
        os.path.join(fig_dir, figname), bbox_inches="tight", transparent=False
    )

if __name__ == "__main__":
    parser = argparse.ArgumentParser()
    parser.add_argument("--qps", type=int, default=20)
    parser.add_argument("run", metavar="run", type=int, default=0)
    args = parser.parse_args()
    run = args.run

    read_p50 = []
    read_p99 = []
    write_p50 = []
    write_p99 = []
    params = [("baseline",None),("boki",None),("optimal","write"),("optimal","read")]
    for baseline, log_mode in params:
        r50,r99,w50,w99 = summary(baseline, f"QPS{args.qps}", run, log_mode)
        read_p50.append(r50)
        read_p99.append(r99)
        write_p50.append(w50)
        write_p99.append(w99)
    plot(read_p50, read_p99, write_p50, write_p99, f"{run}/microbenchmarks.png")
