package handlers

import (
	"context"
	"encoding/gob"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"cs.utexas.edu/zjia/faas/types"
)

type FollowInput struct {
	UserId     string `json:"userId"`
	FolloweeId string `json:"followeeId"`
	Unfollow   bool   `json:"unfollow,omitempty"`
	// Retry      bool   `json:"retry,omitempty"`
}

type FollowOutput struct {
	Success bool   `json:"success"`
	Message string `json:"message,omitempty"`
}

type FollowObject struct {
	UserId string
	Prev   int
	Next   int
}

type FollowMap struct {
	Str2Id    map[string]int
	Map       map[int]*FollowObject
	MaxActive int
	NextId    int
}

func NewFollowMap(maxActive int) Updatable {
	sentinel := &FollowObject{"", 0, 0}
	fm := &FollowMap{
		Str2Id:    make(map[string]int),
		Map:       make(map[int]*FollowObject),
		MaxActive: maxActive,
		NextId:    1,
	}
	fm.Map[0] = sentinel
	return fm
}

func (fm *FollowMap) Update(params WriteOp) (interface{}, error) {
	targetId := params["key"].(string)
	if params["delete"] == true {
		fm.delete(targetId)
	} else {
		fm.insert(targetId)
	}
	return nil, nil
}

func (fm *FollowMap) Len() int {
	return len(fm.Map) - 1
}

// func (fm *FollowMap) Begin() *FollowObject {
// 	return fm.Map[fm.Map[0].Next]
// }

// func (fm *FollowMap) End() *FollowObject {
// 	return fm.Map[0]
// }

// func (fm *FollowMap) Next(obj *FollowObject) *FollowObject {
// 	return fm.Map[obj.Next]
// }

func (fm *FollowMap) insert(targetId string) bool {
	if _, ok := fm.Str2Id[targetId]; ok {
		return false
	}
	id := fm.NextId
	fm.NextId++
	fm.Str2Id[targetId] = id
	obj := &FollowObject{UserId: targetId}
	fm.Map[id] = obj
	sentinel := fm.Map[0]
	obj.Prev = sentinel.Prev
	obj.Next = 0
	prev := fm.Map[sentinel.Prev]
	prev.Next = id
	sentinel.Prev = id
	if len(fm.Map) > fm.MaxActive {
		oldest := fm.Map[sentinel.Next]
		fm.delete(oldest.UserId)
	}
	return true
}

func (fm *FollowMap) delete(targetId string) bool {
	if id, ok := fm.Str2Id[targetId]; ok {
		delete(fm.Str2Id, targetId)
		obj := fm.Map[id]
		delete(fm.Map, id)
		prev := fm.Map[obj.Prev]
		prev.Next = obj.Next
		next := fm.Map[obj.Next]
		next.Prev = obj.Prev
		return true
	}
	return false
}

func init() {
	gob.Register(&FollowMap{})
}

type followHandler struct{ OCCStore }

func NewFollowHandler(env types.Environment) types.FuncHandler {
	h := &followHandler{}
	h.env = env
	return h
}

func (h *followHandler) follow(srcId string, dstId string) *FollowOutput {
	txnId, snapShot := h.TxnStart()
	// if err != nil {
	// 	log.Fatalf("[FATAL] RuntimeError in TxnStart: %s", err.Error())
	// }
	srcKey := NameHash(fmt.Sprintf("%s.followees", srcId))
	dstKey := NameHash(fmt.Sprintf("%s.followers", dstId))
	view, err := h.CCSyncTo(srcKey, snapShot)
	if err != nil {
		log.Fatalf("[FATAL] callId %v failed to sync to seqnum %v tag %v(%s.followees): %s", h.callId, snapShot, srcKey, srcId, err.Error())
	}
	if view == nil {
		log.Fatalf("[FATAL] callId %v no initial view from seqnum %v tag %v(%s.followees)", h.callId, snapShot, srcKey, srcId)
	}
	// if nLogs > 20 {
	// 	log.Printf("[WARN] %v logs ahead of seqnum %v (%s.followees)", nLogs, snapShot, srcId)
	// }
	followeesView := view.(*FollowMap)
	if updated := followeesView.insert(dstId); !updated {
		return &FollowOutput{Success: true, Message: "Already following"}
	}
	logRecords := WriteLog{}
	logRecords[srcKey] = WriteOp{
		"key": dstId,
	}
	logRecords[dstKey] = WriteOp{
		"key": srcId,
	}
	encoded, err := Encode(logRecords)
	if err != nil {
		log.Fatalf("follow: Failed to encode log records: %v", err)
	}
	_, result := h.TxnCommit(txnId, snapShot, nil, []uint64{srcKey, dstKey}, encoded)
	// log.Printf("[FOLLOW] callId %v txnId %x seqNum %v result %v", h.callId, txnId, commitSeq, result)
	if result {
		return &FollowOutput{Success: true}
	}
	return nil
}

