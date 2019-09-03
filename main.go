package main

import (
	"fmt"
	"math/rand"
	"time"
)

const ShardCount = 4

type Command int8

const (
	Exit Command = 0
	Pause  Command  = 1
	Run Command = 2
)

// Periods in seconds
// Probability as float
var FinalisationPeriod = BoundedRange{3, 5}

const FinalisationProbability = .8

var BlockGenerationPeriod = BoundedRange{1, 3}
var BlockTxInNumber = BoundedRange{1, 4}
var BlockTxOutNumber = BoundedRange{1, 4}

const BlockGenerationProbability = 1
const TXPoolSize = 20

var TXGenerationPeriod = BoundedRange{1, 2}
var TXGenerationNumber = BoundedRange{1, 3}
var ShardRange = BoundedRange{1, ShardCount}

var ProbabilityBuildOnLongestChain = 0.90

var StartTime =	time.Now()
const pixelsPerSecond = 5

const debugShard = 1


func main() {

	fmt.Println("Starting sharding simulator:", ShardCount, " shards.");

	// Init random seed
	rand.Seed(time.Now().UnixNano())

	// Establish communication channels
	channels := Communication{}

	for i := range channels.blocks {
		channels.blocks[i] = make(chan *Block, 100)
	}
	for i := range channels.finalisation {
		channels.finalisation[i] = make(chan *Finalisation, 100)
	}
	for i := range channels.control {
		channels.control[i] = make(chan *Command, 10)
	}

	// Create beacon
	beacon := Beacon{channels: &channels}
	beacon.init()
	go beacon.run()

	// Create shards
	shards := make([]Shard, ShardCount+1)
	for i := 1; i <= ShardCount; i++ {
		shards[i] = Shard{
			id:       i,
			channels: channels,
		}

		shards[i].init()

		go shards[i].run()
	}

	// Start shards
	channels.broadCastCommand(Run)

	// Start visualiser
	visualiser :=  Visualiser{
		shards : shards,
		channels: channels,
		viewShard: 0,
		scaleX: 1,
	}

	visualiser.Run()
}

// BoundedRange (min <= max) in seconds
type BoundedRange struct {
	min int
	max int
}

// Generate random time period in range
func (pr *BoundedRange) NextRandomTimePeriod() time.Duration {
	return time.Duration(pr.min*1000+rand.Intn((pr.max-pr.min)*1000)) * time.Millisecond
}

// Generate random time period in range
func (pr *BoundedRange) NextRandomInt() int {
	return pr.min + rand.Intn(pr.max+1 - pr.min)
}

func MinOf(vars ...int) int {
	min := vars[0]

	for _, i := range vars {
		if min > i {
			min = i
		}
	}

	return min
}
