package main

import (
	// "github.com/aws/aws-lambda-go/lambda"
	"os"
	"strconv"

	"github.com/eniac/Beldi/internal/hotel/main/geo"
	"github.com/eniac/Beldi/pkg/cayonlib"
	"github.com/mitchellh/mapstructure"

	"cs.utexas.edu/zjia/faas"
)

var kNearest = 20

func init() {
	if kn, err := strconv.Atoi(os.Getenv("K_NEAREST")); err == nil {
		kNearest = kn
	} else {
		panic("invalid K_NEAREST")
	}
}

func Handler(env *cayonlib.Env) interface{} {
	req := geo.Request{}
	err := mapstructure.Decode(env.Input, &req)
	cayonlib.CHECK(err)
	return geo.Nearby(env, req, kNearest)
}

func main() {
	// lambda.Start(cayonlib.Wrapper(Handler))
	faas.Serve(cayonlib.CreateFuncHandlerFactory(Handler))
}
