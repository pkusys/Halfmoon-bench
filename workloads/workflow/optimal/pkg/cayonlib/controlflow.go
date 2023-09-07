package cayonlib

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"

	// "github.com/aws/aws-lambda-go/lambdacontext"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/dynamodb/expression"

	// lambdaSdk "github.com/aws/aws-sdk-go/service/lambda"
	"github.com/golang/snappy"
	"github.com/lithammer/shortuuid"
	"github.com/mitchellh/mapstructure"

	"cs.utexas.edu/zjia/faas/types"

	"context"
	// "strings"
	"time"
)

type InputWrapper struct {
	CallerName  string      `mapstructure:"CallerName"`
	CallerId    string      `mapstructure:"CallerId"`
	CallerStep  int32       `mapstructure:"CallerStep"`
	SeqNum      uint64      `mapstructure:"SeqNum"`
	InstanceId  string      `mapstructure:"InstanceId"`
	Input       interface{} `mapstructure:"Input"`
	TxnId       string      `mapstructure:"TxnId"`
	Instruction string      `mapstructure:"Instruction"`
	Async       bool        `mapstructure:"Async"`
	LogSize     int         `mapstructure:"LogSize"`
}

func (iw *InputWrapper) Serialize() []byte {
	stream, err := json.Marshal(*iw)
	CHECK(err)
	return stream
}

func (iw *InputWrapper) Deserialize(stream []byte) {
	err := json.Unmarshal(stream, iw)
	CHECK(err)
}

type StackTraceCall struct {
	Label string `json:"label"`
	Line  int    `json:"line"`
	Path  string `json:"path"`
}

func (ie *InvokeError) Deserialize(stream []byte) {
	err := json.Unmarshal(stream, ie)
	CHECK(err)
	if ie.ErrorMessage == "" {
		panic(errors.New("never happen"))
	}
}

type InvokeError struct {
	ErrorMessage string           `json:"errorMessage"`
	ErrorType    string           `json:"errorType"`
	StackTrace   []StackTraceCall `json:"stackTrace"`
}

type OutputWrapper struct {
	Status  string
	LogSize int
	Output  interface{}
}

func (ow *OutputWrapper) Serialize() []byte {
	stream, err := json.Marshal(*ow)
	CHECK(err)
	return stream
}

func (ow *OutputWrapper) Deserialize(stream []byte) {
	err := json.Unmarshal(stream, ow)
	CHECK(err)
	if ow.Status != "Success" && ow.Status != "Failure" {
		ie := InvokeError{}
		ie.Deserialize(stream)
		panic(ie)
	}
}

func ParseInput(raw interface{}) *InputWrapper {
	var iw InputWrapper
	if body, ok := raw.(map[string]interface{})["body"]; ok {
		CHECK(json.Unmarshal([]byte(body.(string)), &iw))
	} else {
		CHECK(mapstructure.Decode(raw, &iw))
	}
	return &iw
}

func PrepareEnv(iw *InputWrapper, lambdaId string) *Env {
	// s := strings.Split(lambdacontext.FunctionName, "-")
	// lambdaId := s[len(s)-1]
	if iw.InstanceId == "" {
		iw.InstanceId = shortuuid.New()
	}
	return &Env{
		LambdaId:    lambdaId,
		InstanceId:  iw.InstanceId,
		StepNumber:  0,
		SeqNum:      iw.SeqNum,
		Input:       iw.Input,
		TxnId:       iw.TxnId,
		Instruction: iw.Instruction,
		LogSize:     iw.LogSize,
	}
}

