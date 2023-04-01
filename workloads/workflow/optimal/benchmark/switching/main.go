package main

import (
	"flag"
	"fmt"
	"log"
	"sort"
	"strconv"
	"strings"
	"time"
)

var FLAGS_faas_gateway string
var FLAGS_fn_prefix string
var FLAGS_concurrency int
var FLAGS_duration int64
var FLAGS_cycle int64
var FLAGS_prewarm bool
var FLAGS_num_ops int
var FLAGS_read_ratios string

var logModes = []string{"READLOG", "WRITELOG"}
var readRatios = []float64{0.2, 0.8}

func init() {
	flag.StringVar(&FLAGS_faas_gateway, "faas_gateway", "127.0.0.1:8081", "")
	flag.StringVar(&FLAGS_fn_prefix, "fn_prefix", "", "")
	flag.IntVar(&FLAGS_concurrency, "concurrency", 1, "")
	flag.Int64Var(&FLAGS_duration, "duration", 10, "")
	flag.Int64Var(&FLAGS_cycle, "cycle", 5, "")
	flag.BoolVar(&FLAGS_prewarm, "prewarm", false, "")
	flag.IntVar(&FLAGS_num_ops, "num_ops", 5, "")
	flag.StringVar(&FLAGS_read_ratios, "read_ratios", "0.2,0.8", "ratio of each cycle")
}

func parseReadRatios(s string) []float64 {
	parts := strings.Split(s, ",")
	if len(parts) != 2 {
		log.Fatalf("Need exactly two parts splitted by comma")
	}
	results := make([]float64, 2)
	for i, part := range parts {
		if parsed, err := strconv.ParseFloat(part, 64); err != nil {
			log.Fatalf("Failed to parse %d-th part", i)
		} else {
			results[i] = parsed
		}
	}
	return results
}

func cycleId(t time.Duration) int {
	return int(t.Microseconds()/(FLAGS_cycle*1000000)) % 2
}

func buildRequest(mode string, readRatio float64) JSONValue {
	return JSONValue{
		"InstanceId": "",
		"CallerName": "",
		"Async":      false,
		"Input": JSONValue{
			"mode":      mode,
			"readRatio": readRatio,
			"nOps":      FLAGS_num_ops,
		},
	}
}

func main() {
	flag.Parse()
	readRatios = parseReadRatios(FLAGS_read_ratios)

	client := NewFaasClient(FLAGS_faas_gateway, FLAGS_concurrency)
	log.Printf("[INFO] Start running for %d seconds with concurrency of %d", FLAGS_duration, FLAGS_concurrency)
	startTime := time.Now()
	currentCycle := 0
	transitionalReqs := make(map[int64]struct{})
	for i := 0; i < FLAGS_concurrency; i++ {
		client.AddJsonFnCall(time.Since(startTime), FLAGS_fn_prefix+"rw", buildRequest(logModes[0], readRatios[0]))
	}
	for {
		call := client.Recv()
		now := time.Since(startTime)
		if now > time.Duration(FLAGS_duration)*time.Second {
			break
		}
		cid := cycleId(now)
		if cid != currentCycle {
			currentCycle = cid
			for k, v := range client.runningReqs {
				transitionalReqs[k] = v
			}
			if len(transitionalReqs) != 0 {
				fmt.Printf("%d: BEGIN TRANSITION\n", now.Microseconds())
			}
		}
		nTransitional := len(transitionalReqs)
		delete(transitionalReqs, call.TS.Microseconds())
		if nTransitional != 0 && len(transitionalReqs) == 0 {
			fmt.Printf("%d: END TRANSITION\n", now.Microseconds())
		}
		mode := logModes[cid]
		readRatio := readRatios[cid]
		if len(transitionalReqs) != 0 {
			mode = "TRANSITIONAL"
		}
		client.AddJsonFnCall(now, FLAGS_fn_prefix+"rw", buildRequest(mode, readRatio))
	}
	elapsed := time.Since(startTime)
	client.Wait()
	log.Printf("Benchmark runs for %v, %.1f request per sec\n", elapsed, float64(len(client.results))/elapsed.Seconds())

	if FLAGS_prewarm {
		return
	}
	sort.Slice(client.results, func(i, j int) bool {
		return client.results[i].TS < client.results[j].TS
	})
	for _, call := range client.results {
		ts := call.TS.Microseconds()
		latency := float64(call.Result.Duration.Microseconds()) / 1000.0
		logMode := call.Input["Input"].(map[string]interface{})["mode"].(string)
		fmt.Printf("%d: %.2fms %s\n", ts, latency, logMode)
	}
}
