package cayonlib

import (
	"encoding/json"
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/aws/aws-sdk-go/service/dynamodb/dynamodbattribute"
	"github.com/aws/aws-sdk-go/service/dynamodb/expression"
	"github.com/golang/snappy"
)

func LibReadMultiVersion(tablename string, key string, version uint64) aws.JSONValue {
	Key, err := dynamodbattribute.MarshalMap(aws.JSONValue{"K": key + fmt.Sprintf("_%v", version)})
	CHECK(err)
	projectionBuilder := BuildProjection([]string{"V"})
	expr, err := expression.NewBuilder().WithProjection(projectionBuilder).Build()
	CHECK(err)
	res, err := DBClient.GetItem(&dynamodb.GetItemInput{
		TableName:                aws.String(kTablePrefix + tablename),
		Key:                      Key,
		ProjectionExpression:     expr.Projection(),
		ExpressionAttributeNames: expr.Names(),
		// ConsistentRead:           aws.Bool(true),
	})
	// this version may not exist
	item := aws.JSONValue{}
	if err != nil {
		AssertResourceNotFound(err)
		log.Printf("[WARN] table %s key %s version %v not found", tablename, key, version)
		return item
	}
	err = dynamodbattribute.UnmarshalMap(res.Item, &item)
	CHECK(err)
	return item
}

func LibReadSingleVersion(tablename string, key string) aws.JSONValue {
	return LibReadMultiVersion(tablename, key, 0)
}

func LibScanWithLast(tablename string, projection []string, last map[string]*dynamodb.AttributeValue) []aws.JSONValue {
	var res *dynamodb.ScanOutput
	var scanErr error
	if last == nil {
		if len(projection) == 0 {
			expr, err := expression.NewBuilder().Build()
			CHECK(err)
			res, scanErr = DBClient.Scan(&dynamodb.ScanInput{
				TableName:                 aws.String(kTablePrefix + tablename),
				ExpressionAttributeNames:  expr.Names(),
				ExpressionAttributeValues: expr.Values(),
				FilterExpression:          expr.Filter(),
				// ConsistentRead:            aws.Bool(true),
			})
		} else {
			expr, err := expression.NewBuilder().WithProjection(BuildProjection(projection)).Build()
			CHECK(err)
			res, scanErr = DBClient.Scan(&dynamodb.ScanInput{
				TableName:                 aws.String(kTablePrefix + tablename),
				ExpressionAttributeNames:  expr.Names(),
				ExpressionAttributeValues: expr.Values(),
				FilterExpression:          expr.Filter(),
				ProjectionExpression:      expr.Projection(),
				// ConsistentRead:            aws.Bool(true),
			})
		}
	} else {
		if len(projection) == 0 {
			expr, err := expression.NewBuilder().Build()
			CHECK(err)
			res, scanErr = DBClient.Scan(&dynamodb.ScanInput{
				TableName:                 aws.String(kTablePrefix + tablename),
				ExpressionAttributeNames:  expr.Names(),
				ExpressionAttributeValues: expr.Values(),
				FilterExpression:          expr.Filter(),
				// ConsistentRead:            aws.Bool(true),
				ExclusiveStartKey: last,
			})
		} else {
			expr, err := expression.NewBuilder().WithProjection(BuildProjection(projection)).Build()
			CHECK(err)
			res, scanErr = DBClient.Scan(&dynamodb.ScanInput{
				TableName:                 aws.String(kTablePrefix + tablename),
				ExpressionAttributeNames:  expr.Names(),
				ExpressionAttributeValues: expr.Values(),
				FilterExpression:          expr.Filter(),
				ProjectionExpression:      expr.Projection(),
				// ConsistentRead:            aws.Bool(true),
				ExclusiveStartKey: last,
			})
		}
	}
	CHECK(scanErr)
	var item []aws.JSONValue
	err := dynamodbattribute.UnmarshalListOfMaps(res.Items, &item)
	CHECK(err)
	if res.LastEvaluatedKey == nil || len(res.LastEvaluatedKey) == 0 {
		return item
	}
	log.Printf("[DEBUG] Exceed Scan limit")
	item = append(item, LibScanWithLast(tablename, projection, res.LastEvaluatedKey)...)
	return item
}

func LibScan(tablename string, projection []string) []aws.JSONValue {
	return LibScanWithLast(tablename, projection, nil)
}

