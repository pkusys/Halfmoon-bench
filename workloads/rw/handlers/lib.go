package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"time"

	"cs.utexas.edu/zjia/faas/types"
)

var LogReadWrite = false

func init() {
	if os.Getenv("LogReadWrite") == "1" {
		LogReadWrite = true
	}
}

type LogEditor struct {
	env types.Environment
	ctx context.Context
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

func (le *LogEditor) Read(tag uint64, seqNum uint64) (int64, error) {
	start := time.Now()
	_, err := le.env.SharedLogReadPrev(le.ctx, tag, seqNum)
	if err != nil {
		log.Fatalf("[FATAL] failed to read log: %s", err.Error())
	}
	if LogReadWrite {
		le.env.SharedLogAppend(le.ctx, []uint64{1}, []byte(fmt.Sprintf("read %v %v", tag, seqNum)))
	}
	elapsed := time.Since(start).Microseconds()
	return elapsed, nil
}

func (le *LogEditor) Write(tags []uint64, value interface{}) (uint64, int64, error) {
	start := time.Now()
	valueBytes, err := Encode(value)
	if err != nil {
		log.Fatalf("[FATAL] failed to encode data: %s", err.Error())
	}
	seqnum, err := le.env.SharedLogAppend(le.ctx, tags, valueBytes)
	if err != nil {
		log.Fatalf("[FATAL] failed to append log: %s", err.Error())
	}
	if LogReadWrite {
		le.env.SharedLogAppend(le.ctx, []uint64{1}, []byte(fmt.Sprintf("write %v %v", tags, seqnum)))
	}
	elapsed := time.Since(start).Microseconds()
	return seqnum, elapsed, nil
}

func (le *LogEditor) onRequest(input *TestInput) *TestOutput {
	start := time.Now()
	seqnum, err := le.env.SharedLogAppend(le.ctx, []uint64{1}, []byte("start"))
	if err != nil {
		log.Fatalf("[FATAL] failed to append start log: %s", err.Error())
	}
	readLatency := make([]int64, 0, len(input.ReadKeys))
	writeLatency := make([]int64, 0, len(input.WriteKeys))
	for _, k := range input.ReadKeys {
		t, _ := le.Read(k, seqnum)
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
	ReadKeys  []uint64 `json:"readkeys"`
	WriteKeys []uint64 `json:"writekeys"`
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
		h.ctx = ctx
		output := h.onRequest(&parsedInput)
		return json.Marshal(output)
	}
}
