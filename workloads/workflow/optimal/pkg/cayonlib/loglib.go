package cayonlib

import (
	"encoding/json"
	"fmt"
	"log"

	"cs.utexas.edu/zjia/faas/protocol"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/golang/snappy"
	// "context"
)

type IntentLogEntry struct {
	SeqNum     uint64        `json:"-"`
	InstanceId string        `json:"instanceId"`
	StepNumber int32         `json:"step"`
	PostStep   bool          `json:"postStep"`
	Data       aws.JSONValue `json:"data"`
}

type IntentFsm struct {
	instanceId string
	stepNumber int32
	tail       *IntentLogEntry
	stepLogs   map[int32]*IntentLogEntry
	// postStepLogs map[int32]*IntentLogEntry
}

func NewIntentFsm(instanceId string) *IntentFsm {
	return &IntentFsm{
		instanceId: instanceId,
		stepNumber: 0,
		tail:       nil,
		stepLogs:   make(map[int32]*IntentLogEntry),
		// postStepLogs: make(map[int32]*IntentLogEntry),
	}
}

func (fsm *IntentFsm) applyLog(intentLog *IntentLogEntry) {
	fsm.tail = intentLog
	step := intentLog.StepNumber
	if _, exists := fsm.stepLogs[step]; !exists {
		if step != fsm.stepNumber {
			panic(fmt.Sprintf("StepNumber is not monotonic: expected=%d, seen=%d", fsm.stepNumber, step))
		}
		fsm.stepNumber += 1
		fsm.stepLogs[step] = intentLog
	}
}

func (fsm *IntentFsm) Catch(env *Env) {
	tag := IntentStepStreamTag(fsm.instanceId)
	seqNum := uint64(0)
	if fsm.tail != nil {
		seqNum = fsm.tail.SeqNum + 1
	}
	for {
		logEntry, err := env.FaasEnv.SharedLogReadNext(env.FaasCtx, tag, seqNum)
		CHECK(err)
		if logEntry == nil {
			break
		}
		decoded, err := snappy.Decode(nil, logEntry.Data)
		CHECK(err)
		var intentLog IntentLogEntry
		err = json.Unmarshal(decoded, &intentLog)
		CHECK(err)
		if intentLog.InstanceId == fsm.instanceId {
			// log.Printf("[INFO] Found my log: seqnum=%d, step=%d", logEntry.SeqNum, intentLog.StepNumber)
			intentLog.SeqNum = logEntry.SeqNum
			fsm.applyLog(&intentLog)
			env.LogSize += len(logEntry.Data)
		} else {
			log.Fatalf("[FATAL] Hash collision on intent step stream of instance %v tag %v: other instance=%v step=%d seqnum=%d", fsm.instanceId, tag, intentLog.InstanceId, intentLog.StepNumber, logEntry.SeqNum)
		}
		seqNum = logEntry.SeqNum + 1
	}
}

func (fsm *IntentFsm) GetStepLog(stepNumber int32) *IntentLogEntry {
	if log, exists := fsm.stepLogs[stepNumber]; exists {
		return log
	} else {
		return nil
	}
}

func ProposeNextStep(env *Env, tags []uint64, data aws.JSONValue) (bool, *IntentLogEntry) {
	step := env.StepNumber
	env.StepNumber += 1
	intentLog := env.Fsm.GetStepLog(step)
	if intentLog != nil {
		env.SeqNum = intentLog.SeqNum
		return false, intentLog
	}
	intentLog = &IntentLogEntry{
		InstanceId: env.InstanceId,
		StepNumber: step,
		PostStep:   false,
		Data:       data,
	}
	seqNum, condOK := LibConditionalAppendLog(env, tags, &intentLog, IntentStepStreamTag(env.InstanceId), uint32(step))
	if !condOK {
		log.Printf("[WARN] Found concurrent intent step log: instance=%v step=%d", env.InstanceId, step)
		env.Instruction = "EXIT"
		return false, nil
	}
	intentLog.SeqNum = seqNum
	env.Fsm.applyLog(intentLog)
	env.SeqNum = intentLog.SeqNum
	return true, intentLog
	// if condOK {
	// 	intentLog.SeqNum = seqNum
	// 	env.Fsm.applyLog(intentLog)
	// } else {
	// 	log.Printf("[INFO] Found concurrent intent step log: instance=%v step=%d", env.InstanceId, step)
	// 	env.Fsm.Catch(env)
	// 	intentLog = env.Fsm.GetStepLog(step)
	// }
	// env.SeqNum = intentLog.SeqNum
	// return condOK, intentLog
}

func LogStepResultForCaller(env *Env, instanceId string, stepNumber int32, data aws.JSONValue) {
	serializedData, err := json.Marshal(data)
	CHECK(err)
	env.LogSize += len(serializedData)
	encoded := snappy.Encode(nil, serializedData)
	err = env.FaasEnv.SharedLogOverwrite(env.FaasCtx, IntentStepStreamTag(instanceId), uint32(stepNumber), encoded)
	CHECK(err)
}

func LibConditionalAppendLog(env *Env, tags []uint64, data interface{}, condTag uint64, condPos uint32) (uint64, bool) {
	serializedData, err := json.Marshal(data)
	CHECK(err)
	encoded := snappy.Encode(nil, serializedData)
	seqNum, err := env.FaasEnv.SharedLogConditionalAppend(env.FaasCtx, tags, encoded, condTag, condPos)
	if seqNum == protocol.InvalidLogSeqnum {
		panic(err)
	}
	if err == nil {
		env.LogSize += len(serializedData)
	}
	return seqNum, err == nil // if not nil then conditon failed but seqnum is valid
}

func LibAppendLog(env *Env, tag uint64, data interface{}) uint64 {
	serializedData, err := json.Marshal(data)
	CHECK(err)
	env.LogSize += len(serializedData)
	encoded := snappy.Encode(nil, serializedData)
	seqNum, err := env.FaasEnv.SharedLogAppend(env.FaasCtx, []uint64{tag}, encoded)
	CHECK(err)
	return seqNum
}

func CheckLogDataField(intentLog *IntentLogEntry, field string, expected string) {
	if tmp := intentLog.Data[field].(string); tmp != expected {
		panic(fmt.Sprintf("Field %s mismatch: expected=%s, have=%s", field, expected, tmp))
	}
}
