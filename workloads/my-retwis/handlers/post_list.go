package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"log"

	"cs.utexas.edu/zjia/faas/types"
)

type PostListInput struct {
	UserId string `json:"userId,omitempty"`
	Skip   int    `json:"skip,omitempty"`
}

type PostListOutput struct {
	Success bool          `json:"success"`
	Message string        `json:"message,omitempty"`
	Posts   []interface{} `json:"posts,omitempty"`
}

const MaxReturnPosts = 8

// func init() {
// 	if n, err := strconv.Atoi(os.Getenv("MAX_RETURN_POSTS")); err == nil {
// 		MaxReturnPosts = n
// 	}
// }

type postListHandler struct{ OCCStore }

func NewPostListHandler(env types.Environment) types.FuncHandler {
	h := &postListHandler{}
	h.env = env
	return h
}

func (h *postListHandler) Call(ctx context.Context, input []byte) ([]byte, error) {
	h.ctx = ctx
	h.callId = ctx.Value("CallId").(uint32)
	parsedInput := &PostListInput{}
	err := json.Unmarshal(input, parsedInput)
	if err != nil {
		return nil, err
	}
	output := h.postList(parsedInput)
	return json.Marshal(output)
}

func (h *postListHandler) postList(input *PostListInput) *PostListOutput {
	snapShot, err := h.SnapShot()
	if err != nil {
		log.Fatalf("[FATAL] RuntimeError in SnapShot: %s", err.Error())
	}
	postsKey := NameHash(fmt.Sprintf("%s.posts", input.UserId))
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
	postList := postsView.(*PostList)
	log.Printf("callId %v User %s has %d posts", h.callId, input.UserId, postList.Len())
	posts := make([]interface{}, 0, MaxReturnPosts)
	for i := postList.Len() - 1; i >= 0; i-- {
		postId := postList.PostIds[i]
		postKey := NameHash(postId)
		postView, err := h.CCSyncTo(postKey, snapShot)
		if err != nil {
			log.Fatalf("[FATAL] callId %v failed to sync to seqnum %v tag %v(postId: %s): %s", h.callId, snapShot, postKey, postId, err.Error())
		}
		if postView == nil {
			log.Fatalf("[POSTLIST] Post %s not found", postId)
		}
		post := postView.(PostRecord)
		returnedPost := map[string]string{
			"userId": post.UserId,
			"body":   post.Body,
		}
		posts = append(posts, returnedPost)
		if len(posts) >= MaxReturnPosts {
			break
		}
	}
	return &PostListOutput{
		Success: true,
		Posts:   posts,
	}
}
