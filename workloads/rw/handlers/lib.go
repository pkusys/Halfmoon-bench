package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"time"

	"cs.utexas.edu/zjia/faas/types"
	"github.com/cespare/xxhash/v2"
	"github.com/lithammer/shortuuid"
)

var logMode = "all"

func init() {
	if mode := os.Getenv("LoggingMode"); mode != "" {
		logMode = mode
	}
	if logMode != "all" && logMode != "read" && logMode != "write" {
		log.Fatalf("[FATAL] invalid log mode: %s", logMode)
	}
	log.Printf("[INFO] log mode: %s", logMode)
}

const nLowBits uint64 = 2
const reservedLowBits uint64 = 0
const keyTagLowBits uint64 = 1
const intentStepStreamLowBits uint64 = 2

func IntentStepStreamTag(instanceId string) uint64 {
	h := xxhash.Sum64String(instanceId)
	tag := (h << nLowBits) + intentStepStreamLowBits
	if tag == 0 || (^tag) == 0 {
		panic("Invalid tag")
	}
	return tag
}

func KeyTag(key uint64) uint64 {
	h := xxhash.Sum64String(fmt.Sprintf("k%v", key))
	tag := (h << nLowBits) + keyTagLowBits
	if tag == 0 || (^tag) == 0 {
		panic("Invalid tag")
	}
	return tag
}

type LogEditor struct {
	env                 types.Environment
	ctx                 context.Context
	instanceId          string
	intentStepStreamTag uint64
	step                uint32
	seqnum              uint64
	callId              uint32
}

// func (le *LogEditor) setAuxData(auxData map[uint64]interface{}, seqNum uint64) error {
// 	// auxData := make(map[uint64]interface{})
// 	// for _, tag := range tags {
// 	// 	auxData[tag] = struct{}{}
// 	// }
// 	if auxDataBytes, err := encode(auxData); err != nil {
// 		log.Printf("[WARN] failed to encode aux data: %s", err.Error())
// 		return err
// 	} else if err := le.env.SharedLogSetAuxData(le.ctx, seqNum, auxDataBytes); err != nil {
// 		log.Fatalf("[FATAL] RuntimeError: %s", err.Error())
// 		// return err
// 	}
// 	return nil
// }

// type EntryWithAux struct {
// 	*types.LogEntry
// 	auxData map[uint64]interface{}
// }

func (le *LogEditor) Read(key uint64, seqNum uint64) (int64, error) {
	tag := KeyTag(key)
	log.Printf("[INFO] callId=%v instanceId=%s tag=%v step=%v read key=%v seqnum=%x", le.callId, le.instanceId, le.intentStepStreamTag, le.step, tag, le.seqnum)
	start := time.Now()
	_, err := le.env.SharedLogReadPrev(le.ctx, tag, seqNum)
	if err != nil {
		log.Fatalf("[FATAL] failed to read log: %s", err.Error())
	}
	switch logMode {
	case "all", "read":
		seqnum, err := le.env.SharedLogConditionalAppend(le.ctx, []uint64{}, []byte(fmt.Sprintf("read %v %v", tag, seqNum)), le.intentStepStreamTag, le.step)
		if err != nil {
			log.Fatalf("[FATAL] instanceId=%s tag=%v conflicts on read: %s", le.instanceId, le.intentStepStreamTag, err.Error())
		} else {
			le.seqnum = seqnum
			le.step++
		}
	default:
	}
	elapsed := time.Since(start).Microseconds()
	return elapsed, nil
}

func (le *LogEditor) Write(keys []uint64, value interface{}) (uint64, int64, error) {
	tags := make([]uint64, 0, len(keys))
	for _, k := range keys {
		tags = append(tags, KeyTag(k))
	}
	log.Printf("[INFO] callId=%v instanceId=%s tag=%v step=%v write keys=%v", le.callId, le.instanceId, le.intentStepStreamTag, le.step, tags)
	start := time.Now()
	valueBytes, err := Encode(value)
	if err != nil {
		log.Fatalf("[FATAL] failed to encode data: %s", err.Error())
	}
	seqnum, err := le.env.SharedLogAppend(le.ctx, tags, valueBytes)
	if err != nil {
		log.Fatalf("[FATAL] failed to append log: %s", err.Error())
	}
	switch logMode {
	case "all", "write":
		writelogSeqnum, err := le.env.SharedLogConditionalAppend(le.ctx, []uint64{}, []byte(fmt.Sprintf("write %v %v", tags, seqnum)), le.intentStepStreamTag, le.step)
		if err != nil {
			log.Fatalf("[FATAL] instanceId=%s tag=%v conflicts on write: %s", le.instanceId, le.intentStepStreamTag, err.Error())
		} else {
			le.seqnum = writelogSeqnum
			le.step++
		}
	default:
	}
	elapsed := time.Since(start).Microseconds()
	return seqnum, elapsed, nil
}

func (le *LogEditor) onRequest(input *TestInput) *TestOutput {
	start := time.Now()
	intentStepStreamTag := IntentStepStreamTag(input.InstanceId)
	seqnum, err := le.env.SharedLogConditionalAppend(le.ctx, []uint64{}, []byte("invoke"), intentStepStreamTag, 0)
	if err != nil {
		log.Fatalf("[FATAL] instanceId=%s tag=%v conflicts on invoke: %s", input.InstanceId, intentStepStreamTag, err.Error())
	}
	log.Printf("[INFO] callId=%v instanceId=%s tag=%v start invoked at seqnum=%x", le.callId, input.InstanceId, intentStepStreamTag, seqnum)
	le.instanceId = input.InstanceId
	le.intentStepStreamTag = intentStepStreamTag
	le.step = 1
	le.seqnum = seqnum
	readLatency := make([]int64, 0, len(input.ReadKeys))
	writeLatency := make([]int64, 0, len(input.WriteKeys))
	for _, k := range input.ReadKeys {
		t, _ := le.Read(k, le.seqnum)
		readLatency = append(readLatency, t)
	}
	for _, k := range input.WriteKeys {
		_, t, _ := le.Write([]uint64{k}, struct{}{})
		writeLatency = append(writeLatency, t)
	}
	elapsed := time.Since(start).Microseconds()
	return &TestOutput{
		Success:      true,
		Duration:     elapsed,
		ReadLatency:  readLatency,
		WriteLatency: writeLatency,
	}
}