func LibWrite(tablename string, key string,
	update map[expression.NameBuilder]expression.OperandBuilder) {
	Key, err := dynamodbattribute.MarshalMap(aws.JSONValue{"K": key + "_0"})
	CHECK(err)
	if len(update) == 0 {
		panic("update never be empty")
	}
	updateBuilder := expression.UpdateBuilder{}
	for k, v := range update {
		updateBuilder = updateBuilder.Set(k, v)
	}
	builder := expression.NewBuilder().WithUpdate(updateBuilder)
	expr, err := builder.Build()
	CHECK(err)
	_, err = DBClient.UpdateItem(&dynamodb.UpdateItemInput{
		TableName:                 aws.String(kTablePrefix + tablename),
		Key:                       Key,
		ExpressionAttributeNames:  expr.Names(),
		ExpressionAttributeValues: expr.Values(),
		UpdateExpression:          expr.Update(),
	})
	CHECK(err)
}

func LibWriteMultiVersion(tablename string, key string, version uint64,
	update map[expression.NameBuilder]expression.OperandBuilder) {
	Key, err := dynamodbattribute.MarshalMap(aws.JSONValue{"K": key + fmt.Sprintf("_%v", version)})
	CHECK(err)
	if len(update) == 0 {
		panic("update never be empty")
	}
	updateBuilder := expression.UpdateBuilder{}
	for k, v := range update {
		updateBuilder = updateBuilder.Set(k, v)
	}
	builder := expression.NewBuilder().WithUpdate(updateBuilder)
	expr, err := builder.Build()
	CHECK(err)
	_, err = DBClient.UpdateItem(&dynamodb.UpdateItemInput{
		TableName:                 aws.String(kTablePrefix + tablename),
		Key:                       Key,
		ExpressionAttributeNames:  expr.Names(),
		ExpressionAttributeValues: expr.Values(),
		UpdateExpression:          expr.Update(),
	})
	CHECK(err)
}

func LibWriteSingleVersion(tablename string, key string, version uint64,
	update map[expression.NameBuilder]expression.OperandBuilder) {
	Key, err := dynamodbattribute.MarshalMap(aws.JSONValue{"K": key + "_0"})
	CHECK(err)
	condBuilder := expression.Or(
		expression.AttributeNotExists(expression.Name("VERSION")),
		expression.Name("VERSION").LessThan(expression.Value(version)))
	updateBuilder := expression.UpdateBuilder{}
	for k, v := range update {
		updateBuilder = updateBuilder.Set(k, v)
	}
	updateBuilder = updateBuilder.
		Set(expression.Name("VERSION"), expression.Value(version))
	expr, err := expression.NewBuilder().WithCondition(condBuilder).WithUpdate(updateBuilder).Build()
	CHECK(err)

	_, err = DBClient.UpdateItem(&dynamodb.UpdateItemInput{
		TableName:                 aws.String(kTablePrefix + tablename),
		Key:                       Key,
		ConditionExpression:       expr.Condition(),
		ExpressionAttributeNames:  expr.Names(),
		ExpressionAttributeValues: expr.Values(),
		UpdateExpression:          expr.Update(),
	})
	if err != nil {
		AssertConditionFailure(err)
		// log.Printf("[ERROR] Write to table %v key %v failed: %v", tablename, key, err.Error())
	}
}

func Write(env *Env, tablename string, key string, update map[expression.NameBuilder]expression.OperandBuilder, dependent bool) {
	if TYPE == "WRITELOG" {
		WriteWithLog(env, tablename, key, update, dependent)
	} else {
		WriteWithoutLog(env, tablename, key, update, dependent)
	}
}

func WriteWithLog(env *Env, tablename string, key string, update map[expression.NameBuilder]expression.OperandBuilder, dependent bool) {
	newLog, preWriteLog := ProposeNextStep(
		env,
		// []uint64{DatabaseKeyTag(tablename, key)},
		nil,
		aws.JSONValue{
			"type":  "PreWrite",
			"key":   key,
			"table": tablename,
		})
	if preWriteLog == nil {
		return
	}
	if preWriteLog.SeqNum != env.SeqNum {
		log.Fatalf("[ERROR] PreWrite log seqnum %x not matching env seqnum %x", preWriteLog.SeqNum, env.SeqNum)
	}
	if !newLog {
		CheckLogDataField(preWriteLog, "type", "PreWrite")
		CheckLogDataField(preWriteLog, "table", tablename)
		CheckLogDataField(preWriteLog, "key", key)
		log.Printf("[INFO] Seen PreWrite log for step %d", preWriteLog.StepNumber)
	}
	postWriteLog := env.Fsm.GetStepLog(env.StepNumber)
	if postWriteLog != nil {
		CheckLogDataField(postWriteLog, "type", "PostWrite")
		CheckLogDataField(postWriteLog, "table", tablename)
		CheckLogDataField(postWriteLog, "key", key)
		log.Printf("[INFO] Seen PostWrite log for step %d", postWriteLog.StepNumber)
		env.StepNumber += 1
		env.SeqNum = postWriteLog.SeqNum
		return
	}
	var version uint64
	if !dependent {
		version = env.SeqNum
	}
	LibWriteMultiVersion(tablename, key, version, update)
	ProposeNextStep(
		env,
		[]uint64{DatabaseKeyTag(tablename, key)},
		aws.JSONValue{
			"type":    "PostWrite",
			"key":     key,
			"table":   tablename,
			"version": version,
		})
}

