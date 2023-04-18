package beldilib

import (
	"fmt"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/aws/aws-sdk-go/service/dynamodb/expression"
	"github.com/aws/aws-sdk-go/service/dynamodbstreams"
)

func CreateMainTable(lambdaId string) {
	if TYPE == "WRITELOG" {
		_, err := DBClient.CreateTable(&dynamodb.CreateTableInput{
			BillingMode: aws.String("PAY_PER_REQUEST"),
			AttributeDefinitions: []*dynamodb.AttributeDefinition{
				{
					AttributeName: aws.String("K"),
					AttributeType: aws.String("S"),
				},
				{
					AttributeName: aws.String("KEY"),
					AttributeType: aws.String("S"),
				},
				{
					AttributeName: aws.String("VERSION"),
					AttributeType: aws.String("N"),
				},
			},
			KeySchema: []*dynamodb.KeySchemaElement{
				{
					AttributeName: aws.String("K"),
					KeyType:       aws.String("HASH"),
				},
			},
			GlobalSecondaryIndexes: []*dynamodb.GlobalSecondaryIndex{
				&dynamodb.GlobalSecondaryIndex{
					IndexName: aws.String("keyidx"),
					KeySchema: []*dynamodb.KeySchemaElement{
						{
							AttributeName: aws.String("KEY"),
							KeyType:       aws.String("HASH"),
						},
						{
							AttributeName: aws.String("VERSION"),
							KeyType:       aws.String("RANGE"),
						},
					},
					Projection: &dynamodb.Projection{
						NonKeyAttributes: []*string{aws.String("V")},
						ProjectionType:   aws.String("INCLUDE"),
					},
				},
			},
			TableName: aws.String(kTablePrefix + lambdaId),
		})
		CHECK(err)
	} else {
		_, err := DBClient.CreateTable(&dynamodb.CreateTableInput{
			BillingMode: aws.String("PAY_PER_REQUEST"),
			AttributeDefinitions: []*dynamodb.AttributeDefinition{
				{
					AttributeName: aws.String("K"),
					AttributeType: aws.String("S"),
				},
			},
			KeySchema: []*dynamodb.KeySchemaElement{
				{
					AttributeName: aws.String("K"),
					KeyType:       aws.String("HASH"),
				},
			},
			TableName: aws.String(kTablePrefix + lambdaId),
		})
		CHECK(err)
	}
}

func CreateEOSTable(lambdaId string) {
	_, err := DBClient.CreateTable(&dynamodb.CreateTableInput{
		BillingMode: aws.String("PAY_PER_REQUEST"),
		AttributeDefinitions: []*dynamodb.AttributeDefinition{
			{
				AttributeName: aws.String("K"),
				AttributeType: aws.String("S"),
			},
			{
				AttributeName: aws.String("ROWHASH"),
				AttributeType: aws.String("S"),
			},
			{
				AttributeName: aws.String("TS"),
				AttributeType: aws.String("N"),
			},
		},
		KeySchema: []*dynamodb.KeySchemaElement{
			{
				AttributeName: aws.String("K"),
				KeyType:       aws.String("HASH"),
			},
			{
				AttributeName: aws.String("ROWHASH"),
				KeyType:       aws.String("RANGE"),
			},
		},
		GlobalSecondaryIndexes: []*dynamodb.GlobalSecondaryIndex{
			&dynamodb.GlobalSecondaryIndex{
				IndexName: aws.String("rowidx"),
				KeySchema: []*dynamodb.KeySchemaElement{
					{
						AttributeName: aws.String("ROWHASH"),
						KeyType:       aws.String("HASH"),
					},
				},
				Projection: &dynamodb.Projection{
					NonKeyAttributes: []*string{aws.String("K"), aws.String("NEXTROW")},
					ProjectionType:   aws.String("INCLUDE"),
				},
			},
			&dynamodb.GlobalSecondaryIndex{
				IndexName: aws.String("tsidx"),
				KeySchema: []*dynamodb.KeySchemaElement{
					{
						AttributeName: aws.String("TS"),
						KeyType:       aws.String("HASH"),
					},
				},
				Projection: &dynamodb.Projection{
					NonKeyAttributes: []*string{aws.String("K"), aws.String("NEXTROW")},
					ProjectionType:   aws.String("INCLUDE"),
				},
			},
		},
		TableName: aws.String(kTablePrefix + lambdaId),
	})
	CHECK(err)
}

