package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strconv"
	"time"

	"cs.utexas.edu/zjia/faas-retwis/utils"

	"cs.utexas.edu/zjia/faas/slib/statestore"
	"cs.utexas.edu/zjia/faas/types"

	"go.mongodb.org/mongo-driver/mongo"
)

const kSleepDuration = 0 * time.Millisecond

var kMaxRetry = 10

func init() {
	if n, err := strconv.Atoi(os.Getenv("MAX_RETRY")); err == nil {
		kMaxRetry = n
	}
}

type InitOutput struct {
	Success bool   `json:"success"`
	Message string `json:"message,omitempty"`
}

type initHandler struct {
	kind   string
	env    types.Environment
	client *mongo.Client
}

func NewSlibInitHandler(env types.Environment) types.FuncHandler {
	return &initHandler{
		kind: "slib",
		env:  env,
	}
}

func NewMongoInitHandler(env types.Environment) types.FuncHandler {
	return &initHandler{
		kind:   "mongo",
		env:    env,
		client: utils.CreateMongoClientOrDie(context.TODO()),
	}
}

func initSlib(ctx context.Context, env types.Environment) error {
	store := statestore.CreateEnv(ctx, env)

	if result := store.Object("timeline").MakeArray("posts", 0); result.Err != nil {
		return result.Err
	}

	if result := store.Object("next_user_id").SetNumber("value", 1); result.Err != nil {
		return result.Err
	}

	return nil
}

func initMongo(ctx context.Context, client *mongo.Client) error {
	db := client.Database("retwis")

	if err := utils.MongoCreateCounter(ctx, db, "next_user_id"); err != nil {
		return err
	}

	if err := utils.MongoCreateIndex(ctx, db.Collection("users"), "userId", true /* unique */); err != nil {
		return err
	}

	if err := utils.MongoCreateIndex(ctx, db.Collection("users"), "username", true /* unique */); err != nil {
		return err
	}

	return nil
}

func (h *initHandler) Call(ctx context.Context, input []byte) ([]byte, error) {
	var err error
	switch h.kind {
	case "slib":
		err = initSlib(ctx, h.env)
	case "mongo":
		err = initMongo(ctx, h.client)
	default:
		panic(fmt.Sprintf("Unknown kind: %s", h.kind))
	}

	if err != nil {
		return nil, err
	} else {
		return json.Marshal(&InitOutput{Success: true})
	}
}
