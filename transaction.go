package main

import (
	"bytes"
	"crypto/sha256"
	"encoding/gob"
	"fmt"
)

type Transaction struct {
	SourceShard int
	TargetShard int
	Hash  string
	Data  string

}

// Calculate Hash of Block
func (tx *Transaction) SetHash() {

	tx.Hash = ""

	var encoded bytes.Buffer
	var hash [32]byte

	encoder := gob.NewEncoder(&encoded)
	encoder.Encode(tx)

	hash = sha256.Sum256(encoded.Bytes())
	tx.Hash = string(hash[:])

}

func (tx *Transaction) prettyPrint() {
	fmt.Printf("SourceShard:  %d - TargetShard: %d  -  Data: %s  \n", tx.SourceShard,  tx.TargetShard, tx.Data)
}


// Remove item from list and return
func ContainsTx(tx *Transaction, txListPointer *[]*Transaction) bool {

	for _, transaction := range *txListPointer {
		if tx.Hash == transaction.Hash {
			return true
		}
	}
	return false
}

// Remove item from list and return
func RemoveTxFromList(deleteTX *Transaction, txListPointer *[]*Transaction) bool {

	txList := *txListPointer

	for i, transaction := range txList {
		if deleteTX.Hash == transaction.Hash {
			txList[i] = txList[len(txList)-1]     // Copy last element to index i
			txList = txList[:len(txList)-1] // Truncate slice
			*txListPointer = txList
			return true
		}
	}

	return false
}