func CreateLogTable(lambdaId string) {
	_, err := DBClient.CreateTable(&dynamodb.CreateTableInput{
		BillingMode: aws.String("PAY_PER_REQUEST"),
		AttributeDefinitions: []*dynamodb.AttributeDefinition{
			{
				AttributeName: aws.String("InstanceStep"),
				AttributeType: aws.String("S"),
			},
			// {
			// 	AttributeName: aws.String("StepNumber"),
			// 	AttributeType: aws.String("N"),
			// },
		},
		KeySchema: []*dynamodb.KeySchemaElement{
			{
				AttributeName: aws.String("InstanceStep"),
				KeyType:       aws.String("HASH"),
			},
			// {
			// 	AttributeName: aws.String("StepNumber"),
			// 	KeyType:       aws.String("RANGE"),
			// },
		},
		TableName: aws.String(kTablePrefix + fmt.Sprintf("%s-log", lambdaId)),
	})
	CHECK(err)
}

func CreateCollectorTable(lambdaId string) {
	_, err := DBClient.CreateTable(&dynamodb.CreateTableInput{
		BillingMode: aws.String("PAY_PER_REQUEST"),
		AttributeDefinitions: []*dynamodb.AttributeDefinition{
			{
				AttributeName: aws.String("InstanceId"),
				AttributeType: aws.String("S"),
			},
			{
				AttributeName: aws.String("TS"),
				AttributeType: aws.String("N"),
			},
		},
		KeySchema: []*dynamodb.KeySchemaElement{
			{
				AttributeName: aws.String("InstanceId"),
				KeyType:       aws.String("HASH"),
			},
		},
		TableName: aws.String(kTablePrefix + fmt.Sprintf("%s-collector", lambdaId)),
		GlobalSecondaryIndexes: []*dynamodb.GlobalSecondaryIndex{
			&dynamodb.GlobalSecondaryIndex{
				IndexName: aws.String("tsidx"),
				KeySchema: []*dynamodb.KeySchemaElement{
					{
						AttributeName: aws.String("TS"),
						KeyType:       aws.String("HASH"),
					},
				},
				Projection: &dynamodb.Projection{
					ProjectionType: aws.String("KEYS_ONLY"),
				},
			},
		},
	})
	CHECK(err)
}

func CreateCounterTable() {
	_, err := DBClient.CreateTable(&dynamodb.CreateTableInput{
		BillingMode: aws.String("PAY_PER_REQUEST"),
		AttributeDefinitions: []*dynamodb.AttributeDefinition{
			{
				AttributeName: aws.String("K"),
				AttributeType: aws.String("S"),
			},
		},
		KeySchema: []*dynamodb.KeySchemaElement{
			{
				AttributeName: aws.String("K"),
				KeyType:       aws.String("HASH"),
			},
		},
		TableName: aws.String(kTablePrefix + "counter"),
	})
	CHECK(err)
}

func GetStreamArn(lambdaId string) string {
	table, err := DBClient.DescribeTable(&dynamodb.DescribeTableInput{
		TableName: aws.String(kTablePrefix + lambdaId),
	})
	CHECK(err)
	return aws.StringValue(table.Table.LatestStreamArn)
}

func EnableStream(lambdaId string) bool {
	fmt.Printf("Enabling stream for %s\n", lambdaId)
	_, err := DBClient.UpdateTable(&dynamodb.UpdateTableInput{
		TableName: aws.String(kTablePrefix + lambdaId),
		StreamSpecification: &dynamodb.StreamSpecification{
			StreamEnabled:  aws.Bool(true),
			StreamViewType: aws.String("KEYS_ONLY"),
		},
	})
	CHECK(err)
	WaitUntilActive(lambdaId)
	arn := GetStreamArn(lambdaId)
	fmt.Printf("Stream arn: %s\n", arn)
	for {
		res, err := DBStreamClient.DescribeStream(&dynamodbstreams.DescribeStreamInput{
			StreamArn: aws.String(arn),
		})
		if err != nil {
			fmt.Printf("%s DescribeStream error: %v\n", lambdaId, err)
		} else {
			fmt.Printf("%s status: %s\n", lambdaId, *res.StreamDescription.StreamStatus)
			if *res.StreamDescription.StreamStatus == "ENABLED" {
				return true
			}
		}
		time.Sleep(3 * time.Second)
	}
}

