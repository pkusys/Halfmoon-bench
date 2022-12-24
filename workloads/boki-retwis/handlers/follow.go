package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"cs.utexas.edu/zjia/faas-retwis/utils"

	"cs.utexas.edu/zjia/faas/slib/statestore"
	"cs.utexas.edu/zjia/faas/types"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type FollowInput struct {
	UserId     string `json:"userId"`
	FolloweeId string `json:"followeeId"`
	Unfollow   bool   `json:"unfollow,omitempty"`
	// Retry      bool   `json:"retry,omitempty"`
}

type FollowOutput struct {
	Success bool   `json:"success"`
	Message string `json:"message,omitempty"`
}

type followHandler struct {
	kind   string
	env    types.Environment
	client *mongo.Client
}

func NewSlibFollowHandler(env types.Environment) types.FuncHandler {
	return &followHandler{
		kind: "slib",
		env:  env,
	}
}

func NewMongoFollowHandler(env types.Environment) types.FuncHandler {
	return &followHandler{
		kind:   "mongo",
		env:    env,
		client: utils.CreateMongoClientOrDie(context.TODO()),
	}
}

const kMaxActiveFollows = 16

func followSlib(ctx context.Context, env types.Environment, input *FollowInput) (*FollowOutput, error) {
	txn, err := statestore.CreateTxnEnv(ctx, env)
	if err != nil {
		return nil, err
	}

	userObj1 := txn.Object(fmt.Sprintf("userid:%s", input.UserId))
	if value, _ := userObj1.Get("username"); value.IsNull() {
		txn.TxnAbort()
		return &FollowOutput{
			Success: false,
			Message: fmt.Sprintf("Cannot find user with ID %s", input.UserId),
		}, nil
	}

	userObj2 := txn.Object(fmt.Sprintf("userid:%s", input.FolloweeId))
	if value, _ := userObj2.Get("username"); value.IsNull() {
		txn.TxnAbort()
		return &FollowOutput{
			Success: false,
			Message: fmt.Sprintf("Cannot find user with ID %s", input.FolloweeId),
		}, nil
	}

	if input.Unfollow {
		if value, _ := userObj1.Get("followees"); !value.IsNull() && value.Size() > 0 {
			followees := value.AsArray()
			idx := -1
			for i := range followees {
				if followees[i].(string) == input.FolloweeId {
					idx = i
					break
				}
			}
			if idx >= 0 {
				userObj1.ArrayRemoveAt("followees", statestore.StringValue(input.FolloweeId), idx)
				userObj2.ArrayRemoveChecked("followers", statestore.StringValue(input.UserId))
			}
		}
		// userObj1.ArrayRemoveChecked("followees", statestore.StringValue(input.FolloweeId), idx)
		// userObj1.ArrayRemoveChecked("followers", statestore.StringValue(input.UserId))
	} else {
		if value, _ := userObj1.Get("followees"); !value.IsNull() {
			followers := value.AsArray()
			found := false
			for i := range followers {
				if followers[i].(string) == input.FolloweeId {
					found = true
					break
				}
			}
			if !found {
				userObj1.ArrayPushBackWithLimit("followees", statestore.StringValue(input.FolloweeId), kMaxActiveFollows)
				userObj2.ArrayPushBackChecked("followers", statestore.StringValue(input.UserId), kMaxActiveFollows)
			}
		}
	}

	if committed, err := txn.TxnCommit(); err != nil {
		return nil, err
	} else if committed {
		return &FollowOutput{
			Success: true,
		}, nil
	} else {
		// return &FollowOutput{
		// 	Success: false,
		// 	Message: "Failed to commit transaction due to conflicts",
		// }, nil
		return nil, nil
	}
}

func followMongo(ctx context.Context, client *mongo.Client, input *FollowInput) (*FollowOutput, error) {
	sess, err := client.StartSession(options.Session())
	if err != nil {
		return nil, err
	}
	defer sess.EndSession(ctx)

	_, err = sess.WithTransaction(ctx, func(sessCtx mongo.SessionContext) (interface{}, error) {
		coll := client.Database("retwis").Collection("users")
		user1Filter := bson.D{{"userId", input.UserId}}
		user2Filter := bson.D{{"userId", input.FolloweeId}}
		var user1Update bson.D
		var user2Update bson.D
		if input.Unfollow {
			user1Update = bson.D{{"$unset", bson.D{{fmt.Sprintf("followees.%s", input.FolloweeId), ""}}}}
			user2Update = bson.D{{"$unset", bson.D{{fmt.Sprintf("followers.%s", input.UserId), ""}}}}
		} else {
			user1Update = bson.D{{"$set", bson.D{{fmt.Sprintf("followees.%s", input.FolloweeId), true}}}}
			user2Update = bson.D{{"$set", bson.D{{fmt.Sprintf("followers.%s", input.UserId), true}}}}
		}
		if _, err := coll.UpdateOne(sessCtx, user1Filter, user1Update); err != nil {
			return nil, err
		}
		if _, err := coll.UpdateOne(sessCtx, user2Filter, user2Update); err != nil {
			return nil, err
		}
		return nil, nil
	}, utils.MongoTxnOptions())

	if err != nil {
		return &FollowOutput{
			Success: false,
			Message: fmt.Sprintf("Mongo failed: %v", err),
		}, nil
	}
	return &FollowOutput{
		Success: true,
	}, nil
}

func (h *followHandler) onRequestOnce(ctx context.Context, input *FollowInput) (*FollowOutput, error) {
	switch h.kind {
	case "slib":
		return followSlib(ctx, h.env, input)
	case "mongo":
		return followMongo(ctx, h.client, input)
	default:
		panic(fmt.Sprintf("Unknown kind: %s", h.kind))
	}
}

func (h *followHandler) onRequest(ctx context.Context, input *FollowInput) (*FollowOutput, error) {
	if input.UserId == input.FolloweeId {
		return &FollowOutput{
			Success: false,
			Message: "userId and followeeId cannot be same",
		}, nil
	}
	numRetry := kMaxRetry
	// if input.Retry {
	// 	numRetry = 15
	// }
	for i := 0; i < numRetry; i++ {
		output, err := h.onRequestOnce(ctx, input)
		if err != nil {
			return nil, err
		}
		if output != nil {
			return output, nil
		}
		time.Sleep(kSleepDuration)
	}

	return &FollowOutput{
		Success: false,
		Message: "(follow) Failed to commit transaction due to conflicts",
	}, nil
}

func (h *followHandler) Call(ctx context.Context, input []byte) ([]byte, error) {
	parsedInput := &FollowInput{}
	err := json.Unmarshal(input, parsedInput)
	if err != nil {
		return nil, err
	}
	output, err := h.onRequest(ctx, parsedInput)
	if err != nil {
		return nil, err
	}
	return json.Marshal(output)
}
