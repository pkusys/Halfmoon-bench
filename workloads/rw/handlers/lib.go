package handlers

import (
	"bytes"
	"context"
	"encoding/gob"
	"encoding/json"
	"log"
	"os"

	"cs.utexas.edu/zjia/faas/types"
)

var SetAllAux bool = false
var SyncWrite bool = false

func init() {
	if os.Getenv("SyncWrite") == "1" {
		SyncWrite = true
	}
	if os.Getenv("SetAllAux") == "1" {
		SetAllAux = true
	}
	gob.Register(struct{}{})
	gob.Register(map[uint64]interface{}{})
}

type LogEditor struct {
	env types.Environment
	ctx context.Context
}

func decode(raw []byte) (value interface{}, err error) {
	dec := gob.NewDecoder(bytes.NewReader(raw))
	err = dec.Decode(&value)
	return
}

func encode(value interface{}) (raw []byte, err error) {
	buf := bytes.Buffer{}
	enc := gob.NewEncoder(&buf)
	err = enc.Encode(&value)
	raw = buf.Bytes()
	return
}

func jsonDecode(raw []byte) (value interface{}, err error) {
	err = json.Unmarshal(raw, &value)
	return
}

func (le *LogEditor) setAuxData(auxData map[uint64]interface{}, seqNum uint64) error {
	// auxData := make(map[uint64]interface{})
	// for _, tag := range tags {
	// 	auxData[tag] = struct{}{}
	// }
	if auxDataBytes, err := encode(auxData); err != nil {
		log.Printf("[WARN] failed to encode aux data: %s", err.Error())
		return err
	} else if err := le.env.SharedLogSetAuxData(le.ctx, seqNum, auxDataBytes); err != nil {
		log.Fatalf("[FATAL] RuntimeError: %s", err.Error())
		// return err
	}
	return nil
}

type EntryWithAux struct {
	*types.LogEntry
	auxData map[uint64]interface{}
}

func (le *LogEditor) SyncTo(tag uint64, seqNum uint64) error {
	log.Printf("start sync %v to %v", tag, seqNum)
	logsAhead := make([]*EntryWithAux, 0, 4)
	tail := seqNum
	head := uint64(0)
	if head == tail {
		return nil
	}
	for tail > head {
		// if tail != protocol.MaxLogSeqnum {
		// 	tail -= 1
		// }
		logEntry, err := le.env.SharedLogReadPrev(le.ctx, tag, tail)
		if err != nil {
			log.Fatalf("[FATAL] RuntimeError: %s", err.Error())
		}
		if logEntry == nil || logEntry.SeqNum < head {
			log.Printf("sync finished <= %v", tail)
			break
		}
		log.Printf("sync %v %v->%v", tag, seqNum, tail)
		record := &EntryWithAux{LogEntry: logEntry}
		if len(logEntry.AuxData) > 0 {
			auxData, err := decode(logEntry.AuxData)
			if err != nil {
				log.Printf("[WARN] failed to decode aux data: %s", err.Error())
				return err
			}
			record.auxData = auxData.(map[uint64]interface{})
			if _, ok := record.auxData[tag]; ok {
				log.Printf("sync finished <= %v with aux data", tail)
				break
			}
		}
		logsAhead = append(logsAhead, record)
		tail = logEntry.SeqNum - 1
	}
	if len(logsAhead) > 0 {
		setAuxAt := 0
		if SetAllAux {
			setAuxAt = len(logsAhead) - 1
		}
		for setAuxAt >= 0 {
			logEntry := logsAhead[setAuxAt]
			auxData := logEntry.auxData
			if auxData == nil {
				auxData = make(map[uint64]interface{})
			}
			auxData[tag] = struct{}{}
			le.setAuxData(auxData, logEntry.SeqNum)
			setAuxAt--
		}
	}
	return nil
}

func (le *LogEditor) Write(tags []uint64, value interface{}) error {
	valueBytes, err := encode(value)
	if err != nil {
		log.Printf("[WARN] failed to encode data: %s", err.Error())
		return err
	}
	log.Println("write tags: ", tags, "data size: ", len(valueBytes))
	seqNum, err := le.env.SharedLogAppend(le.ctx, tags, valueBytes)
	if err != nil {
		log.Fatalf("[FATAL] RuntimeError: %s", err.Error())
	}
	log.Println("write at seqnum ", seqNum)
	if SyncWrite {
		for _, tag := range tags {
			le.SyncTo(tag, seqNum-1)
		}
		auxData := make(map[uint64]interface{})
		for _, t := range tags {
			auxData[t] = struct{}{}
		}
		le.setAuxData(auxData, seqNum)
	}
	return nil
}

func (le *LogEditor) Read(tags []uint64) error {
	log.Println("read tags: ", tags)
	if logTail, err := le.env.SharedLogCheckTail(le.ctx, 0); err != nil {
		log.Fatalf("[FATAL] RuntimeError: %s", err.Error())
	} else if logTail != nil {
		for _, tag := range tags {
			log.Println("read sync to: ", logTail.SeqNum)
			le.SyncTo(tag, logTail.SeqNum)
		}
	} else {
		log.Println("read tail is nil")
	}
	return nil
}

type TestInput struct {
	Keys []uint64 `json:"keys"`
}

type TestOutput struct {
	Success bool   `json:"success"`
	Message string `json:"message,omitempty"`
}

type TestWriteHandler struct{ LogEditor }

type TestReadHandler struct{ LogEditor }

func NewTestWriteHandler(env types.Environment) types.FuncHandler {
	wh := TestWriteHandler{}
	wh.env = env
	return &wh
}

func NewTestReadHandler(env types.Environment) types.FuncHandler {
	rh := TestReadHandler{}
	rh.env = env
	return &rh
}

func (wh *TestWriteHandler) Call(ctx context.Context, input []byte) ([]byte, error) {
	parsedInput := TestInput{}
	if err := json.Unmarshal(input, &parsedInput); err != nil {
		return nil, err
	} else {
		wh.ctx = ctx
		if err := wh.Write(parsedInput.Keys, struct{}{}); err != nil {
			return nil, err
		}
		return json.Marshal(TestOutput{Success: true})
	}
}

func (rh *TestReadHandler) Call(ctx context.Context, input []byte) ([]byte, error) {
	parsedInput := TestInput{}
	if err := json.Unmarshal(input, &parsedInput); err != nil {
		return nil, err
	} else {
		rh.ctx = ctx
		if err := rh.Read(parsedInput.Keys); err != nil {
			return nil, err
		}
		return json.Marshal(TestOutput{Success: true})
	}
}
