module cs.utexas.edu/zjia/faas-retwis

go 1.14

require (
	cs.utexas.edu/zjia/faas v0.0.0
	github.com/cespare/xxhash/v2 v2.2.0
	github.com/davecgh/go-spew v1.1.1 // indirect
	github.com/golang/snappy v0.0.2
	github.com/google/uuid v1.3.0 // indirect
	github.com/kr/pretty v0.1.0 // indirect
	github.com/lithammer/shortuuid v3.0.0+incompatible
	github.com/montanaflynn/stats v0.6.3
	github.com/openacid/low v0.1.21
	gopkg.in/check.v1 v1.0.0-20180628173108-788fd7840127 // indirect
)

replace cs.utexas.edu/zjia/faas => ../../boki/worker/golang

replace cs.utexas.edu/zjia/faas/slib => ../../boki/slib
