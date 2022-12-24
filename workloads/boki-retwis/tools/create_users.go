package main

import (
	"flag"
	"fmt"
	"log"
	"math/rand"

	"cs.utexas.edu/zjia/faas-retwis/utils"
	"github.com/openacid/low/mathext/zipf"
)

var FLAGS_faas_gateway string
var FLAGS_fn_prefix string
var FLAGS_num_users int
var FLAGS_followers_per_user int
var FLAGS_concurrency int
var FLAGS_rand_seed int
var FLAGS_zipf_skew float64
var rng *zipf.Zipf

func init() {
	flag.StringVar(&FLAGS_faas_gateway, "faas_gateway", "127.0.0.1:8081", "")
	flag.StringVar(&FLAGS_fn_prefix, "fn_prefix", "", "")
	flag.IntVar(&FLAGS_num_users, "num_users", 1000, "")
	flag.IntVar(&FLAGS_followers_per_user, "followers_per_user", 16, "")
	flag.IntVar(&FLAGS_concurrency, "concurrency", 1, "")
	flag.IntVar(&FLAGS_rand_seed, "rand_seed", 23333, "")
	flag.Float64Var(&FLAGS_zipf_skew, "zipf_skew", 1.1, "")

	rand.Seed(int64(FLAGS_rand_seed))
}

func sampleId() uint64 {
	id := uint64(rng.Float64(rand.Float64()))
	if id == 0 {
		log.Fatalf("generated id is 0")
	}
	return id
}

func retwisInit() {
	client := utils.NewFaasClient(FLAGS_faas_gateway, 1)
	client.AddJsonFnCall(FLAGS_fn_prefix+"RetwisInit", utils.JSONValue{})
	result := client.WaitForResults()[0]
	if !result.Result.Success {
		log.Fatalf("[FATAL] RetwisInit failed")
	}
}

func createUsers() {
	log.Printf("[INFO] Creating %d users", FLAGS_num_users)
	client := utils.NewFaasClient(FLAGS_faas_gateway, FLAGS_concurrency)
	for i := 1; i <= FLAGS_num_users; i++ {
		client.AddJsonFnCall(FLAGS_fn_prefix+"RetwisRegister", utils.JSONValue{
			"username": fmt.Sprintf("testuser_%d", i),
			"password": fmt.Sprintf("password_%d", i),
		})
	}
	results := client.WaitForResults()

	numSuccess := 0
	for _, result := range results {
		if result.Result.Success {
			numSuccess++
		}
	}
	if numSuccess < FLAGS_num_users {
		log.Printf("[ERROR] %d UserRegister requests failed", FLAGS_num_users-numSuccess)
	}
}

func createFollowers() {
	log.Printf("[INFO] Creating %d followers for each user", FLAGS_followers_per_user)
	userIds1 := make([]int, 0, 1024)
	userIds2 := make([]int, 0, 1024)
	for i := 1; i <= FLAGS_num_users; i++ {
		for j := 0; j < FLAGS_followers_per_user; j++ {
			followeeId := 0
			for {
				followeeId = int(sampleId())
				if followeeId != i {
					break
				}
			}
			userIds1 = append(userIds1, i)
			userIds2 = append(userIds2, followeeId)
		}
	}
	totalRequests := len(userIds1)
	rand.Shuffle(totalRequests, func(i, j int) {
		userIds1[i], userIds1[j] = userIds1[j], userIds1[i]
		userIds2[i], userIds2[j] = userIds2[j], userIds2[i]
	})

	client := utils.NewFaasClient(FLAGS_faas_gateway, FLAGS_concurrency)
	for i := 0; i < totalRequests; i++ {
		client.AddJsonFnCall(FLAGS_fn_prefix+"RetwisFollow", utils.JSONValue{
			"userId":     fmt.Sprintf("%08x", userIds1[i]),
			"followeeId": fmt.Sprintf("%08x", userIds2[i]),
			// "retry":      true,
		})
	}
	results := client.WaitForResults()

	numSuccess := 0
	for _, result := range results {
		if result.Result.Success {
			numSuccess++
		}
	}
	if numSuccess < totalRequests {
		log.Printf("[ERROR] %d Follow requests failed", totalRequests-numSuccess)
	}
}

func main() {
	flag.Parse()

	rng = zipf.New(1, float64(FLAGS_num_users), FLAGS_zipf_skew)

	retwisInit()
	createUsers()
	createFollowers()
}
