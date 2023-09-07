package rate

import (
	"log"
	"sort"
	"time"

	"github.com/eniac/Beldi/internal/hotel/main/data"
	"github.com/eniac/Beldi/pkg/cayonlib"
	"github.com/mitchellh/mapstructure"
)

func GetRates(env *cayonlib.Env, req Request) Result {
	var plans RatePlans
	start := time.Now()
	for _, i := range req.HotelIds {
		plan := data.RatePlan{}
		res := cayonlib.Read(env, data.Trate(), i)
		if res == nil {
			continue
		}
		err := mapstructure.Decode(res, &plan)
		cayonlib.CHECK(err)
		if plan.HotelId != "" {
			plans = append(plans, plan)
		}
	}
	elapsed := time.Since(start).Milliseconds()
	log.Printf("rate reads %d hotels in %d ms", len(req.HotelIds), elapsed)
	sort.Sort(plans)
	return Result{RatePlans: plans}
}
