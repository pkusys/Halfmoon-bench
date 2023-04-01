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

const table = "rw"

var nKeys = 10000
var valueSize = 256 // bytes
var value string

// var nOps float64
// var readRatio float64

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
	// if ops, err := strconv.ParseFloat(os.Getenv("NUM_OPS"), 64); err == nil {
	// 	nOps = ops
	// } else {
	// 	panic("invalid NUM_OPS")
	// }
	// rr, err := strconv.ParseFloat(os.Getenv("READ_RATIO"), 64)
	// if err != nil || rr < 0 || rr > 1 {
	// 	panic("invalid READ_RATIO")
	// } else {
	// 	readRatio = rr
	// }
	value = utils.RandomString(valueSize)
	// nReads := int(nOps * readRatio)
	// if nReads == 0 || nReads == int(nOps) {
	// 	log.Printf("[WARN] Missing read or write ops: read %d, write %d ratio %.1f", int(nOps*readRatio), int(nOps)-nReads, readRatio)
	// }
	rand.Seed(time.Now().UnixNano())
}

func Handler(env *cayonlib.Env) interface{} {
	input := env.Input.(map[string]interface{})
	mode := input["mode"].(string)
	nOps := input["nOps"].(float64)
	readRatio := input["readRatio"].(float64)
	nReads := int(nOps * readRatio)
	log.Printf("[INFO] nOps %d, readRatio %.1f, nReads %d", int(nOps), readRatio, nReads)
	if nReads == 0 || nReads == int(nOps) {
		log.Println("[WARN] Missing read or write ops")
	}
	for i := 0; i < nReads; i++ {
		cayonlib.ReadWithMode(env, mode, table, strconv.Itoa(rand.Intn(nKeys)))
	}
	for i := 0; i < int(nOps)-nReads; i++ {
		cayonlib.WriteWithMode(env, mode, table, strconv.Itoa(rand.Intn(nKeys)), map[expression.NameBuilder]expression.OperandBuilder{
			expression.Name("V"): expression.Value(value),
		})
	}
	return nil
}

func main() {
	faas.Serve(cayonlib.CreateFuncHandlerFactory(Handler))
}
