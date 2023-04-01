package main

import (
	"os"
	"strconv"

	"github.com/eniac/Beldi/pkg/cayonlib"

	"cs.utexas.edu/zjia/faas"
)

const table = "rw"

var nKeys = 10000

func init() {
	if nk, err := strconv.Atoi(os.Getenv("NUM_KEYS")); err == nil {
		nKeys = nk
	} else {
		panic("invalid NUM_KEYS")
	}
}

func Handler(env *cayonlib.Env) interface{} {
	for i := 0; i < nKeys; i++ {
		cayonlib.Read(env, table, strconv.Itoa(i))
	}
	return nil
}

func main() {
	faas.Serve(cayonlib.CreateFuncHandlerFactory(Handler))
}