func WriteWithoutLog(env *Env, tablename string, key string, update map[expression.NameBuilder]expression.OperandBuilder, dependent bool) {
	LibWriteSingleVersion(tablename, key, env.SeqNum, update)
}

func Read(env *Env, tablename string, key string) interface{} {
	if TYPE == "WRITELOG" {
		return ReadWithoutLog(env, tablename, key)
	} else {
		return ReadWithLog(env, tablename, key)
	}
}

func checkPostReadLog(intentLog *IntentLogEntry, tablename string, key string) bool {
	return intentLog.Data["type"] == "PostRead" &&
		intentLog.Data["table"] == tablename &&
		intentLog.Data["key"] == key
}

func getVersionFromLog(env *Env, tablename string, key string) (uint64, error) {
	logEntry, err := env.FaasEnv.SharedLogReadPrev(env.FaasCtx, DatabaseKeyTag(tablename, key), env.SeqNum)
	CHECK(err)
	if logEntry == nil {
		return 0, fmt.Errorf("empty version log")
	}
	decoded, err := snappy.Decode(nil, logEntry.Data)
	CHECK(err)
	var intentLog IntentLogEntry
	err = json.Unmarshal(decoded, &intentLog)
	CHECK(err)
	if intentLog.Data["type"] == "PostWrite" || intentLog.Data["type"] == "PostRead" {
		return uint64(intentLog.Data["version"].(float64)), nil
	}
	panic(fmt.Sprintf("Unknown version log type %s", intentLog.Data["type"]))
}

func ReadWithoutLog(env *Env, tablename string, key string) interface{} {
	// if version log is empty then version is 0 and err is set
	version, err := getVersionFromLog(env, tablename, key)
	if err != nil {
		log.Printf("[INFO] Read version not found table=%s key=%s, using the default version(0)", tablename, key)
	}
	item := LibReadMultiVersion(tablename, key, version)
	var result interface{}
	if value, ok := item["V"]; ok {
		result = value
	} else {
		log.Printf("[INFO] Read without log table %s key %s return nil", tablename, key)
		result = nil
	}
	return result
}

// func ReadWithoutLog(env *Env, tablename string, key string) interface{} {
// 	intentLog := env.Fsm.GetStepLog(env.StepNumber)
// 	if intentLog != nil && checkPostReadLog(intentLog, tablename, key) {
// 		log.Printf("[INFO] Seen Read log instance %s step %d", env.InstanceId, env.StepNumber)
// 		env.StepNumber += 1
// 		env.SeqNum = intentLog.SeqNum
// 		return intentLog.Data["result"]
// 	}
// 	// if version log is empty then version is 0 and err is set
// 	version, err := getVersionFromLog(env, tablename, key)
// 	if err != nil {
// 		log.Printf("[INFO] Read version not found instance %s step %d: table=%s key=%s", env.InstanceId, env.StepNumber, tablename, key)
// 	}
// 	item := LibReadMultiVersion(tablename, key, version)
// 	var result interface{}
// 	if value, ok := item["V"]; ok {
// 		result = value
// 	} else {
// 		result = nil
// 	}
// 	// the returned version is the requested one, can go log-free
// 	if err == nil && result != nil {
// 		return result
// 	}
// 	// the returned version is not the requested one, must append readlog for determinism
// 	// now if the intentlog is not nil and is not a readlog, then this is a conflict
// 	if intentLog != nil {
// 		log.Fatalf("[FATAL] Inconsistent read instance %s step %d: table=%s key=%s", env.InstanceId, env.StepNumber, tablename, key)
// 	}
// 	streamTags := []uint64{}
// 	if err != nil && result != nil {
// 		streamTags = append(streamTags, DatabaseKeyTag(tablename, key))
// 	}
// 	newLog, _ := ProposeNextStep(env, streamTags, aws.JSONValue{
// 		"type":    "PostRead",
// 		"key":     key,
// 		"table":   tablename,
// 		"result":  result,
// 		"version": version,
// 	})
// 	// a concurrent step log, may or may not be a post read log
// 	if !newLog {
// 		return nil
// 	}
// 	return result
// }

