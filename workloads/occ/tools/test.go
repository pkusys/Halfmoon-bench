package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"math"
	"math/rand"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"cs.utexas.edu/zjia/faas-retwis/handlers"
	"cs.utexas.edu/zjia/faas-retwis/utils"
	"cs.utexas.edu/zjia/faas/types"
)

const (
	TestWriteReq int = iota
	TestReadReq
)

type Cache struct {
	slots    int
	maxSlots int
	records  []*types.LogEntry
	index    map[uint64]int
}

func (c *Cache) insert(seqNum uint64, record *types.LogEntry) {
	if idx, ok := c.index[seqNum]; ok {
		c.records[idx] = record
		return
	}
	var idx int
	if c.slots < c.maxSlots {
		idx = c.slots
		c.slots++
	} else {
		idx = rand.Int() % c.maxSlots
		oldSeq := c.records[idx].SeqNum
		delete(c.index, oldSeq)
	}
	c.index[seqNum] = idx
	c.records[idx] = record
}

func (c *Cache) get(seqNum uint64) *types.LogEntry {
	if idx, ok := c.index[seqNum]; ok {
		return c.records[idx]
	}
	return nil
}

type Backend struct {
	mu    sync.Mutex
	index map[uint64][]uint64
	cache Cache
	// recordCache map[uint64]types.LogEntry
	// auxDataCache map[uint64][]byte
	seqnum uint64
}

func (b *Backend) InvokeFunc(ctx context.Context, funcName string, input []byte) ( /* output */ []byte, error) {
	return nil, nil
}
func (b *Backend) InvokeFuncAsync(ctx context.Context, funcName string, input []byte) error {
	return nil
}
func (b *Backend) GrpcCall(ctx context.Context, service string, method string, request []byte) ( /* reply */ []byte, error) {
	return nil, nil
}

func (b *Backend) GenerateUniqueID() uint64 { return 0 }

// Shared log operations
// Append a new log entry, tags must be non-zero
func (b *Backend) SharedLogAppend(ctx context.Context, tags []uint64, data []byte) ( /* seqnum */ uint64, error) {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.seqnum++
	for _, t := range tags {
		if _, ok := b.index[t]; ok {
			b.index[t] = append(b.index[t], b.seqnum)
		} else {
			b.index[t] = []uint64{b.seqnum}
		}
	}
	b.cache.insert(b.seqnum, &types.LogEntry{SeqNum: b.seqnum, Tags: tags, Data: data, AuxData: nil})
	// b.recordCache[b.seqnum] = types.LogEntry{SeqNum: b.seqnum, Tags: tags, Data: data, AuxData: nil}
	return b.seqnum, nil
}

// Read the first log with `tag` whose seqnum >= given `seqNum`
// `tag`==0 means considering log with any tag, including empty tag
func (b *Backend) SharedLogReadNext(ctx context.Context, tag uint64, seqNum uint64) (*types.LogEntry, error) {
	return nil, nil
}
func (b *Backend) SharedLogReadNextBlock(ctx context.Context, tag uint64, seqNum uint64) (*types.LogEntry, error) {
	return nil, nil
}

// Read the last log with `tag` whose seqnum <= given `seqNum`
// `tag`==0 means considering log with any tag, including empty tag
func (b *Backend) SharedLogReadPrev(ctx context.Context, tag uint64, seqNum uint64) (*types.LogEntry, error) {
	b.mu.Lock()
	defer b.mu.Unlock()
	idxForTag, ok := b.index[tag]
	if !ok {
		// panic("read prev: tag not exist")
		return nil, nil
	}
	// found:=false
	// entry:=types.LogEntry{}
	for i := len(idxForTag) - 1; i >= 0; i-- {
		if idxForTag[i] < seqNum {
			// entry := b.recordCache[idxForTag[i]]
			entry := b.cache.get(idxForTag[i])
			if entry == nil {
				b.index[tag] = idxForTag[i+1:]
			}
			return entry, nil
		}
	}
	return nil, nil
}

// Alias for ReadPrev(tag, MaxSeqNum)
func (b *Backend) SharedLogCheckTail(ctx context.Context, tag uint64) (*types.LogEntry, error) {
	if tag == 0 {
		b.mu.Lock()
		// entry := b.recordCache[b.seqnum]
		entry := b.cache.get(b.seqnum)
		b.mu.Unlock()
		return entry, nil
	}
	return b.SharedLogReadPrev(ctx, tag, math.MaxUint64)
}

// Set auxiliary data for log entry of given `seqNum`
func (b *Backend) SharedLogSetAuxData(ctx context.Context, seqNum uint64, auxData []byte) error {
	b.mu.Lock()
	defer b.mu.Unlock()
	entry := b.cache.get(seqNum)
	if entry == nil {
		log.Println("[WARN] set aux on non-existing entry in cache")
		return nil
	}
	newEntry := *entry
	newEntry.AuxData = auxData
	b.cache.insert(seqNum, &newEntry)
	return nil
}

type request struct {
	op int
	// input handlers.TestInput
	input utils.JSONValue
}

type FrontEnd struct {
	wh      types.FuncHandler
	rh      types.FuncHandler
	reqChan chan request
	ctx     context.Context
}

