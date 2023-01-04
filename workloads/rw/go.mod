module cs.utexas.edu/zjia/faas-retwis

go 1.14

require (
	cs.utexas.edu/zjia/faas v0.0.0
	github.com/golang/snappy v0.0.2 // indirect
	github.com/google/go-cmp v0.5.5 // indirect
	github.com/montanaflynn/stats v0.6.3
	github.com/openacid/low v0.1.21 // indirect
	github.com/stretchr/testify v1.7.0 // indirect
	go.mongodb.org/mongo-driver v1.4.6
	golang.org/x/net v0.0.0-20201202161906-c7110b5ffcbb // indirect
	golang.org/x/sync v0.0.0-20201020160332-67f06af15bc9 // indirect
	golang.org/x/xerrors v0.0.0-20200804184101-5ec99f83aff1 // indirect
	gopkg.in/yaml.v2 v2.3.0 // indirect
)

replace cs.utexas.edu/zjia/faas => ../../boki/worker/golang

replace cs.utexas.edu/zjia/faas/slib => ../../boki/slib