func SyncInvoke(env *Env, callee string, input interface{}) (interface{}, string) {
	newLog, preInvokeLog := ProposeNextStep(env, []uint64{IntentLogTag}, aws.JSONValue{
		"type":       "PreInvoke",
		"instanceId": shortuuid.New(),
		"callee":     callee,
		"DONE":       false,
		"ASYNC":      false,
		"INPUT":      input,
		"ST":         time.Now().Unix(),
	})
	// conditonal append failed
	// the outermost function should check env.Instruction = "EXIT"
	if preInvokeLog == nil {
		env.Instruction = "EXIT"
		return nil, ""
	}
	instanceId := preInvokeLog.Data["instanceId"].(string)
	// log.Printf("[INFO] invoking func %v instance %v caller %v", callee, instanceId, env.InstanceId)
	if !newLog {
		if preInvokeLog.Data["type"].(string) == "InvokeResult" {
			log.Printf("[INFO] Seen InvokeResult log instance %v step %d", env.InstanceId, preInvokeLog.StepNumber)
			return preInvokeLog.Data["output"], instanceId
		}
		CheckLogDataField(preInvokeLog, "type", "PreInvoke")
		log.Printf("[INFO] Seen PreInvoke log instance %v step %d", env.InstanceId, preInvokeLog.StepNumber)
	}

	iw := InputWrapper{
		CallerName:  env.LambdaId,
		CallerId:    env.InstanceId,
		CallerStep:  preInvokeLog.StepNumber,
		SeqNum:      preInvokeLog.SeqNum,
		Async:       false,
		InstanceId:  instanceId,
		Input:       input,
		TxnId:       env.TxnId,
		Instruction: env.Instruction,
		LogSize:     env.LogSize,
	}
	if iw.Instruction == "EXECUTE" {
		LibAppendLog(env, TransactionStreamTag(env.LambdaId, env.TxnId), &TxnLogEntry{
			LambdaId: env.LambdaId,
			TxnId:    env.TxnId,
			Callee:   callee,
			WriteOp:  aws.JSONValue{},
		})
	}
	payload := iw.Serialize()
	res, err := env.FaasEnv.InvokeFunc(env.FaasCtx, callee, payload)
	// CHECK(err)
	if err != nil {
		log.Printf("[ERROR] InvokeFunc error instance %v step %d: %s", env.InstanceId, preInvokeLog.StepNumber, err.Error())
		env.Instruction = "EXIT"
		return nil, instanceId
	}
	// log.Printf("[INFO] func %v instance %v returned to caller %v", callee, instanceId, env.InstanceId)
	ow := OutputWrapper{}
	ow.Deserialize(res)
	env.LogSize += ow.LogSize
	switch ow.Status {
	case "Success":
		return ow.Output, iw.InstanceId
	default:
		panic("never happens")
	}
}

func ProposeInvoke(env *Env, callee string, input interface{}) *IntentLogEntry {
	newLog, preInvokeLog := ProposeNextStep(env, []uint64{IntentLogTag}, aws.JSONValue{
		"type":       "PreInvoke",
		"instanceId": shortuuid.New(),
		"callee":     callee,
		"DONE":       false,
		"ASYNC":      false,
		"INPUT":      input,
		"ST":         time.Now().Unix(),
	})
	// conditonal append failed
	// the outermost function should check env.Instruction = "EXIT"
	if preInvokeLog == nil {
		env.Instruction = "EXIT"
		return nil
	}
	if newLog {
		return preInvokeLog
	}
	if preInvokeLog.Data["type"].(string) == "InvokeResult" {
		log.Printf("[INFO] Seen InvokeResult log for step %d", preInvokeLog.StepNumber)
	} else {
		CheckLogDataField(preInvokeLog, "type", "PreInvoke")
		CheckLogDataField(preInvokeLog, "callee", callee)
		log.Printf("[INFO] Seen PreInvoke log for step %d", preInvokeLog.StepNumber)
	}
	return preInvokeLog
}

func AssignedSyncInvoke(env *Env, callee string, preInvokeLog *IntentLogEntry) (interface{}, string) {
	instanceId := preInvokeLog.Data["instanceId"].(string)
	if preInvokeLog.Data["type"].(string) == "InvokeResult" {
		log.Printf("[INFO] Seen InvokeResult log for step %d", preInvokeLog.StepNumber)
		return preInvokeLog.Data["output"], instanceId
	}

	iw := InputWrapper{
		CallerName:  env.LambdaId,
		CallerId:    env.InstanceId,
		CallerStep:  preInvokeLog.StepNumber,
		SeqNum:      preInvokeLog.SeqNum,
		Async:       false,
		InstanceId:  instanceId,
		Input:       preInvokeLog.Data["INPUT"],
		TxnId:       env.TxnId,
		Instruction: env.Instruction,
		LogSize:     env.LogSize,
	}
	if iw.Instruction == "EXECUTE" {
		LibAppendLog(env, TransactionStreamTag(env.LambdaId, env.TxnId), &TxnLogEntry{
			LambdaId: env.LambdaId,
			TxnId:    env.TxnId,
			Callee:   callee,
			WriteOp:  aws.JSONValue{},
		})
	}
	payload := iw.Serialize()
	res, err := env.FaasEnv.InvokeFunc(env.FaasCtx, callee, payload)
	// CHECK(err)
	if err != nil {
		log.Printf("[ERROR] InvokeFunc error instance %v step %d: %s", env.InstanceId, preInvokeLog.StepNumber, err.Error())
		env.Instruction = "EXIT"
		return nil, instanceId
	}
	ow := OutputWrapper{}
	ow.Deserialize(res)
	env.LogSize += ow.LogSize
	switch ow.Status {
	case "Success":
		return ow.Output, iw.InstanceId
	default:
		panic("never happens")
	}
}