// func (le *LogEditor) SyncTo(tag uint64, seqNum uint64) error {
// 	log.Printf("start sync %v to %v", tag, seqNum)
// 	logsAhead := make([]*EntryWithAux, 0, 4)
// 	tail := seqNum
// 	head := uint64(0)
// 	if head == tail {
// 		return nil
// 	}
// 	for tail > head {
// 		// if tail != protocol.MaxLogSeqnum {
// 		// 	tail -= 1
// 		// }
// 		logEntry, err := le.env.SharedLogReadPrev(le.ctx, tag, tail)
// 		if err != nil {
// 			log.Fatalf("[FATAL] RuntimeError: %s", err.Error())
// 		}
// 		if logEntry == nil || logEntry.SeqNum < head {
// 			log.Printf("sync finished <= %v", tail)
// 			break
// 		}
// 		log.Printf("sync %v %v->%v", tag, seqNum, tail)
// 		record := &EntryWithAux{LogEntry: logEntry}
// 		if len(logEntry.AuxData) > 0 {
// 			auxData, err := decode(logEntry.AuxData)
// 			if err != nil {
// 				log.Printf("[WARN] failed to decode aux data: %s", err.Error())
// 				return err
// 			}
// 			record.auxData = auxData.(map[uint64]interface{})
// 			if _, ok := record.auxData[tag]; ok {
// 				log.Printf("sync finished <= %v with aux data", tail)
// 				break
// 			}
// 		}
// 		logsAhead = append(logsAhead, record)
// 		tail = logEntry.SeqNum - 1
// 	}
// 	if len(logsAhead) > 0 {
// 		setAuxAt := 0
// 		if SetAllAux {
// 			setAuxAt = len(logsAhead) - 1
// 		}
// 		for setAuxAt >= 0 {
// 			logEntry := logsAhead[setAuxAt]
// 			auxData := logEntry.auxData
// 			if auxData == nil {
// 				auxData = make(map[uint64]interface{})
// 			}
// 			auxData[tag] = struct{}{}
// 			le.setAuxData(auxData, logEntry.SeqNum)
// 			setAuxAt--
// 		}
// 	}
// 	return nil
// }

// func (le *LogEditor) Write(tags []uint64, value interface{}) error {
// 	valueBytes, err := encode(value)
// 	if err != nil {
// 		log.Printf("[WARN] failed to encode data: %s", err.Error())
// 		return err
// 	}
// 	log.Println("write tags: ", tags, "data size: ", len(valueBytes))
// 	seqNum, err := le.env.SharedLogAppend(le.ctx, tags, valueBytes)
// 	if err != nil {
// 		log.Fatalf("[FATAL] RuntimeError: %s", err.Error())
// 	}
// 	log.Println("write at seqnum ", seqNum)
// 	if SyncWrite {
// 		for _, tag := range tags {
// 			le.SyncTo(tag, seqNum-1)
// 		}
// 		auxData := make(map[uint64]interface{})
// 		for _, t := range tags {
// 			auxData[t] = struct{}{}
// 		}
// 		le.setAuxData(auxData, seqNum)
// 	}
// 	return nil
// }

// func (le *LogEditor) Read(tags []uint64) error {
// 	log.Println("read tags: ", tags)
// 	if logTail, err := le.env.SharedLogCheckTail(le.ctx, 0); err != nil {
// 		log.Fatalf("[FATAL] RuntimeError: %s", err.Error())
// 	} else if logTail != nil {
// 		for _, tag := range tags {
// 			log.Println("read sync to: ", logTail.SeqNum)
// 			le.SyncTo(tag, logTail.SeqNum)
// 		}
// 	} else {
// 		log.Println("read tail is nil")
// 	}
// 	return nil
// }

type TestInput struct {
	InstanceId string   `json:"instanceId"`
	ReadKeys   []uint64 `json:"readkeys"`
	WriteKeys  []uint64 `json:"writekeys"`
}

type TestOutput struct {
	Success      bool    `json:"success"`
	Message      string  `json:"message,omitempty"`
	ReadLatency  []int64 `json:"readLatency,omitempty"`
	WriteLatency []int64 `json:"writeLatency,omitempty"`
	Duration     int64   `json:"duration,omitempty"`
}

type TestHandler struct{ LogEditor }

func NewTestHandler(env types.Environment) types.FuncHandler {
	h := TestHandler{}
	h.env = env
	return &h
}

func (h *TestHandler) Call(ctx context.Context, input []byte) ([]byte, error) {
	parsedInput := TestInput{}
	if err := json.Unmarshal(input, &parsedInput); err != nil {
		return nil, err
	} else {
		if parsedInput.InstanceId == "" {
			parsedInput.InstanceId = shortuuid.New()
		}
		h.ctx = ctx
		h.callId = ctx.Value("CallId").(uint32)
		output := h.onRequest(&parsedInput)
		return json.Marshal(output)
	}
}
