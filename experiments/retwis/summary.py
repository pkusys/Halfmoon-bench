#! /usr/bin/env python3

import os
import sys
import itertools
import parse
import matplotlib.pyplot as plt
import numpy as np

run = sys.argv[1]
baseline = ["boki", "reorder"]
percentage = ["10,10,30,50", "5,5,40,50"]
skew = [0.75, 0.9]
nNotify = [4, 8]
concurrency = [384]

root = os.path.dirname(os.path.realpath(__file__))
expName = lambda c, s, n, p: f"con{c}_skew{s}_n{n}_p{p}_{run}"
expDir = lambda base, c, s, n, p: os.path.join(
    root, base, "results", expName(c, s, n, p)
)

figDir = os.path.join(root, "figs")
os.makedirs(figDir, exist_ok=True)

results = {}
for base in baseline:
    results[base] = {}
    for c, s, n, p in itertools.product(concurrency, skew, nNotify, percentage):
        results[base][c, s, n, p] = {}


def parseResultLog(logFile):
    result = {}
    for l in logFile.readlines():
        tput = parse.parse("Benchmark {},{:f} request per sec\n", l)
        if tput is not None:
            result["tput"] = tput[1]
            continue
        latency = parse.parse("Latency: median = {:f}ms, tail (p99) = {:f}ms\n", l)
        if latency is not None:
            for p, l in zip(["p50", "p99"], latency.fixed):
                if p in result:
                    result[p].append(l)
                else:
                    result[p] = [l]
    result["p50"] = np.array(result["p50"])
    result["p99"] = np.array(result["p99"])
    return result


def parseAll():
    for base, c, s, n, p in itertools.product(
        baseline, concurrency, skew, nNotify, percentage
    ):
        exp = expDir(base, c, s, n, p)
        with open(os.path.join(exp, "results.log"), "r") as f:
            results[base][c, s, n, p] = parseResultLog(f)


factors = {}


def calcFactor():
    for c, s, n, p in itertools.product(concurrency, skew, nNotify, percentage):
        factor = {}
        factor["tput"] = (
            results["reorder"][c, s, n, p]["tput"] / results["boki"][c, s, n, p]["tput"]
        )
        factor["p50"] = (
            results["boki"][c, s, n, p]["p50"] / results["reorder"][c, s, n, p]["p50"]
        )
        factor["p99"] = (
            results["boki"][c, s, n, p]["p99"] / results["reorder"][c, s, n, p]["p99"]
        )
        factors[c, s, n, p] = factor


width = 0.2
nBars = 5


def plotAll():
    for c, p in itertools.product(concurrency, percentage):
        figName = f"concurrency={c} percentage={p}"
        group = {}
        for s, n in itertools.product(skew, nNotify):
            groupName = f"skew{s}_n{n}"
            group[groupName] = factors[c, s, n, p]
        plotGroup(figName, group)


def plotGroup(figName, groups):
    x = np.arange(len(groups)) * (width * nBars / 2) * 3

    tput = [val["tput"] for val in groups.values()]
    plotSingle(x, 4, tput, "tput")

    labels = ["profile", "follow", "post", "postlist"]
    p50 = np.vstack([val["p50"] for val in groups.values()])
    plotMulti(x, [0, 1, 2, 3], p50.T, [l + "-p50" for l in labels])
    p99 = np.vstack([val["p99"] for val in groups.values()])
    plotMulti(x, [0, 1, 2, 3], p99.T, [l + "-p99" for l in labels])

    plt.xticks(x, groups.keys())
    plt.ylabel("factor")
    plt.legend(loc="upper left")
    plt.title(figName)
    plt.savefig(os.path.join(figDir, f"{figName}.png"))
    plt.clf()


def plotSingle(anchors, barIdx, data, label, hatch=None, alpha=None):
    plt.bar(
        anchors + width * (barIdx - (nBars - 1) / 2.0),
        data,
        width,
        label=label,
        hatch=hatch,
        alpha=alpha,
    )


def plotMulti(anchors, barIdxList, dataList, labelList, hatch=None, alpha=None):
    for barIdx, data, label in zip(barIdxList, dataList, labelList):
        plotSingle(anchors, barIdx, data, label, hatch, alpha)


if __name__ == "__main__":
    parseAll()
    calcFactor()
    plotAll()
