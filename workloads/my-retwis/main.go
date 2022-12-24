package main

import (
	"fmt"

	"cs.utexas.edu/zjia/faas-retwis/handlers"

	"cs.utexas.edu/zjia/faas"
	"cs.utexas.edu/zjia/faas/types"
)

type funcHandlerFactory struct {
}

func (f *funcHandlerFactory) New(env types.Environment, funcName string) (types.FuncHandler, error) {
	switch funcName {
	case "RetwisRegister":
		return handlers.NewRegisterHandler(env), nil
	case "RetwisProfile":
		return handlers.NewProfileHandler(env), nil
	case "RetwisFollow":
		return handlers.NewFollowHandler(env), nil
	case "RetwisPost":
		return handlers.NewPostHandler(env), nil
	case "RetwisPostList":
		return handlers.NewPostListHandler(env), nil
	default:
		return nil, fmt.Errorf("Unknown function name: %s", funcName)
	}
}

func (f *funcHandlerFactory) GrpcNew(env types.Environment, service string) (types.GrpcFuncHandler, error) {
	return nil, fmt.Errorf("Not implemented")
}

func main() {
	faas.Serve(&funcHandlerFactory{})
}
