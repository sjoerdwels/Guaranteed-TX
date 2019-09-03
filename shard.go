package main

import (
	"fmt"
	"math/rand"
	"time"
)

type Shard struct {
	id           int
	channels     Communication
	txOutPool    []*Transaction
	chains       []Chain
	finalisation Finalisation
}

func (shard *Shard) init() {

	// Init chains
	shard.chains = make([]Chain, ShardCount+1)
	for i, _ := range shard.chains {
		shard.chains[i] = Chain{}
		shard.chains[i].init(i)
	}

	// Init txPool
	shard.txOutPool = make([]*Transaction, 0)

	// Init finalisation
	shard.finalisation = Finalisation{
		height:         0,
		blocks:         nil,
		inconsistentTX: make([]*Transaction, 0),
	}

}

func (shard *Shard) run() {

	shard.Println("Launching shard")

	state := Pause

	for {

		select {
		case command := <-shard.channels.control[shard.id]:
			shard.Println("Received command:", command)
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
			case command := <-shard.channels.control[shard.id]:
				shard.Println("Received command:", command)
				switch *command {
				case Run:
					state = Run
				case Pause:
					state = Pause
					break;
				case Exit:
					return
				}
			case block := <-shard.channels.blocks[shard.id]:
				shard.receiveBlock(block)

			case finalisation := <-shard.channels.finalisation[shard.id]:
				shard.receiveFinalisation(finalisation)

			case <-time.After(BlockGenerationPeriod.NextRandomTimePeriod()):
				shard.generateBlock()

			case <-time.After(TXGenerationPeriod.NextRandomTimePeriod()):
				shard.generateTransactions()
			}
		}
	}
}

func (shard *Shard) receiveBlock(block *Block) {

	shard.Println(fmt.Sprintf("Received block from  %d.",  block.Shard))

	// Add block
	shard.chains[block.Shard].Insert(block)

	// Update shard block tree
	shard.updateBlockTree()

}

func (shard *Shard) receiveFinalisation(finalisation *Finalisation) {

	// Finalise blocks
	for _, block := range finalisation.blocks {

		// Remove finalised transactions from TX Out Pool
		if shard.id == block.Shard {
			finalisedTXOutList := shard.chains[shard.id].GetTXOutList(block.Hash)

			for _, finalisedTX := range finalisedTXOutList {
				RemoveTxFromList(finalisedTX, &shard.txOutPool)
			}
		}

		// Finalise blocks
		shard.chains[block.Shard].Finalise(block.Hash)

		// Prune other shards
		if shard.id != block.Shard {
			//shard.chains[block.Shard].Prune(block.Hash)
		}
	}

	shard.finalisation = *finalisation

	// Update BlockTree
	shard.updateBlockTree()
}

func (shard *Shard) generateBlock() {

	// Probability finalisation fails
	if rand.Float64() > BlockGenerationProbability {

		shard.Println("Skip block generation.")
		return
	}

	// Get candidate chains to build block upon
	candidateParents := shard.chains[shard.id].GetLongestChains(3, true)

	// Random 'select the longest chain' to simulate forks
	parentChain := candidateParents[0]
	if rand.Float64() > ProbabilityBuildOnLongestChain {
		if rand.Float64() <= 0.5 {
			parentChain = candidateParents[1]
		} else {
			parentChain = candidateParents[2]
		}
	}

	// Include some IN transactions
	txOutOthers := shard.getOtherShardsTxOutList()
	processedTxIn := shard.chains[shard.id].GetTXInList(parentChain.block.Hash)

	for _, txOut := range processedTxIn {
		RemoveTxFromList(txOut, &txOutOthers)
	}

	numberOfTxIn := MinOf(len(txOutOthers), BlockTxOutNumber.NextRandomInt())

	txInList := make([]*Transaction, numberOfTxIn)

	j := 0
	for _, i := range rand.Perm(len(txOutOthers)) {
		if j == numberOfTxIn {
			break;
		}
		txInList[j] = txOutOthers[i]
		j++
	}

	// Include some OUT transactions
	// deep-copy list
	availableTxOut := make([]*Transaction, len( shard.txOutPool))
	copy(availableTxOut,  shard.txOutPool)
	processedTxOut := shard.chains[shard.id].GetTXOutList(parentChain.block.Hash)

	for _, txOut := range processedTxOut {
		RemoveTxFromList(txOut, &availableTxOut)
	}

	numberOfTxOut := MinOf(len(availableTxOut), BlockTxOutNumber.NextRandomInt())

	txOutList := make([]*Transaction, numberOfTxOut)
	j = 0
	for _, i := range rand.Perm(len(availableTxOut)) {
		if j == numberOfTxOut {
			break;
		}
		txOutList[j] = availableTxOut[i]
		j++
	}

	// Publish block
	block := Block{
		Shard:      shard.id,
		Hash:       []byte{},
		ParentHash: parentChain.block.Hash,
		TXIn:       txInList,
		TXOut:      txOutList,
		Validator:  time.Now().String(),
	}
	block.SetHash()

	shard.channels.broadcastBlock(block)

}

// Update valid block tree, based on 'Valid block trees'
// 'Magic Fork Choice Rule'
func (shard *Shard) updateBlockTree() {

	txOutList := shard.getOtherShardsTxOutList()

	// Update consistency of shard chain.
	shard.chains[shard.id].UpdateConsistency(txOutList)
}

func (shard *Shard) getOtherShardsTxOutList() []*Transaction {

	txOutList := make([]*Transaction, 0)

	// Get all finalised inconsistent transactions
	for _, tx := range shard.finalisation.inconsistentTX {
		if tx.TargetShard == shard.id {
			txOutList = append(txOutList, tx)
		}
	}

	// For all other shards, get all outgoing TX related to this shard.
	for i := 1; i <= ShardCount; i++ {

		if i != shard.id {

			shardTxOut := shard.chains[i].GetLongestChainTXOutList()

			for _, tx := range shardTxOut {
				if tx.TargetShard == shard.id {
					txOutList = append(txOutList, tx)
				}
			}
		}

	}

	return txOutList
}

func (shard *Shard) generateTransactions() {

	// Create transactions
	for i := 1; i <= TXGenerationNumber.NextRandomInt(); i++ {

		if len(shard.txOutPool) < TXPoolSize {

			// Random destination
			destShard := ShardRange.NextRandomInt()
			for destShard == shard.id {
				destShard = ShardRange.NextRandomInt()
			}

			tx := Transaction{
				SourceShard: shard.id,
				TargetShard: destShard,
				Data:  time.Now().String(),
			}

			tx.SetHash()

			shard.txOutPool = append(shard.txOutPool, &tx)

			shard.Println("Generated new transactions, txOutPool size:", len(shard.txOutPool))
		}
	}

}

func (shard *Shard) UpdateVisualisation() {

	for i := 1; i <= ShardCount; i++ {
		shard.chains[i].UpdateVisualisation(shard.id == i)
	}
}

func (shard *Shard) Println(a ...interface{}) {
	if shard.id == debugShard {
		fmt.Printf("[shard %d] ", shard.id)
		fmt.Println(a...)
	}
}
