package core

type RPCInput struct {
	Function string
	Input    interface{}
}

type User struct {
	Username  string
	Password  string
	Followers []string
}

func Tuser() string {
	return "user"
}

func Tpost() string {
	return "post"
}

func Tpublish() string {
	return "publish"
}

func Ttimeline() string {
	return "timeline"
}
