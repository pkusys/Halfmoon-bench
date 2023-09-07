package core

import (
	"fmt"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/dynamodb/expression"
	"github.com/eniac/Beldi/internal/utils"
	"github.com/eniac/Beldi/pkg/cayonlib"
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

func Post(env *cayonlib.Env, input PostInput) {
	if len(input.Content) == 0 {
		input.Content = utils.RandomString(valueSize)
	}
	timestamp := time.Now().Format("2006-01-02 15:04:05")
	postID := fmt.Sprintf("%v-%v", timestamp, env.InstanceId)
	cayonlib.Write(env, Tpost(), postID, map[expression.NameBuilder]expression.OperandBuilder{
		expression.Name("V"): expression.Value(input.Content),
	}, false)

	userInfo := cayonlib.Read(env, Tuser(), input.Username)
	var user User
	cayonlib.CHECK(mapstructure.Decode(userInfo, &user))
	postInfo := PostInfo{
		Username:  user.Username,
		Timestamp: timestamp,
	}
	cayonlib.AsyncInvoke(env, Tpublish(), RPCInput{
		Function: "Publish",
		Input: aws.JSONValue{
			"Followers": user.Followers,
			"PostID":    postID,
			"Info":      postInfo,
		},
	})
}