func newFrontEnd(backend types.Environment, concurrency int) *FrontEnd {
	return &FrontEnd{
		wh:      handlers.NewTestWriteHandler(backend),
		rh:      handlers.NewTestReadHandler(backend),
		reqChan: make(chan request, concurrency),
	}
}

func (f *FrontEnd) serve(resChan chan interface{}) {
	var finished int32
	recved := 0
	wg := &sync.WaitGroup{}
	for req := range f.reqChan {
		recved++
		var handler types.FuncHandler
		switch req.op {
		case TestReadReq:
			handler = f.rh
		case TestWriteReq:
			handler = f.wh
		default:
			panic("invalid request")
		}
		wg.Add(1)
		go func(handler types.FuncHandler, input utils.JSONValue) {
			encoded, err := json.Marshal(input)
			if err != nil {
				log.Fatalf("[FATAL] Failed to encode JSON request: %v", err)
			}
			output, err := handler.Call(f.ctx, encoded)
			if err != nil {
				panic(err.Error())
			}
			var response utils.JSONValue
			if json.Unmarshal(output, &response) != nil {
				panic(err.Error())
			}
			if !response["success"].(bool) {
				panic("req failed")
			}
			// resChan <- struct{}{}
			atomic.AddInt32(&finished, 1)
			wg.Done()
		}(handler, req.input)
	}
	wg.Wait()
	fmt.Println(finished, "/", recved)
	close(resChan)
}

var FLAGS_faas_gateway string
var FLAGS_fn_prefix string

// var FLAGS_num_users int
var FLAGS_concurrency int
var FLAGS_duration int
var FLAGS_percentages string

// var FLAGS_bodylen int
var FLAGS_rand_seed int
var FLAGS_zipf_skew float64
var FLAGS_keyspace int
var FLAGS_rw_keys int
var rng *rand.Zipf

func init() {
	flag.StringVar(&FLAGS_faas_gateway, "faas_gateway", "127.0.0.1:8081", "")
	flag.StringVar(&FLAGS_fn_prefix, "fn_prefix", "", "")
	// flag.IntVar(&FLAGS_num_users, "num_users", 1000, "")
	flag.IntVar(&FLAGS_concurrency, "concurrency", 32, "")
	flag.IntVar(&FLAGS_duration, "duration", 5, "")
	flag.StringVar(&FLAGS_percentages, "percentages", "50,50", "write,read")
	// flag.IntVar(&FLAGS_bodylen, "bodylen", 64, "")
	flag.IntVar(&FLAGS_rand_seed, "rand_seed", 23333, "")
	flag.Float64Var(&FLAGS_zipf_skew, "zipf_skew", 1.1, "")
	flag.IntVar(&FLAGS_rw_keys, "rw_keys", 1, "")
	flag.IntVar(&FLAGS_keyspace, "keyspace", 10000, "")

	// rand.Seed(int64(FLAGS_rand_seed))
}

func parsePercentages(s string) ([]int, error) {
	parts := strings.Split(s, ",")
	if len(parts) != 2 {
		return nil, fmt.Errorf("Need exactly four parts splitted by comma")
	}
	results := make([]int, 2)
	for i, part := range parts {
		if parsed, err := strconv.Atoi(part); err != nil {
			return nil, fmt.Errorf("Failed to parse %d-th part", i)
		} else {
			results[i] = parsed
		}
	}
	for i := 1; i < len(results); i++ {
		results[i] += results[i-1]
	}
	if results[len(results)-1] != 100 {
		return nil, fmt.Errorf("Sum of all parts is not 100")
	}
	return results, nil
}

func buildTestRequest() utils.JSONValue {
	tags := make([]uint64, 0, FLAGS_rw_keys)
	for i := 0; i < FLAGS_rw_keys; i++ {
		tags = append(tags, rng.Uint64()+1)
	}
	return utils.JSONValue{"keys": tags}
}

func main() {
	flag.Parse()

	src := rand.NewSource(int64(FLAGS_rand_seed))
	rnd := rand.New(src)
	rng = rand.NewZipf(rnd, FLAGS_zipf_skew, 1, uint64(FLAGS_keyspace)-1)

	percentages, err := parsePercentages(FLAGS_percentages)
	if err != nil {
		log.Fatalf("[FATAL] Invalid \"percentages\" flag: %v", err)
	}

	log.Printf("[INFO] Start running for %d seconds with concurrency of %d", FLAGS_duration, FLAGS_concurrency)

	backend := &Backend{
		index: make(map[uint64][]uint64),
		cache: Cache{
			maxSlots: 4096,
			records:  make([]*types.LogEntry, 4096),
			index:    make(map[uint64]int),
		},
	}
	resChan := make(chan interface{})
	frontend := newFrontEnd(backend, FLAGS_concurrency)
	go frontend.serve(resChan)

	// client := utils.NewFaasClient(FLAGS_faas_gateway, FLAGS_concurrency)
	startTime := time.Now()
	for {
		if time.Since(startTime) > time.Duration(FLAGS_duration)*time.Second {
			break
		}
		k := rand.Intn(100)
		if k < percentages[0] {
			frontend.reqChan <- request{TestWriteReq, buildTestRequest()}
		} else if k < percentages[1] {
			frontend.reqChan <- request{TestReadReq, buildTestRequest()}
		}
	}
	close(frontend.reqChan)
	<-resChan
}
