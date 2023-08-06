package main

import (
	// "fmt"
	"math/rand"
	"os"
	"strconv"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/eniac/Beldi/internal/retwis/core"
	"github.com/eniac/Beldi/pkg/cayonlib"
	// "time"
)

var services = []string{"user", "timeline"}

var nUsers = 10000
var nFollows = 8

func init() {
	if nu, err := strconv.Atoi(os.Getenv("NUM_USERS")); err == nil {
		nUsers = nu
	} else {
		panic("invalid NUM_USERS")
	}
	if nf, err := strconv.Atoi(os.Getenv("NUM_FOLLOWERS")); err == nil {
		nFollows = nf
	} else {
		panic("invalid NUM_FOLLOWERS")
	}
	rand.Seed(1)
}

func tables(baseline bool) {
	if baseline {
		panic("Not implemented for baseline")
	} else {
		for {
			tablenames := []string{}
			for _, service := range services {
				cayonlib.CreateLambdaTables(service)
				tablenames = append(tablenames, service)
			}
			if cayonlib.WaitUntilAllActive(tablenames) {
				break
			}
		}
	}
}

func deleteTables(baseline bool) {
	if baseline {
		panic("Not implemented for baseline")
	} else {
		for _, service := range services {
			cayonlib.DeleteLambdaTables(service)
			// cayonlib.WaitUntilAllDeleted([]string{service})
		}
	}
}

func createUsers(baseline bool) {
	for i := 0; i < nUsers; i++ {
		username := "user" + strconv.Itoa(i)
		password := "pwd" + strconv.Itoa(i)
		followers := []string{}
		for j := 0; j < nFollows; j++ {
			follower := "user" + strconv.Itoa(rand.Intn(nUsers))
			followers = append(followers, follower)
		}
		cayonlib.Populate("user", username, core.User{
			Username:  username,
			Password:  password,
			Followers: followers,
		}, baseline)
	}
}

func createTimelines(baseline bool) {
	for i := 0; i < nUsers; i++ {
		username := "user" + strconv.Itoa(i)
		cayonlib.Populate("timeline", username, aws.JSONValue{
			"Posts": make(map[string]interface{}),
		}, baseline)
	}
}

func populate(baseline bool) {
	createUsers(baseline)
	createTimelines(baseline)
}

func main() {
	option := os.Args[1]
	baseline := os.Args[2] == "baseline"
	if option == "create" {
		tables(baseline)
	} else if option == "populate" {
		populate(baseline)
	} else if option == "clean" {
		deleteTables(baseline)
	}
}
