package core

import (
	"fmt"
	"sync"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/eniac/Beldi/pkg/cayonlib"
)

type ComposeInput struct {
	Username string
	Password string
	Title    string
	Rating   int
	Text     string
}

func Compose(env *cayonlib.Env, input ComposeInput) {
	reqId := env.InstanceId
	res, instanceId := cayonlib.SyncInvoke(env, TComposeReview(), RPCInput{
		Function: "UploadReq",
		Input:    aws.JSONValue{"reqId": reqId},
	})
	if instanceId == "" || res == nil{
		return
	}
	if res.(float64) != 0 {
		fmt.Println(fmt.Sprintf("DEBUG: result is %s", res))
	}
	invoke1 := cayonlib.ProposeInvoke(env, TUniqueId(), RPCInput{
		Function: "UploadUniqueId2",
		Input:    aws.JSONValue{"reqId": reqId},
	})
	if invoke1 == nil {
		return
	}
	invoke2 := cayonlib.ProposeInvoke(env, TUser(), RPCInput{
		Function: "UploadUser",
		Input:    aws.JSONValue{"reqId": reqId, "username": input.Username},
	})
	if invoke2 == nil {
		return
	}
	invoke3 := cayonlib.ProposeInvoke(env, TMovieId(), RPCInput{
		Function: "UploadMovie",
		Input:    aws.JSONValue{"reqId": reqId, "title": input.Title, "rating": input.Rating},
	})
	if invoke3 == nil {
		return
	}
	invoke4 := cayonlib.ProposeInvoke(env, TText(), RPCInput{
		Function: "UploadText2",
		Input:    aws.JSONValue{"reqId": reqId, "text": input.Text},
	})
	if invoke4 == nil {
		return
	}
	var wg sync.WaitGroup
	wg.Add(4)
	go func() {
		defer wg.Done()
		cayonlib.AssignedSyncInvoke(env, TUniqueId(), invoke1)
	}()
	go func() {
		defer wg.Done()
		cayonlib.AssignedSyncInvoke(env, TUser(), invoke2)
	}()
	go func() {
		defer wg.Done()
		cayonlib.AssignedSyncInvoke(env, TMovieId(), invoke3)
	}()
	go func() {
		defer wg.Done()
		cayonlib.AssignedSyncInvoke(env, TText(), invoke4)
	}()
	wg.Wait()
}
