package main

type Communication struct {
	blocks       [ShardCount + 1]chan *Block
	finalisation [ShardCount + 1]chan *Finalisation
	control      [ShardCount + 1]chan *Command
}

// Broadcast block to all shards except source shard
func (communication *Communication) broadcastBlock(block Block) {
	for _, channel := range communication.blocks {
		channel <- &block
	}
}

// Broadcast finalisation to all shard and beacon shard
func (communication *Communication) broadcastFinalisation(finalisation *Finalisation) {
	for _, channel := range communication.finalisation {
		channel <- finalisation
	}
}

// Broadcast finalisation to all shard and beacon shard
func (communication *Communication) broadCastCommand(command Command) {
	for _, channel := range communication.control {
		channel <- &command
	}
}
