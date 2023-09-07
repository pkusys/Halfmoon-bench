package main

import (
	"log"
	"os"
	"strconv"
	"time"

	"math/rand"

	"github.com/aws/aws-sdk-go/service/dynamodb/expression"
	"github.com/eniac/Beldi/internal/utils"
	"github.com/eniac/Beldi/pkg/cayonlib"

	"cs.utexas.edu/zjia/faas"
)

const table = "recovery"

var nKeys = 10000
var valueSize = 256 // bytes
var value string

var nOps float64
var readRatio float64
var nReads int

// var sleepDuration = 5 * time.Millisecond
var failRate float64

func init() {
	if nk, err := strconv.Atoi(os.Getenv("NUM_KEYS")); err == nil {
		nKeys = nk
	} else {
		panic("invalid NUM_KEYS")
	}
	if vs, err := strconv.Atoi(os.Getenv("VALUE_SIZE")); err == nil {
		valueSize = vs
	} else {
		panic("invalid VALUE_SIZE")
	}
	if ops, err := strconv.ParseFloat(os.Getenv("NUM_OPS"), 64); err == nil {
		nOps = ops
	} else {
		panic("invalid NUM_OPS")
	}
	rr, err := strconv.ParseFloat(os.Getenv("READ_RATIO"), 64)
	if err != nil || rr < 0 || rr > 1 {
		panic("invalid READ_RATIO")
	} else {
		readRatio = rr
	}
	nReads = int(nOps * readRatio)
	f, err := strconv.ParseFloat(os.Getenv("FAIL_RATE"), 64)
	if err != nil || f < 0 || f > 1 {
		panic("invalid FAIL_RATE")
	} else {
		failRate = f
	}
	log.Printf("[INFO] nKeys=%d, valueSize=%d, nOps=%d, readRatio=%.2f, nReads=%d, failRate=%.2f", nKeys, valueSize, int(nOps), readRatio, nReads, failRate)

	value = utils.RandomString(valueSize)
	rand.Seed(time.Now().UnixNano())
}

func runOnce(env *cayonlib.Env, keys []int) {
	for i := 0; i < nReads; i++ {
		cayonlib.Read(env, table, strconv.Itoa(keys[i]))
		// time.Sleep(sleepDuration)
	}
	for i := nReads; i < int(nOps); i++ {
		cayonlib.Write(env, table, strconv.Itoa(keys[i]), map[expression.NameBuilder]expression.OperandBuilder{
			expression.Name("V"): expression.Value(value),
		}, false)
		// time.Sleep(sleepDuration)
	}
}

func Handler(env *cayonlib.Env) interface{} {
	keys := []int{}
	for i := 0; i < int(nOps); i++ {
		keys = append(keys, rand.Intn(nKeys))
	}
	InitStep := env.StepNumber
	InitSeqNum := env.SeqNum
	for {
		runOnce(env, keys)
		if rand.Float64() >= failRate {
			break
		}
		env.StepNumber = InitStep
		env.SeqNum = InitSeqNum
	}
	return nil
}

func main() {
	faas.Serve(cayonlib.CreateFuncHandlerFactory(Handler))
}
