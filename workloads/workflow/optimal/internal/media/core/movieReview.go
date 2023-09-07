package core

import (
	"log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/dynamodb/expression"
	"github.com/eniac/Beldi/pkg/cayonlib"
	"github.com/mitchellh/mapstructure"
)

func UploadMovieReview(env *cayonlib.Env, movieId string, reviewId string, timestamp string) {
	reviewInfo := ReviewInfo{ReviewId: reviewId, Timestamp: timestamp}
	item := cayonlib.Read(env, TMovieReview(), movieId)
	if env.Instruction == "EXIT" {
		return
	}
	if item == nil {
		cayonlib.Write(env, TMovieReview(), movieId, map[expression.NameBuilder]expression.OperandBuilder{
			expression.Name("V"): expression.Value(aws.JSONValue{"reviews": []ReviewInfo{reviewInfo}}),
		}, true)
	} else {
		cayonlib.Write(env, TMovieReview(), movieId, map[expression.NameBuilder]expression.OperandBuilder{
			expression.Name("V.reviews"): expression.Name("V.reviews").ListAppend(expression.Value([]ReviewInfo{reviewInfo})),
		}, true)
	}
}

func ReadMovieReviews(env *cayonlib.Env, movieId string) []Review {
	item := cayonlib.Read(env, TMovieReview(), movieId)
	if item == nil {
		log.Printf("no reviews for movie %s", movieId)
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
