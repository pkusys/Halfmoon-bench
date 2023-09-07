.PHONY: build clean

build:
# single operation

singleop:
	env GOOS=linux go build -ldflags="-s -w -X github.com/eniac/Beldi/pkg/cayonlib.TYPE=BELDI" -o bin/singleop/init internal/singleop/init/init.go
	env GOOS=linux go build -ldflags="-s -w -X github.com/eniac/Beldi/pkg/cayonlib.TYPE=BELDI" -o bin/singleop/singleop internal/singleop/main/main.go
	env GOOS=linux go build -ldflags="-s -w -X github.com/eniac/Beldi/pkg/cayonlib.TYPE=BELDI" -o bin/singleop/nop internal/singleop/nop/nop.go

hotel:
	env GOOS=linux go build -ldflags="-s -w -X github.com/eniac/Beldi/pkg/cayonlib.TYPE=BELDI" -o bin/hotel/init internal/hotel/init/init.go
	env GOOS=linux go build -ldflags="-s -w -X github.com/eniac/Beldi/pkg/cayonlib.TYPE=BELDI" -o bin/hotel/geo internal/hotel/main/handlers/geo/geo.go
	env GOOS=linux go build -ldflags="-s -w -X github.com/eniac/Beldi/pkg/cayonlib.TYPE=BELDI" -o bin/hotel/profile internal/hotel/main/handlers/profile/profile.go
	env GOOS=linux go build -ldflags="-s -w -X github.com/eniac/Beldi/pkg/cayonlib.TYPE=BELDI" -o bin/hotel/rate internal/hotel/main/handlers/rate/rate.go
	env GOOS=linux go build -ldflags="-s -w -X github.com/eniac/Beldi/pkg/cayonlib.TYPE=BELDI" -o bin/hotel/recommendation internal/hotel/main/handlers/recommendation/recommendation.go
	env GOOS=linux go build -ldflags="-s -w -X github.com/eniac/Beldi/pkg/cayonlib.TYPE=BELDI" -o bin/hotel/user internal/hotel/main/handlers/user/user.go
	env GOOS=linux go build -ldflags="-s -w -X github.com/eniac/Beldi/pkg/cayonlib.TYPE=BELDI" -o bin/hotel/search internal/hotel/main/handlers/search/search.go
	env GOOS=linux go build -ldflags="-s -w -X github.com/eniac/Beldi/pkg/cayonlib.TYPE=BELDI" -o bin/hotel/hotel internal/hotel/main/handlers/hotel/hotel.go
	env GOOS=linux go build -ldflags="-s -w -X github.com/eniac/Beldi/pkg/cayonlib.TYPE=BELDI" -o bin/hotel/flight internal/hotel/main/handlers/flight/flight.go
	env GOOS=linux go build -ldflags="-s -w -X github.com/eniac/Beldi/pkg/cayonlib.TYPE=BELDI" -o bin/hotel/order internal/hotel/main/handlers/order/order.go
	env GOOS=linux go build -ldflags="-s -w -X github.com/eniac/Beldi/pkg/cayonlib.TYPE=BELDI" -o bin/hotel/frontend internal/hotel/main/handlers/frontend/frontend.go
	env GOOS=linux go build -ldflags="-s -w -X github.com/eniac/Beldi/pkg/cayonlib.TYPE=BELDI" -o bin/hotel/gateway internal/hotel/main/handlers/gateway/gateway.go
	# env GOOS=linux go build -ldflags="-s -w -X github.com/eniac/Beldi/pkg/cayonlib.TYPE=BELDI" -o bin/hotel/gc internal/hotel/main/gc/gc.go
	# env GOOS=linux go build -ldflags="-s -w -X github.com/eniac/Beldi/pkg/cayonlib.TYPE=BELDI" -o bin/hotel/collector internal/hotel/main/collector/collector.go

