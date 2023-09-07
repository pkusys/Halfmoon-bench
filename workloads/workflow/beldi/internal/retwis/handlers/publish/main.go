package main

import (
	"log"

	"github.com/eniac/Beldi/internal/retwis/core"
	"github.com/eniac/Beldi/pkg/beldilib"
	"github.com/mitchellh/mapstructure"

	"cs.utexas.edu/zjia/faas"
)

func Handler(env *beldilib.Env) interface{} {
	var rpcInput core.RPCInput
	beldilib.CHECK(mapstructure.Decode(env.Input, &rpcInput))
	switch rpcInput.Function {
	case "Publish":
		var input core.PublishInput
		beldilib.CHECK(mapstructure.Decode(rpcInput.Input, &input))
		core.Publish(env, input)
		return 0
	default:
		log.Println("ERROR: no such function")
		panic(rpcInput)
	}
}

func main() {
	// lambda.Start(beldilib.Wrapper(Handler))
	faas.Serve(beldilib.CreateFuncHandlerFactory(Handler))
}
