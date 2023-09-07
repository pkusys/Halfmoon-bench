package core

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/dynamodb/expression"
	"github.com/eniac/Beldi/pkg/cayonlib"
	"github.com/mitchellh/mapstructure"
)

func UploadUserReview(env *cayonlib.Env, userId string, reviewId string, timestamp string) {
	reviewInfo := ReviewInfo{ReviewId: reviewId, Timestamp: timestamp}
	item := cayonlib.Read(env, TUserReview(), userId)
	if env.Instruction == "EXIT" {
		return
	}
	if item == nil {
		cayonlib.Write(env, TUserReview(), userId, map[expression.NameBuilder]expression.OperandBuilder{
			expression.Name("V"): expression.Value(aws.JSONValue{"reviews": []ReviewInfo{reviewInfo}}),
		}, true)
	} else {
		cayonlib.Write(env, TUserReview(), userId, map[expression.NameBuilder]expression.OperandBuilder{
			expression.Name("V.reviews"): expression.Name("V.reviews").ListAppend(expression.Value([]ReviewInfo{reviewInfo})),
		}, true)
	}
}

func ReadUserReviews(env *cayonlib.Env, userId string) []Review {
	item := cayonlib.Read(env, TUserReview(), userId)
	if item == nil {
		return []Review{}
	}
	var reviewInfos []ReviewInfo
	cayonlib.CHECK(mapstructure.Decode(item.(map[string]interface{})["reviews"], &reviewInfos))
	var reviewIds []string
	for _, review := range reviewInfos {
		reviewIds = append(reviewIds, review.ReviewId)
	}
	var res []Review
	output, instanceId := cayonlib.SyncInvoke(env, TReviewStorage(), RPCInput{
		Function: "ReadReviews",
		Input:    reviewIds,
	})
	if instanceId == "" || output == nil {
		return []Review{}
	}
	cayonlib.CHECK(mapstructure.Decode(output, &res))
	return res
}
