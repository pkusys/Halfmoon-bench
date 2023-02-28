package main

import (
	// "github.com/aws/aws-lambda-go/lambda"
	"github.com/eniac/Beldi/internal/media/core"
	"github.com/eniac/Beldi/pkg/cayonlib"
	"github.com/mitchellh/mapstructure"

	"cs.utexas.edu/zjia/faas"
)

func Handler(env *cayonlib.Env) interface{} {
	var rpcInput core.RPCInput
	cayonlib.CHECK(mapstructure.Decode(env.Input, &rpcInput))
	req := rpcInput.Input.(map[string]interface{})
	switch rpcInput.Function {
	case "UploadText2":
		core.UploadText2(env, req["reqId"].(string), req["text"].(string))
		return 0
	}
	panic("no such function")
}

func main() {
	// lambda.Start(cayonlib.Wrapper(Handler))
	faas.Serve(cayonlib.CreateFuncHandlerFactory(Handler))
}
