package main

import (
	"flag"
	"fmt"
	"log"
	"math/rand"
	"time"

	"cs.utexas.edu/zjia/faas-retwis/utils"

	"github.com/montanaflynn/stats"
	"github.com/openacid/low/mathext/zipf"
)

var FLAGS_faas_gateway string
var FLAGS_fn_prefix string

// var FLAGS_num_users int
var FLAGS_concurrency int
var FLAGS_duration int

// var FLAGS_percentages string

// var FLAGS_bodylen int
var FLAGS_rand_seed int
var FLAGS_zipf_skew float64
var FLAGS_keyspace int
var FLAGS_read_keys int
var FLAGS_write_keys int
var rng *zipf.Zipf

func init() {
	flag.StringVar(&FLAGS_faas_gateway, "faas_gateway", "127.0.0.1:8081", "")
	flag.StringVar(&FLAGS_fn_prefix, "fn_prefix", "", "")
	// flag.IntVar(&FLAGS_num_users, "num_users", 1000, "")
	flag.IntVar(&FLAGS_concurrency, "concurrency", 1, "")
	flag.IntVar(&FLAGS_duration, "duration", 10, "")
	// flag.StringVar(&FLAGS_percentages, "percentages", "50,50", "write,read")
	// flag.IntVar(&FLAGS_bodylen, "bodylen", 64, "")
	flag.IntVar(&FLAGS_rand_seed, "rand_seed", 23333, "")
	flag.Float64Var(&FLAGS_zipf_skew, "zipf_skew", 0.5, "")
	flag.IntVar(&FLAGS_read_keys, "read_keys", 8, "")
	flag.IntVar(&FLAGS_write_keys, "write_keys", 1, "")
	flag.IntVar(&FLAGS_keyspace, "keyspace", 10000, "")

	rand.Seed(int64(FLAGS_rand_seed))
}

// func parsePercentages(s string) ([]int, error) {
// 	parts := strings.Split(s, ",")
// 	if len(parts) != 2 {
// 		return nil, fmt.Errorf("Need exactly four parts splitted by comma")
// 	}
// 	results := make([]int, 2)
// 	for i, part := range parts {
// 		if parsed, err := strconv.Atoi(part); err != nil {
// 			return nil, fmt.Errorf("Failed to parse %d-th part", i)
// 		} else {
// 			results[i] = parsed
// 		}
// 	}
// 	for i := 1; i < len(results); i++ {
// 		results[i] += results[i-1]
// 	}
// 	if results[len(results)-1] != 100 {
// 		return nil, fmt.Errorf("Sum of all parts is not 100")
// 	}
// 	return results, nil
// }

func sampleId() uint64 {
	id := uint64(rng.Float64(rand.Float64())) + 1
	if id == 0 {
		log.Fatalf("generated id is 0")
	}
	return id
}

func buildTestRequest() utils.JSONValue {
	n := FLAGS_read_keys
	if FLAGS_write_keys > FLAGS_read_keys {
		n = FLAGS_write_keys
	}
	tags := make([]uint64, 0, n)
	tagsMap := make(map[uint64]struct{})
	for i := 0; i < n; i++ {
		for {
			id := sampleId()
			if _, ok := tagsMap[id]; !ok {
				tags = append(tags, id)
				tagsMap[id] = struct{}{}
				break
			}
		}
	}
	return utils.JSONValue{
		"readkeys":  tags[:FLAGS_read_keys],
		"writekeys": tags[:FLAGS_write_keys],
	}
}

const kTxnConflitMsg = "Failed to commit transaction due to conflicts"

