package core

import (
	"github.com/eniac/Beldi/pkg/cayonlib"
	"github.com/mitchellh/mapstructure"
)

type ProfileInput struct {
	Username string
}

type ProfileOutput struct {
	Followers []string
}

func Profile(env *cayonlib.Env, input ProfileInput) ProfileOutput {
	userInfo := cayonlib.Read(env, Tuser(), input.Username)
	var user User
	cayonlib.CHECK(mapstructure.Decode(userInfo, &user))
	return ProfileOutput{Followers: user.Followers}
}
