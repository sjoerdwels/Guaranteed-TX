package main

import (
	"encoding/base64"
	"fmt"
	"github.com/fatih/color"
	"reflect"
	"strconv"
	"strings"
	"time"
)

type Chain struct {
	genesisBlock       *ChainBlock
	lastFinalisedBlock *ChainBlock
}

func (chain *Chain) init(shard int) {

	block := Block{
		Shard:      shard,
		Hash:       []byte(fmt.Sprintf("Shard %d", shard)),
		ParentHash: nil,
		TXIn:       nil,
		TXOut:      nil,
	}

	coordinate := Coordinate{
		x:     0,
		y:     0,
		color: cGENISIS,
	}

	chain.genesisBlock = &ChainBlock{
		height:     0,
		block:      &block,
		parent:     nil,
		children:   nil,
		valid:      true,
		finalised:  true,
		coordinate: coordinate,
	}

	chain.lastFinalisedBlock = chain.genesisBlock
}

// Prune chain up to block if exists
func (chain *Chain) Prune(hash []byte) {

	fmt.Println("Prune blocks.")

	chainBlock := chain.Search(hash);
	if chainBlock != nil {
		chainBlock.parent = nil
		chain.genesisBlock = chainBlock
	}

}

// Insert block if parent exists
func (chain *Chain) Insert(block *Block) {

	parent := chain.Search(block.ParentHash)

	// Calculate X coordinate
	t := time.Now()
	duration := t.Sub(StartTime)
	x := duration.Seconds() * pixelsPerSecond

	// Calculate Y coordinate
	y := parent.coordinate.y
	if chain.BlockInLongestChain(parent) {
		y = 0
	}
	nrOfChild := float32(len(parent.children))
	if nrOfChild > 0 {
		y +=  nrOfChild*20
	}

	coordinate := Coordinate{
		x:     float32(x),
		y:     y,
		color: cSTALE,
	}

	if parent != nil {

		chainBlock := ChainBlock{
			height:     parent.height + 1,
			block:      block,
			parent:     parent,
			children:   []*ChainBlock{},
			valid:      true,
			finalised:  false,
			coordinate: coordinate,
		}

		parent.children = append(parent.children, &chainBlock)
	}
}

// Finalise blocks
func (chain *Chain) Finalise(hash []byte) {

	chainBlock := chain.Search(hash)

	if chainBlock != nil {
		chain.finalise(chainBlock)
		chain.lastFinalisedBlock = chainBlock
	}
}

func (chain *Chain) finalise(chainBlock *ChainBlock) {
	chainBlock.finalised = true
	if chainBlock.parent != nil {
		chain.finalise(chainBlock.parent)
	}
}

// Search block or return nil
func (chain *Chain) Search(hash []byte) *ChainBlock {
	return search(chain.genesisBlock, hash)
}

func search(block *ChainBlock, hash []byte) *ChainBlock {

	if reflect.DeepEqual(block.block.Hash, hash) {
		return block
	}

	for _, child := range block.children {
		childSearch := search(child, hash)

		if childSearch != nil {
			return childSearch
		}
	}

	return nil
}

// Get TX Out list since last finalised block, uptil lastChainBLock.
func (chain *Chain) GetTXOutList(lastBlockHash []byte) []*Transaction {
	chainBlock := search(chain.lastFinalisedBlock, lastBlockHash)
	return chain.getTXOutList(chainBlock)
}

// Get TX Out list since last finalised block, uptil lastChainBLock.
func (chain *Chain) getTXOutList(lastChainBlock *ChainBlock) []*Transaction {

	txOutList := make([]*Transaction, 0)
	if !lastChainBlock.finalised {
		txOutList = append(lastChainBlock.block.TXOut, chain.getTXOutList(lastChainBlock.parent)...)
	}
	return txOutList
}

// Get TX Out list of longest chain since last finalised block.
func (chain *Chain) GetLongestChainTXOutList() []*Transaction {
	longestChains := chain.GetLongestChains(1, true)
	return chain.GetTXOutList(longestChains[0].block.Hash)
}

// Get TX Out list since last finalised block, uptil lastChainBLock.
func (chain *Chain) GetTXInList(lastBlockHash []byte) []*Transaction {
	chainBlock := search(chain.lastFinalisedBlock, lastBlockHash)
	return chain.getTXInList(chainBlock)
}

// Get TX Out list since last finalised block, uptil lastChainBLock.
func (chain *Chain) getTXInList(lastChainBlock *ChainBlock) []*Transaction {

	txInList := make([]*Transaction, 0)
	if !lastChainBlock.finalised {
		txInList = append(lastChainBlock.block.TXIn, chain.getTXInList(lastChainBlock.parent)...)
	}
	return txInList
}

// Get longest three chains
func (chain *Chain) GetLongestChains(numberOfChains int, validOnly bool) []*ChainBlock {
	return chain.lastFinalisedBlock.getLongestChains(numberOfChains, validOnly)
}

