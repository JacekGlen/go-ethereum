package main

import (
	"flag"
	"fmt"
	"math"
	"math/big"
	"os"
	"strings"
	"testing"
	"time"

	_ "unsafe"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/rawdb"
	"github.com/ethereum/go-ethereum/core/state"
	"github.com/ethereum/go-ethereum/core/vm/runtime"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/params"
)

var result []byte
var bytecode []byte
var bytecodeStore string

func BenchmarkBytecodeExecutionSerial(b *testing.B) {
	var calldata []byte

	cfg := new(runtime.Config)
	setDefaults(cfg)
	// from `github.com/ethereum/go-ethereum/core/vm/runtime/runtime.go:109`
	cfg.State, _ = state.New(common.Hash{}, state.NewDatabase(rawdb.NewMemoryDatabase()), nil)

	// Initialize some constant calldata of 128KB, 2^17 bytes.
	// This means, if we offset between 0th and 2^16th byte, we can fetch between 0 and 2^16 bytes (64KB)
	// In consequence, we need args to memory-copying OPCODEs to be between 0 and 2^16, 2^16 fits in a PUSH3,
	// which we'll be using to generate arguments for those OPCODEs.
	calldata = []byte(strings.Repeat("{", 1<<17))

	// Warm-up. **NOTE** we're keeping tracing on during warm-up, otherwise measurements are off
	cfg.EVMConfig.Debug = false
	_, _, errWarmUp := runtime.Execute(bytecode, calldata, cfg)

	if errWarmUp != nil {
		b.Error()
		return
	}

	b.ResetTimer()

	var exBytes []byte

	for i := 0; i < b.N; i++ {
		exBytes2, _, err := runtime.Execute(bytecode, calldata, cfg)
		exBytes = exBytes2

		if err != nil {
			fmt.Fprintln(os.Stderr, err)
		}
	}

	result = exBytes
}

func BenchmarkBytecodeExecutionIndividualWithWarmup(b *testing.B) {
	var calldata []byte
	var exBytes []byte

	for i := 0; i < b.N; i++ {
		b.StopTimer()

		cfg := new(runtime.Config)
		setDefaults(cfg)
		// from `github.com/ethereum/go-ethereum/core/vm/runtime/runtime.go:109`
		cfg.State, _ = state.New(common.Hash{}, state.NewDatabase(rawdb.NewMemoryDatabase()), nil)

		// Initialize some constant calldata of 128KB, 2^17 bytes.
		// This means, if we offset between 0th and 2^16th byte, we can fetch between 0 and 2^16 bytes (64KB)
		// In consequence, we need args to memory-copying OPCODEs to be between 0 and 2^16, 2^16 fits in a PUSH3,
		// which we'll be using to generate arguments for those OPCODEs.
		calldata = []byte(strings.Repeat("{", 1<<17))

		// Warm-up. **NOTE** we're keeping tracing on during warm-up, otherwise measurements are off
		cfg.EVMConfig.Debug = false

		b.StartTimer()

		exBytes2, _, err := runtime.Execute(bytecode, calldata, cfg)
		exBytes = exBytes2

		if err != nil {
			fmt.Fprintln(os.Stderr, err)
		}
	}

	result = exBytes
}

func BenchmarkBytecodeExecutionIndividual(b *testing.B) {
	var calldata []byte
	var exBytes []byte

	for i := 0; i < b.N; i++ {
		b.StopTimer()

		cfg := new(runtime.Config)
		setDefaults(cfg)
		// from `github.com/ethereum/go-ethereum/core/vm/runtime/runtime.go:109`
		cfg.State, _ = state.New(common.Hash{}, state.NewDatabase(rawdb.NewMemoryDatabase()), nil)

		// Initialize some constant calldata of 128KB, 2^17 bytes.
		// This means, if we offset between 0th and 2^16th byte, we can fetch between 0 and 2^16 bytes (64KB)
		// In consequence, we need args to memory-copying OPCODEs to be between 0 and 2^16, 2^16 fits in a PUSH3,
		// which we'll be using to generate arguments for those OPCODEs.
		calldata = []byte(strings.Repeat("{", 1<<17))

		// Warm-up. **NOTE** we're keeping tracing on during warm-up, otherwise measurements are off
		cfg.EVMConfig.Debug = false
		_, _, errWarmUp := runtime.Execute(bytecode, calldata, cfg)

		if errWarmUp != nil {
			b.Error()
			return
		}

		b.StartTimer()

		exBytes2, _, err := runtime.Execute(bytecode, calldata, cfg)
		exBytes = exBytes2

		if err != nil {
			fmt.Fprintln(os.Stderr, err)
		}
	}

	result = exBytes
}

