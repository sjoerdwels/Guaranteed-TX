package main

type Finalisation struct {
	height         int
	blocks         []Block
	inconsistentTX []*Transaction
}

type ShardFinalisation struct {
	shard               int
	newFinalisedBlock   *ChainBlock
	canonicalChainBlock *ChainBlock
	TXIn                *[]*Transaction
	TXOut               *[]*Transaction
}

func tryFinaliseNextBlock(shardIndex int,  finalisations *[]ShardFinalisation) bool {

	finalisation := (*finalisations)[shardIndex]

	// Verify is next block is valid
	nextBlock := finalisation.getNextBlock()

	// Return false is there is no next block
	if nextBlock == finalisation.newFinalisedBlock {
		return false
	}

	// Verify if TXin are valid
	for _, txIn := range nextBlock.block.TXIn {
		if !ContainsTx(txIn, (*finalisations)[txIn.SourceShard-1].TXOut) {
			return false
		}
	}

	// Block is valid, update shardFinalisation
	finalisation.newFinalisedBlock = nextBlock

	// Remove TXin from other finalisations
	for _, txIn := range nextBlock.block.TXIn {
		RemoveTxFromList(txIn, (*finalisations)[txIn.SourceShard-1].TXOut)
	}

	// Add txOUT to TX out list
	txOutList := *finalisation.TXOut
	for _, txOut := range nextBlock.block.TXOut {
		txOutList = append(txOutList, txOut)
	}
	finalisation.TXOut = &txOutList

	(*finalisations)[shardIndex] = finalisation

	return true
}

// Reverse order find next canonical chain block
func (finalisation *ShardFinalisation) getNextBlock() *ChainBlock {

	nextBlock := finalisation.canonicalChainBlock

	for nextBlock != finalisation.newFinalisedBlock {

		if nextBlock.parent == finalisation.newFinalisedBlock {
			break;
		}
		nextBlock = nextBlock.parent
	}
	return nextBlock
}
