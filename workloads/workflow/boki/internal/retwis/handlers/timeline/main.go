package main

import (
	"log"
	"os"
	"strconv"

	"github.com/eniac/Beldi/internal/retwis/core"
	"github.com/eniac/Beldi/pkg/cayonlib"
	"github.com/mitchellh/mapstructure"

	"cs.utexas.edu/zjia/faas"
)

var maxReturnPosts = 4

func init() {
	if mr, err := strconv.Atoi(os.Getenv("MAX_RETURN_POSTS")); err == nil {
		maxReturnPosts = mr
	} else {
		panic("invalid MAX_RETURN_POSTS")
	}
}

func Handler(env *cayonlib.Env) interface{} {
	var rpcInput core.RPCInput
	cayonlib.CHECK(mapstructure.Decode(env.Input, &rpcInput))
	switch rpcInput.Function {
	case "Timeline":
		var input core.TimelineInput
		cayonlib.CHECK(mapstructure.Decode(rpcInput.Input, &input))
		result := core.GetTimeline(env, input, maxReturnPosts)
		return result
	default:
		log.Println("ERROR: no such function")
		panic(rpcInput)
	}
}

func main() {
	// lambda.Start(cayonlib.Wrapper(Handler))
	faas.Serve(cayonlib.CreateFuncHandlerFactory(Handler))
}
