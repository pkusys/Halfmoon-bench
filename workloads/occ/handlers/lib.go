package handlers

import (
	"bytes"
	"context"
	"encoding/gob"
	"encoding/json"
	"log"
	"os"

	"cs.utexas.edu/zjia/faas/protocol"
	"cs.utexas.edu/zjia/faas/types"
)

// var SetAllAux bool = true
// var SyncWrite bool = false
var EnableCCPath bool = false

func init() {
	// if os.Getenv("SyncWrite") == "1" {
	// 	SyncWrite = true
	// }
	// if os.Getenv("SetAllAux") == "1" {
	// 	SetAllAux = true
	// }
	if os.Getenv("EnableCCPath") == "1" {
		EnableCCPath = true
	}
	gob.Register(struct{}{})
	gob.Register(map[uint64]interface{}{})
	gob.Register(SLogTxnCommitRecord{})
}

type TestInput struct {
	ReadSet  []uint64 `json:"readset"`
	WriteSet []uint64 `json:"writeset"`
}

type TestOutput struct {
	Success bool   `json:"success"`
	Message string `json:"message,omitempty"`
}

type SLogTxnCommitRecord struct {
	Snapshot uint64
	SeqNum   uint64
	ReadSet  []uint64
	WriteSet []uint64
	Data     interface{}
	AuxData  map[uint64]interface{}
}

