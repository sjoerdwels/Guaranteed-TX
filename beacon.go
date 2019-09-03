package main

import (
	"fmt"
	"math/rand"
	"time"
)

type Beacon struct {
	channels     *Communication
	chains       []Chain
	finalisation Finalisation
}

func (beacon *Beacon) init() {

	// Init chains
	beacon.chains = make([]Chain, ShardCount+1)
	for i, _ := range beacon.chains {
		beacon.chains[i] = Chain{}
		beacon.chains[i].init(i)
	}

	// Init finalisation
	beacon.finalisation = Finalisation{
		height:         0,
		blocks:         nil,
		inconsistentTX: make([]*Transaction, 0),
	}

}

func (beacon *Beacon) run() {

	fmt.Println("Launching beacon shard...")

	state := Pause

	for {

		select {
		case command := <-beacon.channels.control[0]:
			switch *command {
			case Run:
				state = Run
			case Pause:
				state = Pause
			case Exit:
				return
			}
		default:

			if state == Pause {
				break
			}

			select {
			case command := <-beacon.channels.control[0]:
				switch *command {
				case Run:
					state = Run
				case Pause:
					state = Pause
					break;
				case Exit:
					return
				}

			case block := <-beacon.channels.blocks[0]:
				beacon.receiveBlock(block)

			case <-time.After(FinalisationPeriod.NextRandomTimePeriod()):
				beacon.proposeFinalisation()
			}

		case <-beacon.channels.finalisation[0]:
			beacon.Println("Received finalisation - all ready processed to prevent race conditions.")

		}
	}
}

func (beacon *Beacon) receiveBlock(block *Block) {

	beacon.Println(fmt.Sprintf("Received block from  %d.",  block.Shard))

	// Add block
	beacon.chains[block.Shard].Insert(block)
}

func (beacon *Beacon) proposeFinalisation() {

	// Probability finalisation fails
	if rand.Float64() > FinalisationProbability {

		beacon.Println("BeaconChain finalisation skipped.")
		return
	}

	beacon.Println("Start finalisation process.")

	shardFinalisations := make([]ShardFinalisation, ShardCount)

	// Prepare finalisation object for each shard
	for i := 0; i < ShardCount; i++ {
		shard := i + 1
		txOut := make([]*Transaction, 0)
		txIn := make([]*Transaction, 0)
		shardFinalisations[i] = ShardFinalisation{
			shard:               shard,
			canonicalChainBlock: beacon.chains[shard].GetLongestChains(1, false)[0],
			newFinalisedBlock:   beacon.chains[shard].lastFinalisedBlock,
			TXOut:               &txOut,
			TXIn:                &txIn,
		}
	}

	// Add inconsistent transactions to each finalisation objects
	for _, tx := range beacon.finalisation.inconsistentTX {
		txOut := shardFinalisations[tx.SourceShard-1].TXOut
		*txOut = append(*txOut, tx)
	}

	// Loop over all finalisation objects to include a new finalisation block until no block can be added anymore.
	running := true

	for running {

		running = false

		for shardIndex := 0; shardIndex < len(shardFinalisations); shardIndex++ {
			if tryFinaliseNextBlock(shardIndex, &shardFinalisations) {
				running = true
			}
		}
	}

	// Propose new finalisation
	inconsistentTX := make([]*Transaction, 0)
	blocks := make([]Block, 0)

	for _, finalisation := range shardFinalisations {
		inconsistentTX = append(inconsistentTX, *finalisation.TXOut...)
		blocks = append(blocks, *finalisation.newFinalisedBlock.block)
		//beacon.Println(fmt.Sprintf("Shard %d finalised uptill block: %x - TXin: %d - TXout: %d", finalisation.shard, finalisation.newFinalisedBlock.block.Hash, len(*finalisation.TXIn), len(*finalisation.TXOut)))
	}

	finalisation := Finalisation{
		height:         beacon.finalisation.height + 1,
		inconsistentTX: inconsistentTX,
		blocks:         blocks,
	}

	beacon.processFinalisation(&finalisation)

	beacon.channels.broadcastFinalisation(&finalisation)
}

func (beacon *Beacon) processFinalisation(finalisation *Finalisation) {

	// Finalise blocks
	for _, block := range finalisation.blocks {

		// Finalise blocks
		beacon.chains[block.Shard].Finalise(block.Hash)
	}

	beacon.finalisation = *finalisation

}

func (beacon *Beacon) Println(a ...interface{}) {

	fmt.Printf("[beacon] ")
	fmt.Println(a...)

}
