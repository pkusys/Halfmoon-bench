package cayonlib

import (
	"log"

	// "fmt"
	"cs.utexas.edu/zjia/faas/protocol"
	"github.com/aws/aws-sdk-go/aws/awserr"

	// "github.com/mitchellh/mapstructure"
	// "strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/aws/aws-sdk-go/service/dynamodb/dynamodbattribute"
	"github.com/aws/aws-sdk-go/service/dynamodb/expression"
	// "github.com/lithammer/shortuuid"
)

func LibReadWithoutLog(tablename string, key string, version uint64) aws.JSONValue {
	keyCondBuilder := expression.Key("K").Equal(expression.Value(key)).And(
		expression.Key("VERSION").LessThanEqual(expression.Value(version)))
	projectionBuilder := BuildProjection([]string{"VERSION", "V"})
	expr, err := expression.NewBuilder().WithKeyCondition(keyCondBuilder).WithProjection(projectionBuilder).Build()
	CHECK(err)
	res, err := DBClient.Query(&dynamodb.QueryInput{
		TableName:                 aws.String(kTablePrefix + tablename),
		KeyConditionExpression:    expr.KeyCondition(),
		ProjectionExpression:      expr.Projection(),
		ExpressionAttributeNames:  expr.Names(),
		ExpressionAttributeValues: expr.Values(),
		ScanIndexForward:          aws.Bool(false), // to retrieve latest item first
		Limit:                     aws.Int64(1),    // to retrieve only the latest item
	})
	CHECK(err)
	if len(res.Items) == 0 {
		return nil
	}
	item := aws.JSONValue{}
	err = dynamodbattribute.UnmarshalMap(res.Items[0], &item)
	CHECK(err)
	return item
}

func LibReadWithLog(tablename string, key string) aws.JSONValue {
	Key, err := dynamodbattribute.MarshalMap(aws.JSONValue{"K": key})
	CHECK(err)
	projectionBuilder := BuildProjection([]string{"V"})
	expr, err := expression.NewBuilder().WithProjection(projectionBuilder).Build()
	CHECK(err)
	res, err := DBClient.GetItem(&dynamodb.GetItemInput{
		TableName:                aws.String(kTablePrefix + tablename),
		Key:                      Key,
		ProjectionExpression:     expr.Projection(),
		ExpressionAttributeNames: expr.Names(),
		ConsistentRead:           aws.Bool(true),
	})
	CHECK(err)
	item := aws.JSONValue{}
	err = dynamodbattribute.UnmarshalMap(res.Item, &item)
	CHECK(err)
	return item
}

