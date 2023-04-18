package main

import (
	"math/rand"
	"os"
	"strconv"
	"time"

	"github.com/aws/aws-sdk-go/service/dynamodb/expression"
	"github.com/eniac/Beldi/pkg/cayonlib"

	"cs.utexas.edu/zjia/faas"
)

const table = "singleop"

var nKeys = 10000
var value = 1

func init() {
	if nk, err := strconv.Atoi(os.Getenv("NUM_KEYS")); err == nil {
		nKeys = nk
	} else {
		panic("invalid NUM_KEYS")
	}
	rand.Seed(time.Now().UnixNano())
}

func Handler(env *cayonlib.Env) interface{} {
	results := map[string]int64{}

	start := time.Now()
	if cayonlib.TYPE == "NONE" {
		cayonlib.LibReadSingleVersion(table, strconv.Itoa(rand.Intn(nKeys)))
	} else {
		cayonlib.Read(env, table, strconv.Itoa(rand.Intn(nKeys)))
	}
	results["Read"] = time.Since(start).Microseconds()

	start = time.Now()
	if cayonlib.TYPE == "NONE" {
		cayonlib.LibWriteMultiVersion(table, strconv.Itoa(rand.Intn(nKeys)), 0, map[expression.NameBuilder]expression.OperandBuilder{
			expression.Name("V"): expression.Value(value),
		})
	} else {
		cayonlib.Write(env, table, strconv.Itoa(rand.Intn(nKeys)), map[expression.NameBuilder]expression.OperandBuilder{
			expression.Name("V"): expression.Value(value),
		})
	}
	results["Write"] = time.Since(start).Microseconds()

	start = time.Now()
	if cayonlib.TYPE == "NONE" {
		env.FaasEnv.InvokeFunc(env.FaasCtx, "nop", []byte{})
	} else {
		cayonlib.SyncInvoke(env, "nop", "")
	}
	results["Invoke"] = time.Since(start).Microseconds()

	return results
}

func main() {
	faas.Serve(cayonlib.CreateFuncHandlerFactory(Handler))
}
