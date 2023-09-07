package cayonlib

import (
	"log"
	"sync"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/dynamodb/expression"
)

func ReadWithMode(env *Env, mode string, tablename string, key string) interface{} {
	if mode == "WRITELOG" {
		return ReadWithoutLog(env, tablename, key)
	} else {
		return ReadWithLog(env, tablename, key)
	}
}

// func ReadInTransition(env *Env, tablename string, key string) interface{} {
// 	postReadLog := env.Fsm.GetStepLog(env.StepNumber)
// 	if postReadLog != nil {
// 		if !checkPostReadLog(postReadLog, tablename, key) {
// 			log.Fatalf("[FATAL] Missing read log instance %s step %d: table=%s key=%s", env.InstanceId, env.StepNumber, tablename, key)
// 		}
// 		log.Printf("[INFO] Seen read log instance %s step %d", env.InstanceId, env.StepNumber)
// 		env.StepNumber += 1
// 		env.SeqNum = postReadLog.SeqNum
// 		return postReadLog.Data["result"]
// 	}
// 	var readSingle, readMulti aws.JSONValue
// 	var wg sync.WaitGroup
// 	wg.Add(2)
// 	go func() {
// 		readSingle = LibReadSingleVersion(tablename, key)
// 		wg.Done()
// 	}()
// 	go func() {
// 		// if version log is empty then version is 0 and err is set
// 		version, err := getVersionFromLog(env, tablename, key)
// 		if err == nil {
// 			readMulti = LibReadMultiVersion(tablename, key, version)
// 			if _, ok := readMulti["V"]; ok {
// 				readMulti["VERSION"] = version
// 			}
// 		}
// 		wg.Done()
// 	}()
// 	wg.Wait()
// 	var result interface{}
// 	var version float64
// 	if v, ok := readSingle["VERSION"]; ok && v.(float64) >= version {
// 		result = readSingle["V"]
// 		version = v.(float64)
// 	}
// 	if v, ok := readMulti["VERSION"]; ok && v.(float64) >= version {
// 		result = readMulti["V"]
// 		version = v.(float64)
// 	}
// 	newLog, _ := ProposeNextStep(env, nil, aws.JSONValue{
// 		"type":   "PostRead",
// 		"key":    key,
// 		"table":  tablename,
// 		"result": result,
// 	})
// 	// a concurrent step log, must be a post read log but not necessarily the same version
// 	// possible to catch up with the other version, but choose to exit the function for simplicity
// 	if !newLog {
// 		return nil
// 	}
// 	return result
// }

func WriteWithMode(env *Env, mode string, tablename string, key string, update map[expression.NameBuilder]expression.OperandBuilder) {
	if mode == "WRITELOG" {
		WriteWithLog(env, tablename, key, update, false)
	} else if mode == "READLOG" {
		WriteWithoutLog(env, tablename, key, update, false)
	} else {
		WriteInTransition(env, tablename, key, update)
	}
}

func WriteInTransition(env *Env, tablename string, key string, update map[expression.NameBuilder]expression.OperandBuilder) {
	newLog, preWriteLog := ProposeNextStep(
		env,
		// []uint64{DatabaseKeyTag(tablename, key)},
		nil,
		aws.JSONValue{
			"type": "PreWrite",
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
	var wg sync.WaitGroup
	wg.Add(2)
	go func() {
		LibWriteSingleVersion(tablename, key, env.SeqNum, update)
		wg.Done()
	}()
	go func() {
		LibWriteMultiVersion(tablename, key, env.SeqNum, update)
		wg.Done()
	}()
	ProposeNextStep(
		env,
		[]uint64{DatabaseKeyTag(tablename, key)},
		aws.JSONValue{
			"type": "PostWrite",
			"key":     key,
			"table":   tablename,
			"version": env.SeqNum,
		})
	wg.Wait()
}