type OCCStore struct {
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

func (le *OCCStore) TxnStart() (txnId uint64, seqNum uint64, err error) {
	txnId, seqNum, err = le.env.SLogCCTxnStart(le.ctx)
	if err != nil {
		log.Fatalf("[FATAL] RuntimeError in TxnStart: %s", err.Error())
	}
	// log.Printf("TxnStart: txnId %x seqNum %x", txnId, seqNum)
	return
}

func (le *OCCStore) TxnCommit(txnId uint64, snapShot uint64, readSet []uint64, writeSet []uint64, payload []byte) (uint64, bool) {
	hdr := protocol.NewTxnCommitHeader(txnId, snapShot, readSet, writeSet)
	seqnum, err := le.env.SLogCCTxnCommit(le.ctx, txnId, snapShot, bytes.Join([][]byte{hdr, payload}, nil))
	if err != nil {
		log.Fatalf("[FATAL] RuntimeError in TxnCommit: %s", err.Error())
	}
	if seqnum == protocol.MaxLogSeqnum {
		return seqnum, false
	}
	// for _, key := range writeSet {
	// 	le.CCSyncTo(key, seqnum)
	// }
	return seqnum, true
}

func (le *OCCStore) setCCAuxData(seqNum uint64, key uint64, value interface{}) error {
	encoded, err := encode(value)
	if err != nil {
		log.Fatalf("[FATAL] fail to encode in setCCAuxData: %s", err.Error())
	}
	return le.env.SLogCCSetAuxData(le.ctx, key, seqNum, encoded)
}

func (le *OCCStore) CCSyncTo(tag uint64, seqNum uint64) error {
	tail := seqNum
	logsAhead := []*types.CCLogEntry{}
	for {
		commitLog, err := le.env.SLogCCReadPrev(le.ctx, tag, tail)
		if err != nil {
			log.Fatalf("[FATAL] RuntimeError: %s", err.Error())
			return err
		}
		if commitLog == nil {
			// log.Printf("sync to head tag %v seqnum %v", tag, tail)
			break
		}
		if len(commitLog.AuxData) > 0 {
			// log.Printf("sync to cached view tag %v seqnum %v", tag, commitLog.SeqNum)
			break
		}
		tail = commitLog.SeqNum - 1
		logsAhead = append(logsAhead, commitLog)
	}
	if len(logsAhead) > 10 {
		log.Printf("[WARN] %v logs ahead of tag %v seqnum %v", len(logsAhead), tag, seqNum)
	}
	for i := len(logsAhead) - 1; i >= 0; i-- {
		commitLog := logsAhead[i]
		if err := le.setCCAuxData(commitLog.SeqNum, tag, struct{}{}); err != nil {
			log.Fatalf("[FATAL] failed to set aux data at tag %v batch %v: %s", tag, commitLog.SeqNum, err.Error())
			return err
		}
	}
	return nil
}

func (le *OCCStore) CCReadWriteTxn(callId uint32, readSet []uint64, writeSet []uint64) ([]byte, error) {
	txnId, snapShot, err := le.TxnStart()
	if err != nil {
		log.Fatalf("[FATAL] RuntimeError in TxnStart: %s", err.Error())
	}
	log.Printf("callId %v txnId %x started at snapshot %v", callId, txnId, snapShot)
	for _, tag := range readSet {
		// log.Println("CC read sync to: ", snapShot)
		le.CCSyncTo(tag, snapShot)
	}
	log.Printf("callId %v txnId %x snapshot %v sync readset %v", callId, txnId, snapShot, readSet)
	valueBytes, err := encode(struct{}{})
	if err != nil {
		log.Printf("[WARN] failed to encode data: %s", err.Error())
		return nil, err
	}
	batchId, committed := le.TxnCommit(txnId, snapShot, readSet, writeSet, valueBytes)
	log.Printf("callId %v TxnCommit: txnId %x snapShot %v committed %v seqnum %v", callId, txnId, snapShot, committed, batchId)
	if !committed {
		return json.Marshal(TestOutput{false, "txn aborted"})
	}
	return json.Marshal(TestOutput{Success: true})
}

func (le *OCCStore) SLogReadWriteTxn(callId uint32, readSet []uint64, writeSet []uint64) ([]byte, error) {
	// logTail, err := le.env.SharedLogCheckTail(le.ctx, 0)
	data, err := encode(struct{}{})
	if err != nil {
		log.Fatalf("[FATAL] fail to encode txn start in SLogReadWriteTxn: %s", err.Error())
	}
	snapShot, err := le.env.SharedLogAppend(le.ctx, []uint64{1}, data)
	if err != nil {
		log.Fatalf("[FATAL] fail to append txn start in SLogReadWriteTxn: %s", err.Error())
	}
	log.Printf("txnId %x sync readset %v", snapShot, readSet)
	for _, tag := range readSet {
		// log.Println("read sync to: ", snapShot)
		le.SyncTo(tag, snapShot)
	}
	commitRecord := SLogTxnCommitRecord{Snapshot: snapShot, ReadSet: readSet, WriteSet: writeSet, Data: struct{}{}}
	encoded, err := encode(commitRecord)
	if err != nil {
		log.Fatalf("[FATAL] failed to encode data in Txn commit: %s", err.Error())
		return nil, err
	}
	commitSeq, err := le.env.SharedLogAppend(le.ctx, writeSet, encoded)
	if err != nil {
		log.Fatalf("[FATAL] Failed to append commit record: %s", err.Error())
	}
	commitRecord.SeqNum = commitSeq
	committed, err := commitRecord.checkTxnCommitResult(le.ctx, le.env)
	log.Printf("txn %x committed %v at %x", snapShot, committed, commitSeq)
	if err != nil {
		log.Fatalf("[FATAL] Failed to check commit result: %s", err.Error())
	}
	if !committed {
		return json.Marshal(TestOutput{false, "txn aborted"})
	}
	return json.Marshal(TestOutput{Success: true})
}

func decodeCommitLog(logEntry *types.LogEntry) *SLogTxnCommitRecord {
	// log := SLogTxnCommitRecord{}
	decoded, err := decode(logEntry.Data)
	if err != nil {
		log.Fatalf("[WARN] failed to decode data: %s", err.Error())
	}
	commitLog := decoded.(SLogTxnCommitRecord)
	commitLog.SeqNum = logEntry.SeqNum
	if len(logEntry.AuxData) == 0 {
		return &commitLog
	}
	decoded, err = decode(logEntry.AuxData)
	if err != nil {
		log.Fatalf("[WARN] failed to decode aux data: %s", err.Error())
	}
	commitLog.AuxData = decoded.(map[uint64]interface{})
	return &commitLog
}

func (txnCommitLog *SLogTxnCommitRecord) checkTxnCommitResult(ctx context.Context, env types.Environment) (bool, error) {
	if txnCommitLog.AuxData != nil {
		if v, exists := txnCommitLog.AuxData[0]; exists {
			return v.(bool), nil
		}
	} else {
		txnCommitLog.AuxData = make(map[uint64]interface{})
	}
	// log.Printf("[DEBUG] Failed to load txn status: seqNum=%#016x", txnCommitLog.seqNum)
	commitResult := true
	checkedTag := make(map[uint64]bool)
	for _, readKey := range txnCommitLog.ReadSet {
		// tag := objectLogTag(common.NameHash(op.ObjName))
		if _, exists := checkedTag[readKey]; exists {
			continue
		}
		seqNum := txnCommitLog.SeqNum
		for seqNum > txnCommitLog.Snapshot {
			logEntry, err := env.SharedLogReadPrev(ctx, readKey, seqNum-1)
			if err != nil {
				log.Fatalf("[FATAL] RuntimeError in SharedLogReadPrev: %s", err.Error())
				// return false, err
			}
			if logEntry == nil || logEntry.SeqNum <= txnCommitLog.Snapshot {
				break
			}
			seqNum = logEntry.SeqNum
			commitLog := decodeCommitLog(logEntry)
			// if !txnCommitLog.writeSetOverlapped(objectLog) {
			// 	continue
			// }
			if committed, err := commitLog.checkTxnCommitResult(ctx, env); err != nil {
				return false, err
			} else if committed {
				commitResult = false
				break
			}
			// if objectLog.LogType == LOG_NormalOp {
			// 	commitResult = false
			// 	break
			// } else if objectLog.LogType == LOG_TxnCommit {
			// 	if committed, err := objectLog.checkTxnCommitResult(env); err != nil {
			// 		return false, err
			// 	} else if committed {
			// 		commitResult = false
			// 		break
			// 	}
			// }
		}
		if !commitResult {
			break
		}
		checkedTag[readKey] = true
	}
	txnCommitLog.AuxData[0] = commitResult
	encoded, err := encode(txnCommitLog.AuxData)
	if err != nil {
		log.Fatalf("[FATAL] failed to encode aux data: %s", err.Error())
		// return false, err
	}
	env.SharedLogSetAuxData(ctx, txnCommitLog.SeqNum, encoded)
	return commitResult, nil
}

func (le *OCCStore) setAuxData(auxData map[uint64]interface{}, seqNum uint64) error {
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

func (le *OCCStore) SyncTo(tag uint64, seqNum uint64) error {
	// log.Printf("start sync %v to %x", tag, seqNum)
	logsAhead := make([]*SLogTxnCommitRecord, 0, 4)
	tail := seqNum
	head := uint64(0)
	if head == tail {
		return nil
	}
	for tail > head {
		logEntry, err := le.env.SharedLogReadPrev(le.ctx, tag, tail)
		if err != nil {
			log.Fatalf("[FATAL] RuntimeError: %s", err.Error())
		}
		if logEntry == nil || logEntry.SeqNum < head {
			// log.Printf("sync finished <= %v", tail)
			break
		}
		// log.Printf("sync %v %x->%x", tag, seqNum, tail)
		// TODO: check txn commit result here
		commitLog := decodeCommitLog(logEntry)
		committed, err := commitLog.checkTxnCommitResult(le.ctx, le.env)
		if err != nil {
			log.Fatalf("[FATAL] Failed to check commit result: %s", err.Error())
		}
		if committed {
			if _, ok := commitLog.AuxData[tag]; ok {
				// log.Printf("sync %v %v->%v", tag, seqNum, tail)
				// logsAhead = append(logsAhead, &EntryWithAux{logEntry, commitLog.AuxData})
				// log.Printf("sync finished <= %v with aux data", tail)
				break
			}
			logsAhead = append(logsAhead, commitLog)
		}
		tail = logEntry.SeqNum - 1
	}
	for i := len(logsAhead) - 1; i >= 0; i-- {
		commitLog := logsAhead[i]
		// omit the apply phase
		if commitLog.AuxData == nil {
			commitLog.AuxData = make(map[uint64]interface{})
		}
		if _, ok := commitLog.AuxData[tag]; ok {
			log.Fatalf("[FATAL] aux data already exists for tag %v", tag)
		}
		commitLog.AuxData[tag] = struct{}{}
		le.setAuxData(commitLog.AuxData, commitLog.SeqNum)
	}
	return nil
}

type RWTxnHandler struct {
	OCCStore
	Fn func(callId uint32, readSet []uint64, writeSet []uint64) ([]byte, error)
}

func NewRWTxnHandler(env types.Environment) types.FuncHandler {
	h := RWTxnHandler{}
	h.env = env
	if EnableCCPath {
		h.Fn = h.CCReadWriteTxn
	} else {
		h.Fn = h.SLogReadWriteTxn
	}
	return &h
}

func (h *RWTxnHandler) Call(ctx context.Context, input []byte) ([]byte, error) {
	parsedInput := TestInput{}
	if err := json.Unmarshal(input, &parsedInput); err != nil {
		return nil, err
	} else {
		h.ctx = ctx
		callId := ctx.Value("CallId").(uint32)
		return h.Fn(callId, parsedInput.ReadSet, parsedInput.WriteSet)
	}
}

// func (le *OCCStore) Write(tags []uint64, value interface{}) error {
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
// 	// if SyncWrite {
// 	// 	for _, tag := range tags {
// 	// 		le.SyncTo(tag, seqNum-1)
// 	// 	}
// 	// 	auxData := make(map[uint64]interface{})
// 	// 	for _, t := range tags {
// 	// 		auxData[t] = struct{}{}
// 	// 	}
// 	// 	le.setAuxData(auxData, seqNum)
// 	// }
// 	return nil
// }

// func (le *OCCStore) Read(tags []uint64) error {
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

// type TestWriteHandler struct{ OCCStore }

// type TestReadHandler struct{ OCCStore }

// func NewTestWriteHandler(env types.Environment) types.FuncHandler {
// 	wh := TestWriteHandler{}
// 	wh.env = env
// 	return &wh
// }

// func NewTestReadHandler(env types.Environment) types.FuncHandler {
// 	rh := TestReadHandler{}
// 	rh.env = env
// 	return &rh
// }

// func (wh *TestWriteHandler) Call(ctx context.Context, input []byte) ([]byte, error) {
// 	parsedInput := TestInput{}
// 	if err := json.Unmarshal(input, &parsedInput); err != nil {
// 		return nil, err
// 	} else {
// 		wh.ctx = ctx
// 		if err := wh.Write(parsedInput.Keys, struct{}{}); err != nil {
// 			return nil, err
// 		}
// 		return json.Marshal(TestOutput{Success: true})
// 	}
// }

// func (rh *TestReadHandler) Call(ctx context.Context, input []byte) ([]byte, error) {
// 	parsedInput := TestInput{}
// 	if err := json.Unmarshal(input, &parsedInput); err != nil {
// 		return nil, err
// 	} else {
// 		rh.ctx = ctx
// 		if err := rh.Read(parsedInput.Keys); err != nil {
// 			return nil, err
// 		}
// 		return json.Marshal(TestOutput{Success: true})
// 	}
// }
