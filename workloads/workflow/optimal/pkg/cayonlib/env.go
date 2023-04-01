package cayonlib

import (
	"context"

	"cs.utexas.edu/zjia/faas/types"
)

type LogEntry struct {
	SeqNum uint64
	Data   map[string]interface{}
}

type Env struct {
	LambdaId    string
	InstanceId  string
	StepNumber  int32
	SeqNum      uint64
	Input       interface{}
	TxnId       string
	Instruction string
	Baseline    bool
	FaasCtx     context.Context
	FaasEnv     types.Environment
	Fsm         *IntentFsm
	LogSize     int
}
