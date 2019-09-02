package main

import (
	"bytes"
	"crypto/sha256"
	"encoding/gob"
)

type Block struct {
	Shard      int
	Hash       []byte
	ParentHash []byte
	TXIn       []*Transaction
	TXOut      []*Transaction
	Validator  string
}

type ChainBlock struct {
	height     int
	block      *Block
	parent     *ChainBlock
	children   []*ChainBlock
	valid      bool
	finalised  bool
	coordinate Coordinate
}

// Calculate Hash of Block
func (block *Block) SetHash() {

	block.Hash = []byte{}

	var encoded bytes.Buffer

	encoder := gob.NewEncoder(&encoded)
	encoder.Encode(block)

	hash := sha256.Sum256(encoded.Bytes())
	block.Hash = hash[:]
}

// Update inconsistency of chain based on txOut of other chains.
func (chainBlock *ChainBlock) updateConsistency(txOutList []*Transaction) {

	// Re-validate block.
	chainBlock.valid = true

	// Validate consistent txIn
	for _, txIn := range chainBlock.block.TXIn {

		if !RemoveTxFromList(txIn, &txOutList) {
			chainBlock.invalidateBlock()
			return
		}
	}

	// Validate children
	for _, child := range chainBlock.children {
		// Pass-by-value -  every block has different children.
		copyList := make([]*Transaction, len(txOutList))
		copy(copyList, txOutList)
		child.updateConsistency(copyList)
	}
}

func (chainBlock *ChainBlock) invalidateBlock() {
	chainBlock.valid = false
	for _, child := range chainBlock.children {
		child.invalidateBlock()
	}
}
