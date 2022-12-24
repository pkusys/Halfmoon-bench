package main

import (
	"fmt"
	"log"

	"cs.utexas.edu/zjia/faas-retwis/handlers"

	"cs.utexas.edu/zjia/faas"
	"cs.utexas.edu/zjia/faas/types"
)

type funcHandlerFactory struct {
}

func (f *funcHandlerFactory) New(env types.Environment, funcName string) (types.FuncHandler, error) {
	switch funcName {
	// case "TestWrite":
	// 	return handlers.NewTestWriteHandler(env), nil
	// case "TestRead":
	// 	return handlers.NewTestReadHandler(env), nil
	case "RWTxn":
		return handlers.NewRWTxnHandler(env), nil
	default:
		return nil, fmt.Errorf("Unknown function name: %s", funcName)
	}
}

func (f *funcHandlerFactory) GrpcNew(env types.Environment, service string) (types.GrpcFuncHandler, error) {
	return nil, fmt.Errorf("Not implemented")
}

func main() {
	if handlers.EnableCCPath {
		log.Println("CCPath is enabled")
	}
	faas.Serve(&funcHandlerFactory{})
}
