package core

import (
	"log"
	"time"

	"github.com/eniac/Beldi/pkg/cayonlib"
	"github.com/mitchellh/mapstructure"
)

type TimelineInput struct {
	Username string
}

type PostData struct {
	PostID    string
	Username  string
	Content   string
	Timestamp string
}

func GetTimeline(env *cayonlib.Env, input TimelineInput, maxReturnPosts int) []PostData {
	item := cayonlib.Read(env, Ttimeline(), input.Username)
	itemMap, ok := item.(map[string]interface{})
	if !ok {
		return nil
	}
	posts, ok := itemMap["Posts"]
	timeline := []PostData{}
	if !ok {
		return timeline
	}
	for k, v := range posts.(map[string]interface{}) {
		var postInfo PostInfo
		cayonlib.CHECK(mapstructure.Decode(v, &postInfo))
		timeline = append(timeline, PostData{
			PostID:    k,
			Username:  postInfo.Username,
			Timestamp: postInfo.Timestamp,
		})
		if len(timeline) >= maxReturnPosts {
			break
		}
	}
	log.Printf("%s has %d posts in timeline", input.Username, len(timeline))
	start := time.Now()
	for i := range timeline {
		post := cayonlib.Read(env, Tpost(), timeline[i].PostID)
		timeline[i].Content = post.(string)
	}
	elapsed := time.Since(start).Milliseconds()
	log.Printf("%s reads %d posts in %d ms", input.Username, len(timeline), elapsed)
	// var wg sync.WaitGroup
	// wg.Add(len(timeline))
	// for i := range timeline {
	// 	go func(i int) {
	// 		post := cayonlib.Read(env, Tpost(), timeline[i].PostID)
	// 		timeline[i].Content = post.(string)
	// 		wg.Done()
	// 	}(i)
	// }
	// wg.Wait()
	return timeline
}