// Get longest chains (array of numberOfChain blocks)
func (chainBlock *ChainBlock) getLongestChains(numberOfChains int, validOnly bool) []*ChainBlock {

	longestChains := make([]*ChainBlock, numberOfChains)

	for i := 0; i < numberOfChains; i++ {
		longestChains[i] = chainBlock;
	}

	for _, child := range chainBlock.children {

		childChains := child.getLongestChains(numberOfChains, validOnly)

		for i := 0; i < numberOfChains; i++ {
			if longestChains[i].height < childChains[0].height {
				if !validOnly || childChains[0].valid {
					longestChains[i] = childChains[0]
					break;
				}
			}
		}

		for j := 1; j < numberOfChains; j++ {
			if !reflect.DeepEqual(childChains[j].block.Hash, childChains[j-1].block.Hash) {
				for i := 0; i < numberOfChains; i++ {
					if longestChains[i].height < childChains[j].height {
						if !validOnly || childChains[0].valid {
							longestChains[i] = childChains[j]
							break;
						}
					}
				}
			}
		}
	}
	return longestChains
}

func (chain *Chain) UpdateConsistency(txOutList []*Transaction) {
	// paste-by-value list
	copyList := make([]*Transaction, len(txOutList))
	copy(copyList, txOutList)
	chain.lastFinalisedBlock.updateConsistency(copyList)
}

func (chain *Chain) PrettyPrint() {
	chain.prettyPrint("", chain.genesisBlock, true)
}

// Retrieve whether block is part of longest chain.
func (chain *Chain) BlockInLongestChain(chainBlock *ChainBlock) bool {
	longestChains := chain.GetLongestChains(1, true)
	return blockInLongestChain(chainBlock, longestChains[0])
}

// Recursive search if block is part of chain.
func blockInLongestChain(chainBlock *ChainBlock, longestChainBlock *ChainBlock) bool {

	if reflect.DeepEqual(chainBlock.block.Hash, longestChainBlock.block.Hash) {
		return true
	} else if longestChainBlock.parent == nil {
		return false
	} else {
		return blockInLongestChain(chainBlock, longestChainBlock.parent)
	}
}

func (chain *Chain) UpdateVisualisation(isShardChain bool) {
	chain.updateVisualisation(isShardChain, chain.genesisBlock)
}

func (chain *Chain) updateVisualisation(isShardChain bool, chainBlock *ChainBlock) {

	if isShardChain {

		if reflect.DeepEqual(chainBlock.block.Hash, chain.genesisBlock.block.Hash) {
			chainBlock.coordinate.color = cGENISIS
		} else if chainBlock.finalised {
			chainBlock.coordinate.color = cFINALISED
		} else if !chainBlock.valid {
			chainBlock.coordinate.color = cINVALID
		} else if chain.BlockInLongestChain(chainBlock) {
			chainBlock.coordinate.color = cCANONICAL
		} else {
			chainBlock.coordinate.color = cSTALE
		}

	} else {

		if reflect.DeepEqual(chainBlock.block.Hash, chain.lastFinalisedBlock.block.Hash) {
			chainBlock.coordinate.color = cGENISIS
		} else if chainBlock.finalised {
			chainBlock.coordinate.color = cFINALISEDOTHER
		} else if chain.BlockInLongestChain(chainBlock) {
			chainBlock.coordinate.color = cCANONICAL
		} else {
			chainBlock.coordinate.color = cPRUNED
		}
	}

	for _, child := range chainBlock.children {
		chain.updateVisualisation(isShardChain,child)
	}
}

func (chain *Chain) prettyPrint(prefix string, chainBlock *ChainBlock, lastChild bool) {

	fmt.Printf(prefix)

	if lastChild {
		fmt.Print("└──")
	} else {
		fmt.Print("├──")
	}

	output := "(" + base64.URLEncoding.EncodeToString(chainBlock.block.Hash) + ")"
	for _, txIn := range chainBlock.block.TXIn {
		output = output + " TXin: " + strconv.Itoa(txIn.TargetShard) + ","
	}
	for _, txOut := range chainBlock.block.TXOut {
		output = output + " TXout: " + strconv.Itoa(txOut.TargetShard) + ","
	}

	// Print colored block
	if chainBlock.finalised {
		color.Green(output)
	} else if !chainBlock.valid {
		color.Red(output)
	} else if chain.BlockInLongestChain(chainBlock) {
		color.Yellow(output)
	} else {
		color.White(output)
	}

	for i, child := range chainBlock.children {

		addPrefix := "│   "
		if lastChild {
			addPrefix = "    "
		}

		if i == len(chainBlock.children)-1 {
			chain.prettyPrint(prefix+addPrefix, child, true);
		} else {
			chain.prettyPrint(prefix+addPrefix, child, false);
		}
	}
}

func (chain *Chain) GetChainBlock(txOUt *Transaction)[]*ChainBlock {
	return chain.getChainBlock(txOUt,chain.genesisBlock)
}

func (chain *Chain) getChainBlock(txOUt *Transaction, chainBlock *ChainBlock) []*ChainBlock{

	blocks := []*ChainBlock{}

	for _, tx := range chainBlock.block.TXOut {
		if tx.Hash == txOUt.Hash  {
			blocks = append(blocks, chainBlock)
		}
	}

	for _, child := range chainBlock.children{
		blocks = append(blocks, chain.getChainBlock(txOUt,child)...)
	}

	return blocks
}


func rightPad2Len(s string, padStr string, overallLen int) string {
	var padCountInt int
	padCountInt = 1 + ((overallLen - len(padStr)) / len(padStr))
	var retStr = s + strings.Repeat(padStr, padCountInt)
	return retStr[:overallLen]
}
