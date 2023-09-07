package main

import (
	// "fmt"
	"fmt"
	"math/rand"
	"os"
	"strconv"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/eniac/Beldi/internal/retwis/core"
	"github.com/eniac/Beldi/internal/utils"
	"github.com/eniac/Beldi/pkg/cayonlib"
	// "time"
)

var services = []string{"user", "post", "timeline"}

var nUsers = 10000
var nFollows = 8
var maxReturnPosts = 4
var valueSize = 256

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
	if mr, err := strconv.Atoi(os.Getenv("MAX_RETURN_POSTS")); err == nil {
		maxReturnPosts = mr
	} else {
		panic("invalid MAX_RETURN_POSTS")
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

func createPosts(baseline bool) {
	content := utils.RandomString(valueSize)
	initialTL := make(map[string]interface{})
	for i := 0; i < maxReturnPosts; i++ {
		timestamp := time.Now().Format("2006-01-02 15:04:05")
		postID := fmt.Sprintf("%v-%v", timestamp, i)
		initialTL[postID] = core.PostInfo{
			Timestamp: timestamp,
		}
		cayonlib.Populate("post", postID, content, baseline)
	}
	for i := 0; i < nUsers; i++ {
		username := "user" + strconv.Itoa(i)
		cayonlib.Populate("timeline", username, aws.JSONValue{
			"Posts": initialTL,
		}, baseline)
	}
}

func populate(baseline bool) {
	createUsers(baseline)
	createPosts(baseline)
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
