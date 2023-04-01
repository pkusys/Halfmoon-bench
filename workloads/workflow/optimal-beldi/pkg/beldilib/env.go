package beldilib

import (
	"context"

	"cs.utexas.edu/zjia/faas/types"
)

type Env struct {
	LambdaId    string
	InstanceId  string
	LogTable    string
	IntentTable string
	LocalTable  string
	StepNumber  int32
	CounterTS   int64
	Input       interface{}
	TxnId       string
	Instruction string
	Baseline    bool
	FaasCtx     context.Context
	FaasEnv     types.Environment
}
