#! /usr/bin/env python3

import os
import itertools
import parse
import matplotlib.pyplot as plt
import numpy as np


baseDir = os.path.dirname(os.path.realpath(__file__))

concurrency = [64, 128, 256]
ratio = [(10, 90), (30, 70), (50, 50), (70, 30), (90, 10)]
skew = [1.1, 1.2, 1.5]

# concurrency = [64]
# ratio = [(10, 90)]
# skew = [1.1]


experiment = lambda c, r, s, prefetch: f"con{c}_rw{r[0]},{r[1]}_skew{s}_{int(prefetch)}"

# print(experiment(64, (50, 50), 1.1, False))

results = {}

for c, r, s, prefetch in itertools.product(concurrency, ratio, skew, (False, True)):
    exp = os.path.join(baseDir, "results", experiment(c, r, s, prefetch), "results.log")
    with open(exp, "r") as f:
        result = {}
        for l in f.readlines():
            if "Benchmark" in l:
                tput = parse.parse("Benchmark {},{:f} request per sec\n", l)[1]
                result["tput"] = tput
            latency = parse.parse("Latency: median = {:f}ms, tail (p99) = {:f}ms\n", l)
            if latency is not None:
                p50, p99 = latency.fixed
                if "lat" in result:
                    result["lat"] += (p50, p99)
                else:
                    result["lat"] = (p50, p99)
    results[(c, r, s, prefetch)] = result

groups = {}
for c, s in itertools.product(concurrency, skew):
    factors = []
    for r in ratio:
        tput_factor = (
            results[(c, r, s, True)]["tput"] / results[(c, r, s, False)]["tput"]
        )
        lat_factor = []
        for lat1, lat2 in zip(
            results[(c, r, s, True)]["lat"], results[(c, r, s, False)]["lat"]
        ):
            lat_factor.append(lat2 / lat1)
        factors.append([tput_factor] + lat_factor)
    groups[(c, s)] = factors

figs = os.path.join(baseDir, "figs")
os.makedirs(figs, exist_ok=True)


width = 0.2
nBars = 5
x = np.arange(len(ratio)) * (width * nBars / 2) * 3
for c, s in itertools.product(concurrency, skew):
    name = f"con{c}_skew{s}"
    factors = groups[(c, s)]
    for i in range(nBars):
        plt.bar(x + width * (i - (nBars - 1) / 2.0), [f[i] for f in factors], width)

    plt.xticks(x, [f"{w}:{r}" for w, r in ratio])
    plt.xlabel("write-read ratio")
    plt.ylabel("factor")
    plt.legend(
        ["tput", "p50-write", "p99-write", "p50-read", "p99-read"], loc="upper left"
    )
    plt.title(name)
    plt.savefig(os.path.join(figs, f"{name}.png"))
    plt.clf()
