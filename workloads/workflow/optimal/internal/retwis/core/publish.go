package core

import (
	"fmt"

	"github.com/aws/aws-sdk-go/service/dynamodb/expression"
	"github.com/eniac/Beldi/pkg/cayonlib"
)

type PublishInput struct {
	Followers []string
	PostID    string
	Info      PostInfo
}

func Publish(env *cayonlib.Env, input PublishInput) {
	for i := range input.Followers {
		cayonlib.Write(env, Ttimeline(), input.Followers[i], map[expression.NameBuilder]expression.OperandBuilder{
			expression.Name(fmt.Sprintf("V.Posts.%s", input.PostID)): expression.Value(input.Info),
		}, true)
	}
	// var wg sync.WaitGroup
	// wg.Add(len(followers))
	// for i := range followers {
	// 	go func(i int) {
	// 		cayonlib.Write(env, Ttimeline(), followers[i], map[expression.NameBuilder]expression.OperandBuilder{
	// 			expression.Name(fmt.Sprintf("V.Posts.%s", postID)): expression.Value(postInfo),
	// 		}, true)
	// 		wg.Done()
	// 	}(i)
	// }
	// wg.Wait()
}