media:
	env GOOS=linux go build -ldflags="-s -w -X github.com/eniac/Beldi/pkg/cayonlib.TYPE=BELDI" -o bin/media/init internal/media/init/init.go
	env GOOS=linux go build -ldflags="-s -w -X github.com/eniac/Beldi/pkg/cayonlib.TYPE=BELDI" -o bin/media/CastInfo internal/media/core/handlers/castInfo/main.go
	env GOOS=linux go build -ldflags="-s -w -X github.com/eniac/Beldi/pkg/cayonlib.TYPE=BELDI" -o bin/media/ComposeReview internal/media/core/handlers/composeReview/main.go
	env GOOS=linux go build -ldflags="-s -w -X github.com/eniac/Beldi/pkg/cayonlib.TYPE=BELDI" -o bin/media/Frontend internal/media/core/handlers/frontend/main.go
	env GOOS=linux go build -ldflags="-s -w -X github.com/eniac/Beldi/pkg/cayonlib.TYPE=BELDI" -o bin/media/MovieId internal/media/core/handlers/movieId/main.go
	env GOOS=linux go build -ldflags="-s -w -X github.com/eniac/Beldi/pkg/cayonlib.TYPE=BELDI" -o bin/media/MovieInfo internal/media/core/handlers/movieInfo/main.go
	env GOOS=linux go build -ldflags="-s -w -X github.com/eniac/Beldi/pkg/cayonlib.TYPE=BELDI" -o bin/media/MovieReview internal/media/core/handlers/movieReview/main.go
	env GOOS=linux go build -ldflags="-s -w -X github.com/eniac/Beldi/pkg/cayonlib.TYPE=BELDI" -o bin/media/Page internal/media/core/handlers/page/main.go
	env GOOS=linux go build -ldflags="-s -w -X github.com/eniac/Beldi/pkg/cayonlib.TYPE=BELDI" -o bin/media/Plot internal/media/core/handlers/plot/main.go
	env GOOS=linux go build -ldflags="-s -w -X github.com/eniac/Beldi/pkg/cayonlib.TYPE=BELDI" -o bin/media/Rating internal/media/core/handlers/rating/main.go
	env GOOS=linux go build -ldflags="-s -w -X github.com/eniac/Beldi/pkg/cayonlib.TYPE=BELDI" -o bin/media/ReviewStorage internal/media/core/handlers/reviewStorage/main.go
	env GOOS=linux go build -ldflags="-s -w -X github.com/eniac/Beldi/pkg/cayonlib.TYPE=BELDI" -o bin/media/Text internal/media/core/handlers/text/main.go
	env GOOS=linux go build -ldflags="-s -w -X github.com/eniac/Beldi/pkg/cayonlib.TYPE=BELDI" -o bin/media/UniqueId internal/media/core/handlers/uniqueId/main.go
	env GOOS=linux go build -ldflags="-s -w -X github.com/eniac/Beldi/pkg/cayonlib.TYPE=BELDI" -o bin/media/User internal/media/core/handlers/user/main.go
	env GOOS=linux go build -ldflags="-s -w -X github.com/eniac/Beldi/pkg/cayonlib.TYPE=BELDI" -o bin/media/UserReview internal/media/core/handlers/userReview/main.go
	# env GOOS=linux go build -ldflags="-s -w -X github.com/eniac/Beldi/pkg/cayonlib.TYPE=BELDI" -o bin/media/gc internal/media/core/gc/gc.go
	# env GOOS=linux go build -ldflags="-s -w -X github.com/eniac/Beldi/pkg/cayonlib.TYPE=BELDI" -o bin/media/collector internal/media/core/collector/collector.go

rw:
	env GOOS=linux go build -ldflags="-s -w -X github.com/eniac/Beldi/pkg/cayonlib.TYPE=BELDI" -o bin/rw/init internal/rw/init/init.go
	env GOOS=linux go build -ldflags="-s -w -X github.com/eniac/Beldi/pkg/cayonlib.TYPE=BELDI" -o bin/rw/rw internal/rw/main/main.go

recovery:
	env GOOS=linux go build -ldflags="-s -w -X github.com/eniac/Beldi/pkg/cayonlib.TYPE=BELDI" -o bin/recovery/init internal/recovery/init/init.go
	env GOOS=linux go build -ldflags="-s -w -X github.com/eniac/Beldi/pkg/cayonlib.TYPE=BELDI" -o bin/recovery/recovery internal/recovery/main/main.go

retwis:
	env GOOS=linux go build -o bin/retwis/init internal/retwis/init/init.go
	env GOOS=linux go build -o bin/retwis/login internal/retwis/handlers/login/main.go
	env GOOS=linux go build -o bin/retwis/profile internal/retwis/handlers/profile/main.go
	env GOOS=linux go build -o bin/retwis/post internal/retwis/handlers/post/main.go
	env GOOS=linux go build -o bin/retwis/publish internal/retwis/handlers/publish/main.go
	env GOOS=linux go build -o bin/retwis/timeline internal/retwis/handlers/timeline/main.go

clean:
	rm -rf ./bin
