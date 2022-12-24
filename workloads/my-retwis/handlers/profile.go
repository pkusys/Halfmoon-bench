package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"log"

	"cs.utexas.edu/zjia/faas/types"
)

type ProfileInput struct {
	UserId string `json:"userId"`
}

type ProfileOutput struct {
	Success bool   `json:"success"`
	Message string `json:"message,omitempty"`
	// UserName     string `json:"username,omitempty"`
	NumFollowers int `json:"numFollowers"`
	NumFollowees int `json:"numFollowees"`
	NumPosts     int `json:"numPosts"`
}

type profileHandler struct{ OCCStore }

func NewProfileHandler(env types.Environment) types.FuncHandler {
	h := &profileHandler{}
	h.env = env
	return h
}

func (h *profileHandler) Call(ctx context.Context, input []byte) ([]byte, error) {
	h.ctx = ctx
	h.callId = ctx.Value("CallId").(uint32)
	parsedInput := &ProfileInput{}
	err := json.Unmarshal(input, parsedInput)
	if err != nil {
		return nil, err
	}
	output := h.profile(parsedInput)
	return json.Marshal(output)
}

func (h *profileHandler) profile(input *ProfileInput) *ProfileOutput {
	snapShot, err := h.SnapShot()
	if err != nil {
		log.Fatalf("[FATAL] RuntimeError in SnapShot: %s", err.Error())
	}
	followersKey := NameHash(fmt.Sprintf("%s.followers", input.UserId))
	followeesKey := NameHash(fmt.Sprintf("%s.followees", input.UserId))
	postsKey := NameHash(fmt.Sprintf("%s.posts", input.UserId))
	numFollowers, numFollowees, numPosts := 0, 0, 0
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
	numFollowers = followersView.(*FollowMap).Len()
	followeesView, err := h.CCSyncTo(followeesKey, snapShot)
	if err != nil {
		log.Fatalf("[FATAL] callId %v failed to sync to seqnum %v tag %v(%s.followees): %s", h.callId, snapShot, followeesKey, input.UserId, err.Error())
	}
	if followeesView == nil {
		log.Fatalf("[FATAL] callId %v no initial view from seqnum %v tag %v(%s.followees)", h.callId, snapShot, followeesKey, input.UserId)
	}
	// if nLogs > 20 {
	// 	log.Printf("[WARN] %v logs ahead of seqnum %v (%s.followees)", nLogs, snapShot, input.UserId)
	// }
	numFollowees = followeesView.(*FollowMap).Len()
	postsView, err := h.CCSyncTo(postsKey, snapShot)
	if err != nil {
		log.Fatalf("[FATAL] callId %v failed to sync to seqnum %v tag %v(%s.posts): %s", h.callId, snapShot, postsKey, input.UserId, err.Error())
	}
	if postsView == nil {
		log.Fatalf("[FATAL] callId %v no initial view from seqnum %v tag %v(%s.posts)", h.callId, snapShot, postsKey, input.UserId)
	}
	// if nLogs > 20 {
	// 	log.Printf("[WARN] %v logs ahead of seqnum %v (%s.posts)", nLogs, snapShot, input.UserId)
	// }
	numPosts = postsView.(*PostList).Len()
	return &ProfileOutput{
		Success:      true,
		NumFollowers: numFollowers,
		NumFollowees: numFollowees,
		NumPosts:     numPosts,
	}
}
