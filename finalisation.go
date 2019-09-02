package main

type Finalisation struct {
	height	int
	blocks  []Block
	inconsistentTX []*Transaction
}
