package core

import (
	"sync"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/eniac/Beldi/pkg/cayonlib"
	"github.com/mitchellh/mapstructure"
)

func ReadPage(env *cayonlib.Env, movieId string) Page {
	var movieInfo MovieInfo
	var reviews []Review
	var castInfos []CastInfo
	var plot string
	invoke1 := cayonlib.ProposeInvoke(env, TMovieInfo())
	if invoke1 == nil {
		return Page{}
	}
	invoke2 := cayonlib.ProposeInvoke(env, TCastInfo())
	if invoke2 == nil {
		return Page{}
	}
	invoke3 := cayonlib.ProposeInvoke(env, TPlot())
	if invoke3 == nil {
		return Page{}
	}
	invoke4 := cayonlib.ProposeInvoke(env, TMovieReview())
	if invoke4 == nil {
		return Page{}
	}
	var wg sync.WaitGroup
	wg.Add(4)
	go func() {
		defer wg.Done()
		res, _ := cayonlib.AssignedSyncInvoke(env, TMovieInfo(), RPCInput{
			Function: "ReadMovieInfo",
			Input:    aws.JSONValue{"movieId": movieId},
		}, invoke1)
		cayonlib.CHECK(mapstructure.Decode(res, &movieInfo))
		var ids []string
		for _, cast := range movieInfo.Casts {
			ids = append(ids, cast.CastInfoId)
		}
		go func() {
			defer wg.Done()
			res, _ := cayonlib.AssignedSyncInvoke(env, TCastInfo(), RPCInput{
				Function: "ReadCastInfo",
				Input:    ids,
			}, invoke2)
			cayonlib.CHECK(mapstructure.Decode(res, &castInfos))
		}()
		go func() {
			defer wg.Done()
			res, _ := cayonlib.AssignedSyncInvoke(env, TPlot(), RPCInput{
				Function: "ReadPlot",
				Input:    aws.JSONValue{"plotId": movieInfo.PlotId},
			}, invoke3)
			cayonlib.CHECK(mapstructure.Decode(res, &plot))
		}()
	}()
	go func() {
		defer wg.Done()
		res, _ := cayonlib.AssignedSyncInvoke(env, TMovieReview(), RPCInput{
			Function: "ReadMovieReviews",
			Input:    aws.JSONValue{"movieId": movieId},
		}, invoke4)
		cayonlib.CHECK(mapstructure.Decode(res, &reviews))
	}()
	wg.Wait()
	return Page{CastInfos: castInfos, Reviews: reviews, MovieInfo: movieInfo, Plot: plot}
}
