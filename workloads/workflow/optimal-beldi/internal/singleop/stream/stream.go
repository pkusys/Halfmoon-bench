package main

import (
	"fmt"
	"log"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/dynamodbstreams"
	"github.com/eniac/Beldi/pkg/beldilib"
)

const table = "singleop"

var interval = time.Millisecond * 200

func main() {
	if beldilib.TYPE != "WRITELOG" {
		return
	}
	// get counter ts
	counterItem := beldilib.LibRead("counter", aws.JSONValue{"K": "counter"}, nil)
	counter := int64(counterItem["V"].(float64))
	// get stream info
	streamClient := beldilib.DBStreamClient
	arn := beldilib.GetStreamArn(fmt.Sprintf("%s", table))
	stream, err := streamClient.DescribeStream(&dynamodbstreams.DescribeStreamInput{
		StreamArn: aws.String(arn),
		Limit:     aws.Int64(1),
	})
	beldilib.CHECK(err)
	shardId := *stream.StreamDescription.Shards[0].ShardId
	iter, err := streamClient.GetShardIterator(&dynamodbstreams.GetShardIteratorInput{
		ShardId:           aws.String(shardId),
		ShardIteratorType: aws.String("LATEST"),
		StreamArn:         aws.String(arn),
	})
	beldilib.CHECK(err)
	shardIterator := aws.StringValue(iter.ShardIterator)
	// Process records from the stream
	for {
		records, err := streamClient.GetRecords(&dynamodbstreams.GetRecordsInput{
			ShardIterator: aws.String(shardIterator),
		})
		if err != nil {
			beldilib.AssertResourceNotFound(err)
			log.Printf("Stream closed")
			break
		}
		for _, record := range records.Records {
			counter += 1
			beldilib.AssignEventTS(fmt.Sprintf("%s", table), record, counter)
		}
		beldilib.AdvanceCounterTS(counter)
		// Set the shard iterator for the next batch of records
		shardIterator = aws.StringValue(records.NextShardIterator)
		if shardIterator == "" {
			log.Printf("Stream closed")
			break
		}
		// Wait for a short period of time before getting the next batch of records
		time.Sleep(interval)
	}
}
