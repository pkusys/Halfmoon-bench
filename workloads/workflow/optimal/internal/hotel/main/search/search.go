package search

import (
	"github.com/eniac/Beldi/internal/hotel/main/data"
	"github.com/eniac/Beldi/internal/hotel/main/geo"
	"github.com/eniac/Beldi/internal/hotel/main/rate"
	"github.com/eniac/Beldi/pkg/cayonlib"
	"github.com/mitchellh/mapstructure"
)

func Nearby(env *cayonlib.Env, req Request) Result {
	res, instanceId := cayonlib.SyncInvoke(env, data.Tgeo(), geo.Request{Lat: req.Lat, Lon: req.Lon})
	if instanceId == "" || res == nil {
		return Result{}
	}
	var geoRes geo.Result
	cayonlib.CHECK(mapstructure.Decode(res, &geoRes))
	res, instanceId = cayonlib.SyncInvoke(env, data.Trate(), rate.Request{
		HotelIds: geoRes.HotelIds,
		Indate:   req.InDate,
		Outdate:  req.OutDate,
	})
	if instanceId == "" || res == nil {
		return Result{}
	}
	var rateRes rate.Result
	cayonlib.CHECK(mapstructure.Decode(res, &rateRes))
	var hts []string
	for _, r := range rateRes.RatePlans {
		hts = append(hts, r.HotelId)
	}
	return Result{HotelIds: hts}
}
