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
	case "Login":
		var input core.LoginInput
		beldilib.CHECK(mapstructure.Decode(rpcInput.Input, &input))
		result := core.Login(env, input)
		return result
	default:
		log.Println("ERROR: no such function")
		panic(rpcInput)
	}
}

func main() {
	// lambda.Start(beldilib.Wrapper(Handler))
	faas.Serve(beldilib.CreateFuncHandlerFactory(Handler))
}
