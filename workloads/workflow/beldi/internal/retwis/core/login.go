package core

import (
	"github.com/eniac/Beldi/pkg/beldilib"
	"github.com/mitchellh/mapstructure"
)

type LoginInput struct {
	Username string
	Password string
}

type LoginOutput struct {
	Success bool
}

func Login(env *beldilib.Env, input LoginInput) LoginOutput {
	userInfo := beldilib.Read(env, Tuser(), input.Username)
	var user User
	beldilib.CHECK(mapstructure.Decode(userInfo, &user))
	success := input.Password == user.Password
	return LoginOutput{Success: success}
}
