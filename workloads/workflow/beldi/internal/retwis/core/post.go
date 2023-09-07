package core

import (
	"fmt"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/dynamodb/expression"
	"github.com/eniac/Beldi/internal/utils"
	"github.com/eniac/Beldi/pkg/beldilib"
	"github.com/mitchellh/mapstructure"
)

var valueSize = 256

type PostInput struct {
	Username string
	Content  string
}

type PostInfo struct {
	Username  string
	Timestamp string
}

func Post(env *beldilib.Env, input PostInput) {
	if len(input.Content) == 0 {
		input.Content = utils.RandomString(valueSize)
	}
	timestamp := time.Now().Format("2006-01-02 15:04:05")
	postID := fmt.Sprintf("%v-%v", timestamp, env.InstanceId)
	beldilib.Write(env, Tpost(), postID, map[expression.NameBuilder]expression.OperandBuilder{
		expression.Name("V"): expression.Value(input.Content),
	})

	userInfo := beldilib.Read(env, Tuser(), input.Username)
	var user User
	beldilib.CHECK(mapstructure.Decode(userInfo, &user))
	postInfo := PostInfo{
		Username:  user.Username,
		Timestamp: timestamp,
	}
	beldilib.AsyncInvoke(env, Tpublish(), RPCInput{
		Function: "Publish",
		Input: aws.JSONValue{
			"Followers": user.Followers,
			"PostID":    postID,
			"Info":      postInfo,
		},
	})
}