func CreateBaselineTable(lambdaId string) {
	_, _ = DBClient.CreateTable(&dynamodb.CreateTableInput{
		BillingMode: aws.String("PAY_PER_REQUEST"),
		AttributeDefinitions: []*dynamodb.AttributeDefinition{
			{
				AttributeName: aws.String("K"),
				AttributeType: aws.String("S"),
			},
		},
		KeySchema: []*dynamodb.KeySchemaElement{
			{
				AttributeName: aws.String("K"),
				KeyType:       aws.String("HASH"),
			},
		},
		TableName: aws.String(kTablePrefix + lambdaId),
	})
}

func CreateLambdaTables(lambdaId string) {
	CreateMainTable(lambdaId)
	CreateLogTable(lambdaId)
	CreateCollectorTable(lambdaId)
}

func CreateTxnTables(lambdaId string) {
	CreateEOSTable(lambdaId)
	CreateLogTable(lambdaId)
	CreateCollectorTable(lambdaId)
}

func DeleteTable(tablename string) {
	_, _ = DBClient.DeleteTable(&dynamodb.DeleteTableInput{TableName: aws.String(kTablePrefix + tablename)})
}

func DeleteLambdaTables(lambdaId string) {
	DeleteTable(lambdaId)
	DeleteTable(fmt.Sprintf("%s-log", lambdaId))
	DeleteTable(fmt.Sprintf("%s-collector", lambdaId))
}

func WaitUntilDeleted(tablename string) {
	for {
		res, err := DBClient.DescribeTable(&dynamodb.DescribeTableInput{TableName: aws.String(kTablePrefix + tablename)})
		if err != nil {
			if aerr, ok := err.(awserr.Error); ok {
				switch aerr.Code() {
				case dynamodb.ErrCodeResourceNotFoundException:
					return
				}
			}
		} else if *res.Table.TableStatus != "DELETING" {
			DeleteTable(tablename)
		}
		time.Sleep(3 * time.Second)
	}
}

func WaitUntilAllDeleted(tablenames []string) {
	for _, tablename := range tablenames {
		WaitUntilDeleted(tablename)
	}
}

func WaitUntilActive(tablename string) bool {
	counter := 0
	for {
		res, err := DBClient.DescribeTable(&dynamodb.DescribeTableInput{TableName: aws.String(kTablePrefix + tablename)})
		if err != nil {
			counter += 1
			fmt.Printf("%s DescribeTable error: %v\n", tablename, err)
		} else {
			if *res.Table.TableStatus == "ACTIVE" {
				fmt.Printf("%s status: %s\n", tablename, *res.Table.TableStatus)
				return true
			}
			fmt.Printf("%s status: %s\n", tablename, *res.Table.TableStatus)
			// if *res.Table.TableStatus != "CREATING" && counter > 6 {
			// 	fmt.Printf("[error] %s status: %s\n", tablename, *res.Table.TableStatus)
			// 	return false
			// }
		}
		time.Sleep(3 * time.Second)
	}
}

func WaitUntilAllActive(tablenames []string) bool {
	for _, tablename := range tablenames {
		res := WaitUntilActive(tablename)
		if !res {
			return false
		}
	}
	return true
}

func Populate(tablename string, key string, value interface{}, baseline bool) {
	if TYPE == "WRITELOG" {
		LibWrite(tablename, aws.JSONValue{"K": key}, map[expression.NameBuilder]expression.OperandBuilder{
			expression.Name("KEY"):     expression.Value(key),
			expression.Name("VERSION"): expression.Value(1),
			expression.Name("V"):       expression.Value(value),
		})
	} else {
		LibWrite(tablename, aws.JSONValue{"K": key}, map[expression.NameBuilder]expression.OperandBuilder{
			expression.Name("VERSION"): expression.Value(1),
			expression.Name("V"): expression.Value(value),
		})
	}
}

func PopulateEOS(tablename string, key string, value interface{}, baseline bool) {
	LibWrite(tablename, aws.JSONValue{"K": key, "ROWHASH": "HEAD"},
		map[expression.NameBuilder]expression.OperandBuilder{
			expression.Name("GCSIZE"):  expression.Value(0),
			expression.Name("LOGSIZE"): expression.Value(0),
			expression.Name("LOGS"):    expression.Value(aws.JSONValue{"ignore": nil}),
			expression.Name("V"):       expression.Value(value),
		})
}
