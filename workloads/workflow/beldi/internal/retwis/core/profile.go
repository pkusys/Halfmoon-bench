package core

import (
	"github.com/eniac/Beldi/pkg/beldilib"
	"github.com/mitchellh/mapstructure"
)

type ProfileInput struct {
	Username string
}

type ProfileOutput struct {
	Followers []string
}

func Profile(env *beldilib.Env, input ProfileInput) ProfileOutput {
	userInfo := beldilib.Read(env, Tuser(), input.Username)
	var user User
	beldilib.CHECK(mapstructure.Decode(userInfo, &user))
	return ProfileOutput{Followers: user.Followers}
}
