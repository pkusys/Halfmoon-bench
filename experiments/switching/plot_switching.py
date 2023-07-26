#!/usr/bin/python3
import os
import parse
import argparse
import numpy as np
import matplotlib.pyplot as plt

def parse_trace(exp_name, sample_factor=1):
    # read the trace file and extract timestamp, latency, and mode
    transitions = []
    results = []
    base_dir = os.path.join(os.path.dirname(os.path.realpath(__file__)), "results")
    with open(os.path.join(base_dir,exp_name,'trace.txt'), 'r') as f:
        for line in f:
            if "BEGIN" in line or "END" in line:
                result = parse.parse("{timestamp:d}: {}", line)
                transitions.append(result['timestamp'] / 1e6)
                continue
            result = parse.parse("{timestamp:d}: {latency:f}ms {mode}", line)
            results.append((result['timestamp'] / 1e6, result['latency']))
            # timestamps.append(result['timestamp'] / 1000) # convert to ms
            # latencies.append(result['latency'])
            # modes.append(result['mode'])

    threshold = np.median([x[1] for x in results]) * 5
    filtered = list(filter(lambda x: x[1] < threshold, results))[::sample_factor]
    timestamps = [x[0] for x in filtered]
    latencies = [x[1] for x in filtered]
    
    return timestamps, latencies, transitions

def plot(exp_names, figname):
    font_size = 35
    text_size = 30
    annotate_size = 30
    plt.rc('font',**{'size': font_size})
    ylabel = "Median latency (ms)"
    fig, axs = plt.subplots(nrows=1, ncols=2, figsize=(20, 6))
    ylabel="Latency (ms)"
    xlabel="Time (s)"
    axs[0].set_ylabel(ylabel, labelpad=0)
    for i in range(len(axs)):
        # axs[i].spines['top'].set_visible(False)
        # axs[i].get_xaxis().set_tick_params(direction='in')
        # axs[i].get_yaxis().set_tick_params(direction='in')
        axs[i].set_xlabel(xlabel, labelpad=0)
        axs[i].set_ylim(0, 100)
        axs[i].set_xlim(0, 15)
        axs[i].text(2.5, 15, "HM-W.", horizontalalignment='center', verticalalignment='center', fontsize=text_size)
        axs[i].text(7.8, 15, "HM-R.", horizontalalignment='center', verticalalignment='center', fontsize=text_size)
        axs[i].text(12.5, 15, "HM-W.", horizontalalignment='center', verticalalignment='center', fontsize=text_size)
    plt.subplots_adjust(wspace=0.15)
    # plt.tight_layout()

    for i, exp in enumerate(exp_names):
        timestamps, latencies, transitions = parse_trace(exp)
        axs[i].scatter(timestamps, latencies, marker='.', s=1, color='red')
        for j in range(len(transitions))[:4]:
            if j%2 == 0:
                axs[i].axvline(x=transitions[j], color='blue', linestyle='--')
            else:
                axs[i].axvline(x=transitions[j], color='black', linestyle='--')
        axs[i].annotate(f"{(transitions[1]-transitions[0])*1000:.0f} ms", xy=(5, 70), xytext=(6, 80),
            arrowprops=dict(facecolor='black', arrowstyle='->'), fontsize=annotate_size)
        axs[i].annotate(f"{(transitions[3]-transitions[2])*1000:.0f} ms", xy=(10, 70), xytext=(11, 80),
            arrowprops=dict(facecolor='black', arrowstyle='->'), fontsize=annotate_size)

    fig_dir = os.path.join(os.path.dirname(os.path.realpath(__file__)), "figures")
    fig_path = os.path.join(fig_dir, figname)
    os.makedirs(os.path.dirname(fig_path), exist_ok=True)
    plt.savefig(
        os.path.join(fig_dir, figname), bbox_inches="tight", transparent=False
    )


if __name__ == "__main__":
    parser = argparse.ArgumentParser()
    parser.add_argument("--concurrency", nargs="+", type=int, default=[16, 40])
    parser.add_argument("--nops", type=int, default=10)
    parser.add_argument("--read-ratios", type=str, default="0.2,0.8")
    parser.add_argument("run", metavar="run", type=int, default=0)
    args = parser.parse_args()
    run = args.run
    
    exp_names = [f"con{con}_n{args.nops}_rr{args.read_ratios}_{run}" for con in args.concurrency]
    plot(exp_names, f"{run}/switching.png")