func printFnResult(fnName string, duration time.Duration, results []*utils.FaasCall) {
	total := 0
	succeeded := 0
	txnConflit := 0
	latencies := make([]float64, 0, 128)
	runningLatencies := make([]float64, 0, 128)
	readLatencies := make([]float64, 0, 128)
	writeLatencies := make([]float64, 0, 128)
	for _, result := range results {
		if result.FnName == FLAGS_fn_prefix+fnName {
			total++
			if result.Result.Success {
				succeeded++
			} else if result.Result.Message == kTxnConflitMsg {
				txnConflit++
			}
			if result.Result.StatusCode == 200 {
				d := result.Result.Duration
				latencies = append(latencies, float64(d.Microseconds()))
				runningLatencies = append(runningLatencies, result.Result.Response["duration"].(float64))
				for _, v := range result.Result.Response["readLatency"].([]interface{}) {
					readLatencies = append(readLatencies, v.(float64))
				}
				for _, v := range result.Result.Response["writeLatency"].([]interface{}) {
					writeLatencies = append(writeLatencies, v.(float64))
				}
			}
		}
	}
	if total == 0 {
		return
	}
	failed := total - succeeded - txnConflit
	fmt.Printf("[%s]\n", fnName)
	fmt.Printf("Throughput: %.1f requests per sec\n", float64(total)/duration.Seconds())
	if txnConflit > 0 {
		ratio := float64(txnConflit) / float64(txnConflit+succeeded)
		fmt.Printf("Transaction conflits: %d (%.2f%%)\n", txnConflit, ratio*100.0)
	}
	if failed > 0 {
		ratio := float64(failed) / float64(total)
		fmt.Printf("Transaction conflits: %d (%.2f%%)\n", failed, ratio*100.0)
	}
	if len(latencies) > 0 {
		median, _ := stats.Median(latencies)
		p99, _ := stats.Percentile(latencies, 99.0)
		fmt.Printf("End2End Latency: median = %.3fms, tail (p99) = %.3fms\n", median/1000.0, p99/1000.0)
	}
	if len(runningLatencies) > 0 {
		median, _ := stats.Median(runningLatencies)
		p99, _ := stats.Percentile(runningLatencies, 99.0)
		fmt.Printf("Running Latency: median = %.3fms, tail (p99) = %.3fms\n", median/1000.0, p99/1000.0)
	}
	if len(readLatencies) > 0 {
		median, _ := stats.Median(readLatencies)
		p99, _ := stats.Percentile(readLatencies, 99.0)
		fmt.Printf("Read Latency: median = %.3fms, tail (p99) = %.3fms\n", median/1000.0, p99/1000.0)
	}
	if len(writeLatencies) > 0 {
		median, _ := stats.Median(writeLatencies)
		p99, _ := stats.Percentile(writeLatencies, 99.0)
		fmt.Printf("Write Latency: median = %.3fms, tail (p99) = %.3fms\n", median/1000.0, p99/1000.0)
	}
}

func prewarm() {
	client := utils.NewFaasClient(FLAGS_faas_gateway, FLAGS_concurrency)
	log.Printf("[INFO] Start prewarm for %d seconds with concurrency of %d", 5, FLAGS_concurrency)
	startTime := time.Now()
	for {
		if time.Since(startTime) > time.Duration(5)*time.Second {
			break
		}
		client.AddJsonFnCall(FLAGS_fn_prefix+"Test", buildTestRequest())
	}
	// elapsed := time.Since(startTime)
	client.WaitForResults()
	// fmt.Printf("Benchmark runs for %v, %.1f request per sec\n", elapsed, float64(len(results))/elapsed.Seconds())

	// printFnResult("TestWrite", elapsed, results)
	// printFnResult("TestRead", elapsed, results)
}

func benchmark() {
	client := utils.NewFaasClient(FLAGS_faas_gateway, FLAGS_concurrency)
	log.Printf("[INFO] Start running for %d seconds with concurrency of %d", FLAGS_duration, FLAGS_concurrency)
	startTime := time.Now()
	for {
		if time.Since(startTime) > time.Duration(FLAGS_duration)*time.Second {
			break
		}
		client.AddJsonFnCall(FLAGS_fn_prefix+"Test", buildTestRequest())
	}
	elapsed := time.Since(startTime)
	results := client.WaitForResults()
	fmt.Printf("Benchmark runs for %v, %.1f request per sec\n", elapsed, float64(len(results))/elapsed.Seconds())
	printFnResult("Test", elapsed, results)
}

func main() {
	flag.Parse()

	rng = zipf.New(1, float64(FLAGS_keyspace), FLAGS_zipf_skew)

	prewarm()
	benchmark()
}
