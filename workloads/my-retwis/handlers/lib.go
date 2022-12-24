package handlers

import (
	"bytes"
	"context"
	"fmt"
	"hash/fnv"
	"log"

	"cs.utexas.edu/zjia/faas/protocol"
	"cs.utexas.edu/zjia/faas/types"
)

const LogTagReserveBits = 0

func NameHash(name string) uint64 {
	h := fnv.New64a()
	h.Write([]byte(name))
	hash := h.Sum64()
	return hash << LogTagReserveBits
}

type OCCStore struct {
	env    types.Environment
	ctx    context.Context
	callId uint32
}

func (le *OCCStore) SnapShot() (seqNum uint64, err error) {
	tail, err := le.env.SLogCCReadPrev(le.ctx, 0, protocol.InvalidLogSeqnum)
	if err != nil {
		log.Fatalf("[FATAL] SnapShot callId %v: %s", le.callId, err.Error())
	}
	snapShot := tail.SeqNum
	log.Printf("SnapShot: callId %v start read-only txn at snapshot %v", le.callId, snapShot)
	return snapShot, nil
}

func (le *OCCStore) TxnStart() (uint64, uint64) {
	txnId, seqNum, err := le.env.SLogCCTxnStart(le.ctx)
	if err != nil {
		log.Fatalf("[FATAL] TxnStart callId %v: %s", le.callId, err.Error())
	}
	log.Printf("TxnStart: callId %v txnId %x started at snapshot %v", le.callId, txnId, seqNum)
	return txnId, seqNum
}

func (le *OCCStore) TxnCommit(txnId uint64, snapShot uint64, readSet []uint64, writeSet []uint64, payload []byte) (uint64, bool) {
	hdr := protocol.NewTxnCommitHeader(txnId, snapShot, readSet, writeSet)
	seqNum, err := le.env.SLogCCTxnCommit(le.ctx, txnId, snapShot, bytes.Join([][]byte{hdr, payload}, nil))
	if err != nil {
		log.Fatalf("[FATAL] TxnCommit callId %v txnId %x snapShot %v: %s", le.callId, txnId, snapShot, err.Error())
	}
	committed := true
	if seqNum == protocol.MaxLogSeqnum {
		committed = false
	}
	log.Printf("TxnCommit: callId %v txnId %x snapShot %v committed %v seqNum %v", le.callId, txnId, snapShot, committed, seqNum)
	return seqNum, committed
}

// engine returns auxdata if present, otherwise returns the writelog
func (le *OCCStore) CacheView(seqNum uint64, key uint64, view Updatable) error {
	encoded, err := Encode(view)
	if err != nil {
		return err
	}
	if err := le.env.SLogCCSetAuxData(le.ctx, key, seqNum, encoded); err != nil {
		return err
	}
	return nil
}

func (le *OCCStore) CCSyncTo(tag uint64, seqNum uint64) (Updatable, error) {
	tail := seqNum
	lastSeq := seqNum
	opsAhead := make([]WriteOp, 0, 16)
	var view Updatable
	for {
		commitLog, err := le.env.SLogCCReadPrev(le.ctx, tag, tail)
		if err != nil {
			return nil, err
		}
		if commitLog == nil {
			// log.Printf("sync to head tag %v batch %v", tag, batchId)
			break
		}
		if len(commitLog.AuxData) > 0 {
			// log.Printf("sync to cached view tag %v batch %v", tag, batchId)
			decoded, err := Decode(commitLog.AuxData)
			if err != nil {
				return nil, fmt.Errorf("failed to decode view at seqnum %v: %s", commitLog.SeqNum, err.Error())
			}
			view = decoded.(Updatable)
			break
		}
		data, err := protocol.ParseTxnCommitPayload(commitLog.Data)
		if err != nil {
			return nil, fmt.Errorf("failed to parse txn commit payload at seqnum %v: %s", commitLog.SeqNum, err.Error())
		}
		decoded, err := Decode(data)
		if err != nil {
			return nil, fmt.Errorf("failed to decode write op: %s", err.Error())
		}
		if op, ok := decoded.(WriteLog)[tag]; ok {
			if len(opsAhead) == 0 {
				lastSeq = commitLog.SeqNum
			}
			opsAhead = append(opsAhead, op)
		} else {
			return nil, fmt.Errorf("failed to find write op for tag %v", tag)
		}
		tail = commitLog.SeqNum - 1
	}
	if len(opsAhead) == 0 {
		if view == nil {
			return nil, fmt.Errorf("empty log")
		}
		return view, nil
	}
	// if len(opsAhead) > 10 {
	// 	log.Printf("[WARN] %v logs ahead of tag %v seqnum %v", len(opsAhead), tag, seqNum)
	// }
	if view == nil {
		if initialView, ok := opsAhead[len(opsAhead)-1]["view"]; ok {
			view = initialView.(Updatable)
			opsAhead = opsAhead[:len(opsAhead)-1]
		} else {
			return nil, fmt.Errorf("no initial view")
		}
	}
	for i := len(opsAhead) - 1; i >= 0; i-- {
		op := opsAhead[i]
		view.Update(op)
	}
	if len(opsAhead) > 0 {
		if err := le.CacheView(lastSeq, tag, view); err != nil {
			return nil, fmt.Errorf("failed to cache view: %s", err.Error())
		}
	}
	return view, nil
}