func AsyncInvoke(env *Env, callee string, input interface{}) string {
	newLog, preInvokeLog := ProposeNextStep(env, []uint64{IntentLogTag}, aws.JSONValue{
		"type":       "PreInvoke",
		"instanceId": shortuuid.New(),
		"callee":     callee,
		"DONE":       false,
		"ASYNC":      false,
		"INPUT":      input,
		"ST":         time.Now().Unix(),
	})
	// conditonal append failed
	// the outermost function should check env.Instruction = "EXIT"
	if preInvokeLog == nil {
		env.Instruction = "EXIT"
		return ""
	}
	instanceId := preInvokeLog.Data["instanceId"].(string)
	if !newLog {
		if preInvokeLog.Data["type"].(string) == "InvokeResult" {
			log.Printf("[INFO] Seen InvokeResult log for step %d", preInvokeLog.StepNumber)
			return instanceId
		}
		CheckLogDataField(preInvokeLog, "type", "PreInvoke")
		log.Printf("[INFO] Seen PreInvoke log for step %d", preInvokeLog.StepNumber)
	}

	iw := InputWrapper{
		CallerName: env.LambdaId,
		CallerId:   env.InstanceId,
		CallerStep: preInvokeLog.StepNumber,
		SeqNum:     preInvokeLog.SeqNum,
		Async:      true,
		InstanceId: instanceId,
		Input:      input,
		// logsize = 0
	}

	/*
		Should we handle this?
		LibWrite(env.LogTable, pk, map[expression.NameBuilder]expression.OperandBuilder{
			expression.Name("RET"): expression.Value(1),
		})
	*/

	payload := iw.Serialize()
	err := env.FaasEnv.InvokeFuncAsync(env.FaasCtx, callee, payload)
	CHECK(err)
	return iw.InstanceId
}

func getAllTxnLogs(env *Env) []*TxnLogEntry {
	tag := TransactionStreamTag(env.LambdaId, env.TxnId)
	seqNum := uint64(0)
	results := make([]*TxnLogEntry, 0)
	for {
		logEntry, err := env.FaasEnv.SharedLogReadNext(env.FaasCtx, tag, seqNum)
		CHECK(err)
		if logEntry == nil {
			break
		}
		decoded, err := snappy.Decode(nil, logEntry.Data)
		CHECK(err)
		var txnLog TxnLogEntry
		err = json.Unmarshal(decoded, &txnLog)
		CHECK(err)
		if txnLog.LambdaId == env.LambdaId && txnLog.TxnId == env.TxnId {
			txnLog.SeqNum = logEntry.SeqNum
			results = append(results, &txnLog)
		}
		seqNum = logEntry.SeqNum + 1
	}
	return results
}

func TPLCommit(env *Env) {
	txnLogs := getAllTxnLogs(env)
	for _, txnLog := range txnLogs {
		if txnLog.Callee != "" {
			continue
		}
		tablename := txnLog.WriteOp["tablename"].(string)
		key := txnLog.WriteOp["key"].(string)
		update := map[expression.NameBuilder]expression.OperandBuilder{}
		for kk, vv := range txnLog.WriteOp["value"].(map[string]interface{}) {
			update[expression.Name(kk)] = expression.Value(vv)
		}
		Write(env, tablename, key, update, true)
		Unlock(env, tablename, key)
	}
	for _, txnLog := range txnLogs {
		if txnLog.Callee != "" {
			log.Printf("[INFO] Commit transaction %s for callee %s", env.TxnId, txnLog.Callee)
			SyncInvoke(env, txnLog.Callee, aws.JSONValue{})
		}
	}
}

