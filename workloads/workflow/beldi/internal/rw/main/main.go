package main

import (
	"fmt"
	"log"
	"os"
	"strconv"

	"math/rand"

	"github.com/aws/aws-sdk-go/service/dynamodb/expression"
	"github.com/eniac/Beldi/internal/rw/utils"
	"github.com/eniac/Beldi/pkg/beldilib"

	"cs.utexas.edu/zjia/faas"
)

const table = "rw"

var nKeys = 10000
var valueSize = 256 // bytes
var value []byte

var nOps float64
var readRatio float64

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
	value = utils.RandomString(valueSize)
}

func Handler(env *beldilib.Env) interface{} {
	if beldilib.TYPE != "BASELINE" {
		log.Fatalf("TYPE should be BASELINE, but got %s", beldilib.TYPE)
	}
	for i := 0; i < int(nOps*readRatio); i++ {
		beldilib.Read(env, fmt.Sprintf("b%s", table), strconv.Itoa(rand.Intn(nKeys)))
	}
	for i := 0; i < int(nOps*(1-readRatio)); i++ {
		beldilib.Write(env, fmt.Sprintf("b%s", table), strconv.Itoa(rand.Intn(nKeys)), map[expression.NameBuilder]expression.OperandBuilder{
			expression.Name("V"): expression.Value(value),
		})
	}
	return nil
}

func main() {
	faas.Serve(beldilib.CreateFuncHandlerFactory(Handler))
}