// // copied directly from github.com/ethereum/go-ethereum/core/vm/runtime/runtime.go
// // so that we skip this in measured code
// sets defaults on the config
func setDefaults(cfg *runtime.Config) {
	if cfg.ChainConfig == nil {
		cfg.ChainConfig = &params.ChainConfig{
			ChainID:             big.NewInt(1),
			HomesteadBlock:      new(big.Int),
			DAOForkBlock:        new(big.Int),
			DAOForkSupport:      false,
			EIP150Block:         new(big.Int),
			EIP150Hash:          common.Hash{},
			EIP155Block:         new(big.Int),
			EIP158Block:         new(big.Int),
			ByzantiumBlock:      new(big.Int),
			ConstantinopleBlock: new(big.Int),
			PetersburgBlock:     new(big.Int),
			IstanbulBlock:       new(big.Int),
			MuirGlacierBlock:    new(big.Int),
			BerlinBlock:         new(big.Int),
			LondonBlock:         new(big.Int),
		}
	}

	if cfg.Difficulty == nil {
		cfg.Difficulty = new(big.Int)
	}
	if cfg.Time == nil {
		cfg.Time = big.NewInt(time.Now().Unix())
	}
	if cfg.GasLimit == 0 {
		cfg.GasLimit = math.MaxUint64
	}
	if cfg.GasPrice == nil {
		cfg.GasPrice = new(big.Int)
	}
	if cfg.Value == nil {
		cfg.Value = new(big.Int)
	}
	if cfg.BlockNumber == nil {
		cfg.BlockNumber = new(big.Int)
	}
	if cfg.GetHashFn == nil {
		cfg.GetHashFn = func(n uint64) common.Hash {
			return common.BytesToHash(crypto.Keccak256([]byte(new(big.Int).SetUint64(n).String())))
		}
	}
	if cfg.BaseFee == nil {
		cfg.BaseFee = big.NewInt(params.InitialBaseFee)
	}
}

func runBenchmark(sampleSize int) {

	bytecode = common.Hex2Bytes(bytecodeStore)

	for i := 0; i < sampleSize; i++ {
		r1 := testing.Benchmark(BenchmarkBytecodeExecutionSerial)
		outputResults("BenchmarkBytecodeExecutionSerial", i, r1)

		r2 := testing.Benchmark(BenchmarkBytecodeExecutionIndividualWithWarmup)
		outputResults("BenchmarkBytecodeExecutionIndividualWithWarmup", i, r2)

		r3 := testing.Benchmark(BenchmarkBytecodeExecutionIndividual)
		outputResults("BenchmarkBytecodeExecutionIndividual", i, r3)
	}
}

func outputResults(desc string, sampleId int, r testing.BenchmarkResult) {
	fmt.Printf("%s,%d,%v,%v\n", desc, sampleId, r.N, r.NsPerOp())
	fmt.Printf("%v, %v\n", r.AllocsPerOp(), r.AllocedBytesPerOp())
}

func runOverheadBenchmark(sampleSize int) {
	for i := 0; i < sampleSize; i++ {

		bytecode = common.Hex2Bytes("00" + bytecodeStore)
		rEmpty := testing.Benchmark(BenchmarkBytecodeExecutionSerial)

		bytecode = common.Hex2Bytes(bytecodeStore)
		rActual := testing.Benchmark(BenchmarkBytecodeExecutionSerial)

		outputOverheadResults(i, rEmpty, rActual)
	}
}

func outputOverheadResults(sampleId int, rEmpty testing.BenchmarkResult, rActual testing.BenchmarkResult) {
	overheadTime := rEmpty.NsPerOp()
	var executionLoopTime int64 = 0
	var totalTime int64 = rEmpty.NsPerOp()

	if rActual.NsPerOp() > rEmpty.NsPerOp() {
		executionLoopTime = rActual.NsPerOp() - rEmpty.NsPerOp()
		totalTime = rActual.NsPerOp()
	}

	fmt.Printf("%v,%v,%v,%v,%v,%v,%v\n", sampleId, rActual.N, overheadTime, executionLoopTime, totalTime, rActual.AllocsPerOp(), rActual.AllocedBytesPerOp())
}

func main() {
	bytecodePtr := flag.String("bytecode", "", "EVM bytecode to execute and measure")
	sampleSizePtr := flag.Int("sampleSize", 1, "Size of the sample - number of measured repetitions of execution")

	flag.Parse()

	bytecodeStore = *bytecodePtr
	sampleSize := *sampleSizePtr

	runOverheadBenchmark(sampleSize)
}