func (h *followHandler) unfollow(srcId string, dstId string) *FollowOutput {
	txnId, snapShot := h.TxnStart()
	// if err != nil {
	// 	log.Fatalf("[FATAL] RuntimeError in TxnStart: %s", err.Error())
	// }
	srcKey := NameHash(fmt.Sprintf("%s.followees", srcId))
	dstKey := NameHash(fmt.Sprintf("%s.followers", dstId))
	view, err := h.CCSyncTo(srcKey, snapShot)
	if err != nil {
		log.Fatalf("[FATAL] callId %v failed to sync to seqnum %v tag %v(%s.followees): %s", h.callId, snapShot, srcKey, srcId, err.Error())
	}
	if view == nil {
		log.Fatalf("[FATAL] callId %v no initial view from seqnum %v tag %v(%s.followees)", h.callId, snapShot, srcKey, srcId)
	}
	// if nLogs > 20 {
	// 	log.Printf("[WARN] %v logs ahead of seqnum %v (%s.followees)", nLogs, snapShot, srcId)
	// }
	followeesView := view.(*FollowMap)
	if updated := followeesView.delete(dstId); !updated {
		return &FollowOutput{Success: true, Message: "Not following"}
	}
	logRecords := WriteLog{}
	logRecords[srcKey] = WriteOp{
		"key":    dstId,
		"delete": true,
	}
	logRecords[dstKey] = WriteOp{
		"key":    srcId,
		"delete": true,
	}
	encoded, err := Encode(logRecords)
	if err != nil {
		log.Fatalf("follow: Failed to encode log records: %v", err)
	}
	_, result := h.TxnCommit(txnId, snapShot, nil, []uint64{srcKey, dstKey}, encoded)
	// log.Printf("[FOLLOW] callId %v txnId %x seqNum %v result %v", h.callId, txnId, commitSeq, result)
	if result {
		return &FollowOutput{Success: true}
	}
	return nil
}

func (h *followHandler) onRequestOnce(input *FollowInput) *FollowOutput {
	if input.Unfollow {
		return h.unfollow(input.UserId, input.FolloweeId)
	} else {
		return h.follow(input.UserId, input.FolloweeId)
	}
}

func (h *followHandler) onRequest(input *FollowInput) *FollowOutput {
	if input.UserId == input.FolloweeId {
		return &FollowOutput{
			Success: false,
			Message: "userId and followeeId cannot be same",
		}
	}
	numRetry := kMaxRetry
	// if input.Retry {
	// 	numRetry = 15
	// }
	for i := 0; i < numRetry; i++ {
		output := h.onRequestOnce(input)
		if output != nil {
			return output
		}
		time.Sleep(kSleepDuration)
	}

	return &FollowOutput{
		Success: false,
		Message: "txn aborted",
	}
}

func (h *followHandler) Call(ctx context.Context, input []byte) ([]byte, error) {
	h.ctx = ctx
	h.callId = ctx.Value("CallId").(uint32)
	parsedInput := &FollowInput{}
	err := json.Unmarshal(input, parsedInput)
	if err != nil {
		return nil, err
	}
	output := h.onRequest(parsedInput)
	return json.Marshal(output)
}
