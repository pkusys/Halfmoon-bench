package cayonlib

import (
	"log"
	"os"

	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/dynamodb"
)

var sess = session.Must(session.NewSessionWithOptions(session.Options{
	SharedConfigState: session.SharedConfigEnable,
}))

var DBClient = dynamodb.New(sess)

var T = int64(60)

var TYPE = "WRITELOG" // options: READLOG, WRITELOG

func init() {
	switch os.Getenv("LoggingMode") {
	case "read":
		TYPE = "READLOG"
	case "write":
		TYPE = "WRITELOG"
	case "":
		TYPE = "WRITELOG"
		log.Println("[INFO] LoggingMode not set, defaulting to WRITELOG")
	default:
		log.Fatalf("[FATAL] invalid LoggingMode: %s", os.Getenv("LoggingMode"))
	}
	log.Printf("[INFO] log mode: %s", TYPE)
}

func CHECK(err error) {
	if err != nil {
		panic(err)
	}
}

var kTablePrefix = os.Getenv("TABLE_PREFIX")
