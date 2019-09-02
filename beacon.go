package main

import (
	"fmt"
	"math/rand"
	"time"
)

type Beacon struct {
	channels *Communication
}

func (beacon *Beacon) run() {

	fmt.Println("Launching beacon shard...")

	for {
		select {
			case block := <- beacon.channels.blocks[0]:
					beacon.receiveBlock(block)

			case <-time.After(FinalisationPeriod.NextRandomTimePeriod()):
			//	beacon.proposeFinalisation()
		}
	}
}

func (beacon *Beacon) receiveBlock(block *Block) {

	//fmt.Println("Beacon received block.")

}

func (beacon *Beacon) proposeFinalisation() {

	// Probability finalisation fails
	if rand.Float64() > FinalisationProbability {

		fmt.Println("BeaconChain could not finalise blocks.")
		return
	}

}