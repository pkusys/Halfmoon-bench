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
	invoke1 := cayonlib.ProposeInvoke(env, TMovieInfo(), RPCInput{
		Function: "ReadMovieInfo",
		Input:    aws.JSONValue{"movieId": movieId},
	})
	if invoke1 == nil {
		return Page{}
	}
	// invoke2&3 is proposed after invoke1 has executed because of dependency
	invoke4 := cayonlib.ProposeInvoke(env, TMovieReview(), RPCInput{
		Function: "ReadMovieReviews",
		Input:    aws.JSONValue{"movieId": movieId},
	})
	if invoke4 == nil {
		return Page{}
	}
	var wg sync.WaitGroup
	wg.Add(4)
	go func() {
		defer wg.Done()
		res, _ := cayonlib.AssignedSyncInvoke(env, TMovieInfo(), invoke1)
		cayonlib.CHECK(mapstructure.Decode(res, &movieInfo))
		var ids []string
		for _, cast := range movieInfo.Casts {
			ids = append(ids, cast.CastInfoId)
		}
		invoke2 := cayonlib.ProposeInvoke(env, TCastInfo(), RPCInput{
			Function: "ReadCastInfo",
			Input:    ids,
		})
		if invoke2 == nil {
			return
		}
		invoke3 := cayonlib.ProposeInvoke(env, TPlot(), RPCInput{
			Function: "ReadPlot",
			Input:    aws.JSONValue{"plotId": movieInfo.PlotId},
		})
		if invoke3 == nil {
			return
		}
		go func() {
			defer wg.Done()
			res, _ := cayonlib.AssignedSyncInvoke(env, TCastInfo(), invoke2)
			cayonlib.CHECK(mapstructure.Decode(res, &castInfos))
		}()
		go func() {
			defer wg.Done()
			res, _ := cayonlib.AssignedSyncInvoke(env, TPlot(), invoke3)
			cayonlib.CHECK(mapstructure.Decode(res, &plot))
		}()
	}()
	go func() {
		defer wg.Done()
		res, _ := cayonlib.AssignedSyncInvoke(env, TMovieReview(), invoke4)
		cayonlib.CHECK(mapstructure.Decode(res, &reviews))
	}()
	wg.Wait()
	if env.Instruction == "EXIT" {
		return Page{}
	}
	return Page{CastInfos: castInfos, Reviews: reviews, MovieInfo: movieInfo, Plot: plot}
}
