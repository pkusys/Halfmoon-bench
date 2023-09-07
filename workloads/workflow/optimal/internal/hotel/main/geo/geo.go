package geo

import (
	"github.com/eniac/Beldi/internal/hotel/main/data"
	"github.com/eniac/Beldi/pkg/cayonlib"
	"github.com/hailocab/go-geoindex"
	"github.com/mitchellh/mapstructure"
)

func newGeoIndex(env *cayonlib.Env) *geoindex.ClusteringIndex {
	var ps []data.Point
	res := cayonlib.Scan(env, data.Tgeo())
	if res == nil {
		return nil
	}
	err := mapstructure.Decode(res, &ps)
	cayonlib.CHECK(err)
	index := geoindex.NewClusteringIndex()
	for _, e := range ps {
		index.Add(e)
	}
	return index
}

func getNearbyPoints(env *cayonlib.Env, lat float64, lon float64, kNearest int) []geoindex.Point {
	center := &geoindex.GeoPoint{
		Pid:  "",
		Plat: lat,
		Plon: lon,
	}
	index := newGeoIndex(env)
	if index == nil {
		return []geoindex.Point{}
	}
	res := index.KNearest(
		center,
		kNearest,
		geoindex.Km(100), func(p geoindex.Point) bool {
			return true
		},
	)
	return res
}

func Nearby(env *cayonlib.Env, req Request, kNearest int) Result {
	var (
		points = getNearbyPoints(env, req.Lat, req.Lon, kNearest)
	)
	res := Result{HotelIds: []string{}}
	for _, p := range points {
		res.HotelIds = append(res.HotelIds, p.Id())
	}
	return res
}
