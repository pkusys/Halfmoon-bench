package main

import (
	"log"

	"github.com/eniac/Beldi/internal/retwis/core"
	"github.com/eniac/Beldi/pkg/cayonlib"
	"github.com/mitchellh/mapstructure"

	"cs.utexas.edu/zjia/faas"
)

func Handler(env *cayonlib.Env) interface{} {
	var rpcInput core.RPCInput
	cayonlib.CHECK(mapstructure.Decode(env.Input, &rpcInput))
	switch rpcInput.Function {
	case "Publish":
		var input core.PublishInput
		cayonlib.CHECK(mapstructure.Decode(rpcInput.Input, &input))
		core.Publish(env, input)
		return 0
	default:
		log.Println("ERROR: no such function")
		panic(rpcInput)
	}
}

func main() {
	// lambda.Start(cayonlib.Wrapper(Handler))
	faas.Serve(cayonlib.CreateFuncHandlerFactory(Handler))
}