func LibWrite(tablename string, key aws.JSONValue,
	update map[expression.NameBuilder]expression.OperandBuilder) {
	Key, err := dynamodbattribute.MarshalMap(key)
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

func LibScanWithLast(tablename string, projection []string, last map[string]*dynamodb.AttributeValue) []aws.JSONValue {
	var res *dynamodb.ScanOutput
	var err error
	if last == nil {
		if len(projection) == 0 {
			expr, err := expression.NewBuilder().Build()
			CHECK(err)
			res, err = DBClient.Scan(&dynamodb.ScanInput{
				TableName:                 aws.String(tablename),
				ExpressionAttributeNames:  expr.Names(),
				ExpressionAttributeValues: expr.Values(),
				FilterExpression:          expr.Filter(),
				ConsistentRead:            aws.Bool(true),
			})
		} else {
			expr, err := expression.NewBuilder().WithProjection(BuildProjection(projection)).Build()
			CHECK(err)
			res, err = DBClient.Scan(&dynamodb.ScanInput{
				TableName:                 aws.String(tablename),
				ExpressionAttributeNames:  expr.Names(),
				ExpressionAttributeValues: expr.Values(),
				FilterExpression:          expr.Filter(),
				ProjectionExpression:      expr.Projection(),
				ConsistentRead:            aws.Bool(true),
			})
		}
	} else {
		if len(projection) == 0 {
			expr, err := expression.NewBuilder().Build()
			CHECK(err)
			res, err = DBClient.Scan(&dynamodb.ScanInput{
				TableName:                 aws.String(tablename),
				ExpressionAttributeNames:  expr.Names(),
				ExpressionAttributeValues: expr.Values(),
				FilterExpression:          expr.Filter(),
				ConsistentRead:            aws.Bool(true),
				ExclusiveStartKey:         last,
			})
		} else {
			expr, err := expression.NewBuilder().WithProjection(BuildProjection(projection)).Build()
			CHECK(err)
			res, err = DBClient.Scan(&dynamodb.ScanInput{
				TableName:                 aws.String(tablename),
				ExpressionAttributeNames:  expr.Names(),
				ExpressionAttributeValues: expr.Values(),
				FilterExpression:          expr.Filter(),
				ProjectionExpression:      expr.Projection(),
				ConsistentRead:            aws.Bool(true),
				ExclusiveStartKey:         last,
			})
		}
	}
	CHECK(err)
	var item []aws.JSONValue
	err = dynamodbattribute.UnmarshalListOfMaps(res.Items, &item)
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

func Write(env *Env, tablename string, key string, update map[expression.NameBuilder]expression.OperandBuilder) {
	if TYPE == "WRITELOG" {
		WriteWithLog(env, tablename, key, update)
	} else {
		WriteWithoutLog(env, tablename, key, update)
	}
}

func WriteWithLog(env *Env, tablename string, key string, update map[expression.NameBuilder]expression.OperandBuilder) {
	newLog, preWriteLog := ProposeNextStep(
		env,
		[]uint64{DatabaseKeyTag(tablename, key)},
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
	Key, err := dynamodbattribute.MarshalMap(aws.JSONValue{"K": key, "VERSION": env.SeqNum})
	CHECK(err)
	updateBuilder := expression.UpdateBuilder{}
	for k, v := range update {
		updateBuilder = updateBuilder.Set(k, v)
	}
	expr, err := expression.NewBuilder().WithUpdate(updateBuilder).Build()
	CHECK(err)

	_, err = DBClient.UpdateItem(&dynamodb.UpdateItemInput{
		TableName: aws.String(kTablePrefix + tablename),
		Key:       Key,
		// ConditionExpression:       expr.Condition(),
		ExpressionAttributeNames:  expr.Names(),
		ExpressionAttributeValues: expr.Values(),
		UpdateExpression:          expr.Update(),
	})
	if err != nil {
		AssertConditionFailure(err)
	}
}

func WriteWithoutLog(env *Env, tablename string, key string, update map[expression.NameBuilder]expression.OperandBuilder) {
	Key, err := dynamodbattribute.MarshalMap(aws.JSONValue{"K": key})
	CHECK(err)
	condBuilder := expression.Or(
		expression.AttributeNotExists(expression.Name("VERSION")),
		expression.Name("VERSION").LessThan(expression.Value(env.SeqNum)))
	// if _, err = expression.NewBuilder().WithCondition(cond).Build(); err == nil {
	// 	condBuilder = expression.And(condBuilder, cond)
	// }
	updateBuilder := expression.UpdateBuilder{}
	for k, v := range update {
		updateBuilder = updateBuilder.Set(k, v)
	}
	updateBuilder = updateBuilder.
		Set(expression.Name("VERSION"), expression.Value(env.SeqNum))
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
	}
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

func ReadWithoutLog(env *Env, tablename string, key string) interface{} {
	intentLog := env.Fsm.GetStepLog(env.StepNumber)
	if intentLog != nil && checkPostReadLog(intentLog, tablename, key) {
		log.Printf("[INFO] Seen Read log instance %s step %d", env.InstanceId, env.StepNumber)
		env.StepNumber += 1
		env.SeqNum = intentLog.SeqNum
		return intentLog.Data["result"]
	}
	preWriteLog, err := env.FaasEnv.SharedLogReadPrev(env.FaasCtx, DatabaseKeyTag(tablename, key), env.SeqNum)
	CHECK(err)
	var readSeqNum uint64
	if preWriteLog != nil {
		readSeqNum = preWriteLog.SeqNum
	} else {
		readSeqNum = protocol.MaxLogSeqnum
	}
	item := LibReadWithoutLog(tablename, key, readSeqNum)
	if item == nil {
		log.Fatalf("[FATAL] Empty read instance %s step %d: table=%s key=%s seqnum<=%x", env.InstanceId, env.StepNumber, tablename, key, readSeqNum)
	}
	if item["VERSION"] == readSeqNum {
		return item["V"]
	} else if intentLog != nil {
		log.Fatalf("[FATAL] Inconsistent read instance %s step %d: table=%s key=%s seqnum=%x(previously %x)", env.InstanceId, env.StepNumber, tablename, key, item["VERSION"], readSeqNum)
	}
	// the returned version is not the requested, must append readlog for determinism
	// NOTE: the intent step is not cached as we have filterd out intentLog != nil
	newLog, _ := ProposeNextStep(env, nil, aws.JSONValue{
		"type":   "PostRead",
		"key":    key,
		"table":  tablename,
		"result": item["V"],
	})
	// a concurrent step log, may or may not be a post read log
	if !newLog {
		return nil
	}
	return item["V"]
}

func ReadWithLog(env *Env, tablename string, key string) interface{} {
	postReadLog := env.Fsm.GetStepLog(env.StepNumber)
	if postReadLog != nil {
		// NOTE: we assume all keys use the same logging scheme
		// otherwise this might not be fatal
		if !checkPostReadLog(postReadLog, tablename, key) {
			log.Fatalf("[FATAL] Missing read log instance %s step %d: table=%s key=%s", env.InstanceId, env.StepNumber, tablename, key)
		}
		log.Printf("[INFO] Seen read log instance %s step %d", env.InstanceId, env.StepNumber)
		env.StepNumber += 1
		env.SeqNum = postReadLog.SeqNum
		return postReadLog.Data["result"]
	}
	item := LibReadWithLog(tablename, key)
	newLog, _ := ProposeNextStep(env, nil, aws.JSONValue{
		"type":   "PostRead",
		"key":    key,
		"table":  tablename,
		"result": item["V"],
	})
	// a concurrent step log, must be a post read log but not necessarily the same version
	// possible to catch up with the other version, but choose to exit the function for simplicity
	if !newLog {
		return nil
	}
	return item["V"]
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
	var res []interface{}
	for _, item := range items {
		res = append(res, item["V"])
	}
	// log.Printf("[INFO] Finished scanning table %v(%v items): instance %s step %d", tablename, len(res), env.InstanceId, env.StepNumber)
	newLog, _ := ProposeNextStep(env, nil, aws.JSONValue{
		"type":   "PostScan",
		"table":  tablename,
		"result": res,
	})
	if !newLog {
		return nil
	}
	return res
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