func TPLAbort(env *Env) {
	txnLogs := getAllTxnLogs(env)
	for _, txnLog := range txnLogs {
		if txnLog.Callee != "" {
			continue
		}
		tablename := txnLog.WriteOp["tablename"].(string)
		key := txnLog.WriteOp["key"].(string)
		Unlock(env, tablename, key)
	}
	for _, txnLog := range txnLogs {
		if txnLog.Callee != "" {
			log.Printf("[INFO] Abort transaction %s for callee %s", env.TxnId, txnLog.Callee)
			SyncInvoke(env, txnLog.Callee, aws.JSONValue{})
		}
	}
}

func wrapperInternal(f func(*Env) interface{}, iw *InputWrapper, env *Env) (OutputWrapper, error) {
	if TYPE == "BASELINE" {
		panic("Baseline type not supported")
	}
	if iw.CallerName == "" {
		newLog, intentLog := ProposeNextStep(env, []uint64{IntentLogTag}, aws.JSONValue{
			"InstanceId": env.InstanceId,
			"DONE":       false,
			"ASYNC":      iw.Async,
			"INPUT":      iw.Input,
			"ST":         time.Now().Unix(),
		})
		if intentLog == nil {
			log.Printf("[WARN] Seen conflicting intent log of instance %v", env.InstanceId)
			return OutputWrapper{Status: "Exit"}, fmt.Errorf("exit")
		}
		if !newLog {
			CheckLogDataField(intentLog, "InstanceId", env.InstanceId)
			log.Printf("[INFO] Seen intent log of instance %v", env.InstanceId)
		}
		env.SeqNum = intentLog.SeqNum
	}
	// log.Printf("[INFO] starting instance %v caller %v", env.InstanceId, iw.CallerName)

	var output interface{}
	if env.Instruction == "COMMIT" {
		TPLCommit(env)
		output = 0
	} else if env.Instruction == "ABORT" {
		TPLAbort(env)
		output = 0
	} else {
		output = f(env)
	}
	if env.Instruction == "EXIT" {
		log.Printf("[WARN] instance %v caller %v exited", env.InstanceId, iw.CallerName)
		return OutputWrapper{Status: "Exit"}, fmt.Errorf("exit")
	}
	if iw.CallerName != "" {
		// log.Printf("[INFO] writing result to caller's log: instance %v caller %v", env.InstanceId, iw.CallerName)
		// overwrite is inherently idempotent
		LogStepResultForCaller(env, iw.CallerId, iw.CallerStep, aws.JSONValue{
			"type":       "InvokeResult",
			"instanceId": env.InstanceId,
			"DONE":       true,
			"TS":         time.Now().Unix(),
			"output":     output,
		})
	} else {
		// no need to use conditional append for finish events
		LibAppendLog(env, IntentLogTag, aws.JSONValue{
			"InstanceId": env.InstanceId,
			"DONE":       true,
			"TS":         time.Now().Unix(),
		})
	}
	// log.Printf("[INFO] finishing instance %v caller %v", env.InstanceId, iw.CallerName)

	return OutputWrapper{
		Status:  "Success",
		LogSize: env.LogSize,
		Output:  output,
	}, nil
}

type funcHandlerWrapper struct {
	fnName  string
	handler func(env *Env) interface{}
	env     types.Environment
}

func (w *funcHandlerWrapper) Call(ctx context.Context, input []byte) ([]byte, error) {
	var jsonInput map[string]interface{}
	err := json.Unmarshal(input, &jsonInput)
	if err != nil {
		return nil, err
	}
	iw := ParseInput(jsonInput)
	env := PrepareEnv(iw, w.fnName)
	env.FaasCtx = ctx
	env.FaasEnv = w.env
	env.Fsm = NewIntentFsm(env.InstanceId)
	env.Fsm.Catch(env)
	ow, err := wrapperInternal(w.handler, iw, env)
	if err != nil {
		return nil, err
	}
	return ow.Serialize(), nil
}

type funcHandlerFactory struct {
	handler func(env *Env) interface{}
}

func (f *funcHandlerFactory) New(env types.Environment, funcName string) (types.FuncHandler, error) {
	return &funcHandlerWrapper{
		fnName:  funcName,
		handler: f.handler,
		env:     env,
	}, nil
}

func (f *funcHandlerFactory) GrpcNew(env types.Environment, service string) (types.GrpcFuncHandler, error) {
	return nil, fmt.Errorf("Not implemented")
}

func CreateFuncHandlerFactory(f func(env *Env) interface{}) types.FuncHandlerFactory {
	return &funcHandlerFactory{handler: f}
}
