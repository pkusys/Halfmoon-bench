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
	"github.com/eniac/Beldi/pkg/beldilib"
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

func tables() {
	if beldilib.TYPE != "BASELINE" {
		panic("Not implemented for baseline")
	}
	for {
		tablenames := []string{}
		for _, service := range services {
			// tablename := fmt.Sprintf("b%s", service)
			tablename := service
			beldilib.CreateBaselineTable(tablename)
			time.Sleep(2 * time.Second)
			tablenames = append(tablenames, tablename)
		}
		if beldilib.WaitUntilAllActive(tablenames) {
			break
		}
	}
}

func deleteTables() {
	if beldilib.TYPE != "BASELINE" {
		panic("Not implemented for baseline")
	}
	for _, service := range services {
		// beldilib.DeleteTable(fmt.Sprintf("b%s", service))
		beldilib.DeleteTable(service)
	}
}

func createUsers() {
	for i := 0; i < nUsers; i++ {
		username := "user" + strconv.Itoa(i)
		password := "pwd" + strconv.Itoa(i)
		followers := []string{}
		for j := 0; j < nFollows; j++ {
			follower := "user" + strconv.Itoa(rand.Intn(nUsers))
			followers = append(followers, follower)
		}
		beldilib.PopulateBaseline("user", username, core.User{
			Username:  username,
			Password:  password,
			Followers: followers,
		})
	}
}

func createPosts() {
	content := utils.RandomString(valueSize)
	initialTL := make(map[string]interface{})
	for i := 0; i < maxReturnPosts; i++ {
		timestamp := time.Now().Format("2006-01-02 15:04:05")
		postID := fmt.Sprintf("%v-%v", timestamp, i)
		initialTL[postID] = core.PostInfo{
			Timestamp: timestamp,
		}
		beldilib.PopulateBaseline("post", postID, content)
	}
	for i := 0; i < nUsers; i++ {
		username := "user" + strconv.Itoa(i)
		beldilib.PopulateBaseline("timeline", username, aws.JSONValue{
			"Posts": initialTL,
		})
	}
}

func populate() {
	createUsers()
	createPosts()
}

func main() {
	option := os.Args[1]
	if option == "create" {
		tables()
	} else if option == "populate" {
		populate()
	} else if option == "clean" {
		deleteTables()
	}
}
