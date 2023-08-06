package core

import (
	"github.com/eniac/Beldi/pkg/cayonlib"
	"github.com/mitchellh/mapstructure"
)

type LoginInput struct {
	Username string
	Password string
}

type LoginOutput struct {
	Success bool
}

func Login(env *cayonlib.Env, input LoginInput) LoginOutput {
	userInfo := cayonlib.Read(env, Tuser(), input.Username)
	var user User
	cayonlib.CHECK(mapstructure.Decode(userInfo, &user))
	success := input.Password == user.Password
	return LoginOutput{Success: success}
}
