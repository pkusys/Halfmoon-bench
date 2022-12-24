package handlers

import (
	"context"
	"encoding/gob"
	"encoding/json"
	"fmt"
	"log"
	"math/rand"
	"time"

	"cs.utexas.edu/zjia/faas/types"
)

type PostInput struct {
	UserId string `json:"userId"`
	Body   string `json:"body"`
	Notify int    `json:"notify"`
}

type PostOutput struct {
	Success bool   `json:"success"`
	Message string `json:"message,omitempty"`
}

type PostRecord struct {
	Id     string
	UserId string
	Body   string
}

func (pr PostRecord) Update(params WriteOp) (interface{}, error) {
	return nil, nil
}

type PostList struct {
	MaxPosts int
	PostIds  []string
}

func NewPostList(maxPosts int) Updatable {
	return &PostList{
		MaxPosts: maxPosts,
		PostIds:  make([]string, 0, maxPosts),
	}
}

func (pl *PostList) Len() int {
	return len(pl.PostIds)
}

func (pl *PostList) Update(params WriteOp) (interface{}, error) {
	postId := params["postId"].(string)
	pl.PostIds = append(pl.PostIds, postId)
	if len(pl.PostIds) > pl.MaxPosts {
		pl.PostIds = pl.PostIds[1:]
	}
	return nil, nil
}

type postHandler struct{ OCCStore }

func NewPostHandler(env types.Environment) types.FuncHandler {
	h := &postHandler{}
	h.env = env
	return h
}

func init() {
	gob.Register(PostRecord{})
	gob.Register(&PostList{})
}

func (h *postHandler) post(input *PostInput) *PostOutput {
	txnId, snapShot := h.TxnStart()
	// if err != nil {
	// 	log.Fatalf("[FATAL] RuntimeError in TxnStart: %s", err.Error())
	// }
	followersKey := NameHash(fmt.Sprintf("%s.followers", input.UserId))
	followersView, err := h.CCSyncTo(followersKey, snapShot)
	if err != nil {
		log.Fatalf("[FATAL] callId %v failed to sync to seqnum %v tag %v(%s.followers): %s", h.callId, snapShot, followersKey, input.UserId, err.Error())
	}
	if followersView == nil {
		log.Fatalf("[FATAL] callId %v no initial view from seqnum %v tag %v(%s.followers)", h.callId, snapShot, followersKey, input.UserId)
	}
	// if nLogs > 20 {
	// 	log.Printf("[WARN] %v logs ahead of seqnum %v (%s.followers)", nLogs, snapShot, input.UserId)
	// }
	followersMap := followersView.(*FollowMap)
	if followersMap.Len() == 0 {
		log.Printf("[WARN] no followers for %s", input.UserId)
		return &PostOutput{Success: true, Message: "no followers"}
	}
	logRecords := WriteLog{}
	writeSet := make([]uint64, 0, followersMap.Len()+1)
	// generate post log record
	postId := fmt.Sprintf("%016x", h.env.GenerateUniqueID())
	postKey := NameHash(postId)
	writeSet = append(writeSet, postKey)
	logRecords[postKey] = WriteOp{
		"view": PostRecord{Id: postId, UserId: input.UserId, Body: input.Body},
	}
	// select k followers
	followers := make([]uint64, 0, followersMap.Len())
	// for obj := followersMap.Begin(); obj != followersMap.End(); obj = followersMap.Next(obj) {
	for userId := range followersMap.Str2Id {
		userKey := NameHash(fmt.Sprintf("%s.posts", userId))
		followers = append(followers, userKey)
	}
	rand.Shuffle(len(followers), func(i, j int) {
		followers[i], followers[j] = followers[j], followers[i]
	})
	if len(followers) > input.Notify {
		followers = followers[0:input.Notify]
	}
	// notify followers
	for _, userKey := range followers {
		writeSet = append(writeSet, userKey)
		logRecords[userKey] = WriteOp{
			"postId": postId,
		}
	}
	encoded, err := Encode(logRecords)
	if err != nil {
		log.Fatalf("follow: Failed to encode log records: %v", err)
	}
	_, result := h.TxnCommit(txnId, snapShot, []uint64{followersKey}, writeSet, encoded)
	// log.Printf("[POST] callId %v txnId %x seqNum %v result %v", h.callId, txnId, commitSeq, result)
	if result {
		return &PostOutput{Success: true}
	}
	return nil // txn aborted
}

func (h *postHandler) onRequest(input *PostInput) *PostOutput {
	for i := 0; i < kMaxRetry; i++ {
		output := h.post(input)
		if output != nil {
			return output
		}
		time.Sleep(kSleepDuration)
	}
	return &PostOutput{
		Success: false,
		Message: "txn aborted",
	}
}

func (h *postHandler) Call(ctx context.Context, input []byte) ([]byte, error) {
	h.ctx = ctx
	h.callId = ctx.Value("CallId").(uint32)
	parsedInput := &PostInput{}
	err := json.Unmarshal(input, parsedInput)
	if err != nil {
		return nil, err
	}
	output := h.onRequest(parsedInput)
	return json.Marshal(output)
}
