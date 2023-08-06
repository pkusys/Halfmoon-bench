package core

import (
	"fmt"
	"sync"

	"github.com/aws/aws-sdk-go/service/dynamodb/expression"
	"github.com/eniac/Beldi/internal/utils"
	"github.com/eniac/Beldi/pkg/cayonlib"
	"github.com/mitchellh/mapstructure"
)

var valueSize = 256

// func init() {
// 	if vs, err := strconv.Atoi(os.Getenv("VALUE_SIZE")); err == nil {
// 		valueSize = vs
// 	} else {
// 		panic("invalid VALUE_SIZE")
// 	}
// }

type PostInput struct {
	Username string
	Content  string
}

type PostInfo struct {
	Username  string
	Timestamp uint64
}

func updateTimeline(env *cayonlib.Env, user string, followers []string, postID string) {
	postInfo := PostInfo{
		Username:  user,
		Timestamp: env.SeqNum,
	}
	var wg sync.WaitGroup
	wg.Add(len(followers))
	for i := range followers {
		go func(i int) {
			cayonlib.Write(env, Ttimeline(), followers[i], map[expression.NameBuilder]expression.OperandBuilder{
				expression.Name(fmt.Sprintf("V.Posts.%s", postID)): expression.Value(postInfo),
			}, true)
			wg.Done()
		}(i)
	}
	wg.Wait()
}

func Post(env *cayonlib.Env, input PostInput) {
	if len(input.Content) == 0 {
		input.Content = utils.RandomString(valueSize)
	}
	userInfo := cayonlib.Read(env, Tuser(), input.Username)
	var user User
	cayonlib.CHECK(mapstructure.Decode(userInfo, &user))
	postID := fmt.Sprintf("%d-%d", env.SeqNum, env.StepNumber)
	cayonlib.Write(env, Tpost(), postID, map[expression.NameBuilder]expression.OperandBuilder{
		expression.Name("V"): expression.Value(input.Content),
	}, false)
	updateTimeline(env, user.Username, user.Followers, postID)
}
