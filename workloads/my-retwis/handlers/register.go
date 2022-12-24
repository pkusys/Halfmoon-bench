package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"strconv"
	"time"

	"cs.utexas.edu/zjia/faas/types"
)

type RegisterInput struct {
	UserId string `json:"userId"`
}

type RegisterOutput struct {
	Success bool   `json:"success"`
	Message string `json:"message,omitempty"`
}

const kSleepDuration = 0 * time.Millisecond
const kMaxActiveFollows = 16
const kMaxPosts = 24

var kMaxRetry = 10

func init() {
	if n, err := strconv.Atoi(os.Getenv("MAX_RETRY")); err == nil {
		kMaxRetry = n
	}
}

type registerHandler struct{ OCCStore }

func NewRegisterHandler(env types.Environment) types.FuncHandler {
	h := &registerHandler{}
	h.env = env
	return h
}

func (h *registerHandler) Call(ctx context.Context, input []byte) ([]byte, error) {
	h.ctx = ctx
	h.callId = ctx.Value("CallId").(uint32)
	parsedInput := &RegisterInput{}
	err := json.Unmarshal(input, parsedInput)
	if err != nil {
		return nil, err
	}
	output := h.onRequest(parsedInput)
	return json.Marshal(output)
}

func (h *registerHandler) onRequest(input *RegisterInput) *RegisterOutput {
	output := h.register(input)
	if output == nil {
		return &RegisterOutput{
			Success: false,
			Message: "txn aborted",
		}
	}
	return output
}

func (h *registerHandler) register(input *RegisterInput) *RegisterOutput {
	txnId, snapShot := h.TxnStart()
	// if err != nil {
	// 	log.Fatalf("[FATAL] RuntimeError in SnapShot: %s", err.Error())
	// }
	followersKey := NameHash(fmt.Sprintf("%s.followers", input.UserId))
	followersObj := NewFollowMap(kMaxActiveFollows)
	followeesKey := NameHash(fmt.Sprintf("%s.followees", input.UserId))
	followeesObj := NewFollowMap(kMaxActiveFollows)
	postsKey := NameHash(fmt.Sprintf("%s.posts", input.UserId))
	postsObj := NewPostList(kMaxPosts)
	logRecords := WriteLog{}
	logRecords[followersKey] = WriteOp{
		"view": followersObj,
	}
	logRecords[followeesKey] = WriteOp{
		"view": followeesObj,
	}
	logRecords[postsKey] = WriteOp{
		"view": postsObj,
	}
	encoded, err := Encode(logRecords)
	if err != nil {
		log.Fatalf("follow: Failed to encode log records: %v", err)
	}
	_, result := h.TxnCommit(txnId, snapShot, nil, []uint64{followersKey, followeesKey, postsKey}, encoded)
	// log.Printf("[Register] callId %v txnId %x seqNum %v result %v", h.callId, txnId, commitSeq, result)
	if result {
		return &RegisterOutput{Success: true}
	}
	return nil
}