func ReadWithLog(env *Env, tablename string, key string) interface{} {
	postReadLog := env.Fsm.GetStepLog(env.StepNumber)
	if postReadLog != nil {
		if !checkPostReadLog(postReadLog, tablename, key) {
			log.Fatalf("[FATAL] Missing read log instance %s step %d: table=%s key=%s", env.InstanceId, env.StepNumber, tablename, key)
		}
		log.Printf("[INFO] Seen read log instance %s step %d", env.InstanceId, env.StepNumber)
		env.StepNumber += 1
		env.SeqNum = postReadLog.SeqNum
		return postReadLog.Data["result"]
	}
	item := LibReadSingleVersion(tablename, key)
	var result interface{}
	if value, ok := item["V"]; ok {
		result = value
	} else {
		log.Printf("[INFO] Read with log table %s key %s return nil", tablename, key)
		result = nil
	}
	newLog, _ := ProposeNextStep(env, nil, aws.JSONValue{
		"type":   "PostRead",
		"key":    key,
		"table":  tablename,
		"result": result,
	})
	// a concurrent step log, must be a post read log but not necessarily the same version
	// possible to catch up with the other version, but choose to exit the function for simplicity
	if !newLog {
		return nil
	}
	return result
}

func checkPostScanLog(intentLog *IntentLogEntry, tablename string) bool {
	return intentLog.Data["type"] == "PostScan" &&
		intentLog.Data["table"] == tablename
}

func Scan(env *Env, tablename string) interface{} {
	postScanLog := env.Fsm.GetStepLog(env.StepNumber)
	if postScanLog != nil {
		if !checkPostScanLog(postScanLog, tablename) {
			log.Fatalf("[FATAL] Missing scan log instance %s step %d: table=%s", env.InstanceId, env.StepNumber, tablename)
		}
		log.Printf("[INFO] Seen scan log instance %s step %d", env.InstanceId, env.StepNumber)
		env.StepNumber += 1
		env.SeqNum = postScanLog.SeqNum
		return postScanLog.Data["result"]
	}
	// log.Printf("[INFO] Scanning table %v: instance %s step %d", tablename, env.InstanceId, env.StepNumber)
	items := LibScan(tablename, []string{"V"})
	var result []interface{}
	for _, item := range items {
		result = append(result, item["V"])
	}
	// log.Printf("[INFO] Finished scanning table %v(%v items): instance %s step %d", tablename, len(result), env.InstanceId, env.StepNumber)
	newLog, _ := ProposeNextStep(env, nil, aws.JSONValue{
		"type":   "PostScan",
		"table":  tablename,
		"result": result,
	})
	if !newLog {
		return nil
	}
	return result
}

func BuildProjection(names []string) expression.ProjectionBuilder {
	if len(names) == 0 {
		panic("Projection must > 0")
	}
	var builder expression.ProjectionBuilder
	for _, name := range names {
		builder = builder.AddNames(expression.Name(name))
	}
	return builder
}

func AssertConditionFailure(err error) {
	if aerr, ok := err.(awserr.Error); ok {
		switch aerr.Code() {
		case dynamodb.ErrCodeConditionalCheckFailedException:
			return
		case dynamodb.ErrCodeResourceNotFoundException:
			log.Printf("ERROR: DyanombDB ResourceNotFound")
			return
		default:
			log.Printf("ERROR: %s", aerr)
			panic("ERROR detected")
		}
	} else {
		log.Printf("ERROR: %s", err)
		panic("ERROR detected")
	}
}

func AssertResourceNotFound(err error) {
	if aerr, ok := err.(awserr.Error); ok {
		switch aerr.Code() {
		case dynamodb.ErrCodeResourceNotFoundException:
			return
		default:
			log.Printf("ERROR: %s", aerr)
			panic("ERROR detected")
		}
	} else {
		log.Printf("ERROR: %s", err)
		panic("ERROR detected")
	}
}
