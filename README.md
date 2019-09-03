# Smart-Shard  Visualiser
The  Smart-Shard Visualiser is a graphical interface written in Golang to visualise and simulate the reliable messaging system of Smart-Shard.
Smart-Shard is a reliable asynchronous messaging protocol that gives rise to low message overhead and computation costs. 
The reliable messaging protocol is a cryptoeconomic design in which delaying cross-shard transaction processing is punished. An additional advantage of the reliable messaging protocol is that it simplifies the cross-linking of shards.

It was created as a graduation project to visualise a reliable and asynchronous messaging protocol for [Ethereum 2.0](https://github.com/ethereum/eth2.0-specs).

## Compiling the Smart-Shard source code
The graphic interface provides some features to modify the simulation, however, many more variables can be modified in the source code. 

For building the Smart-Shard Visualiser the source code needs to be available on the build system and Golang version >1.4+. Smart-Shard uses Go bindings for nuklear.h — a small ANSI C gui library and requires a GNU Compiler Collection to build nuklear. Windows users can use MinGW. An extended installation description for nuklear can be found in the [Nuklear Go binding](https://github.com/golang-ui/nuklear) repository.

Subsequently, one can compile the code with `go build`.

## About Smart-Shard
Smart-shard exist of two parts:
1. A reliable messaging protocol for Ethereum 2.0.
2. A sharding scheme that enables fast cross-shard transactions and smart contract balancing.

The Smart-Shard visualiser only simulates part one. 

Part one works as follows; a validator ```v``` that is part of the validator committee of shard ```s``` acts as a full node for shard ```s``` and as a light client for all other shards.  With Smart-Shards protocol, two extra fields are added to each block header; a ```txIn``` field and a ```txOut``` field.  

The ```txOut``` field maintains the list of all generated receipts hashes for that specific block including a prefix of the destination shard. The ```txIn``` maintains  the list of processed receipts of that particular shard. A processed receipt therefore has a entry in the ```txOut``` list on the source shard and a ```txIn``` in the destination shard.

In addition, every beacon block that cross-links shard blocks at epoch checkpoints includes a list of all `inconsistent` cross-shard transactions, i.e. a receipt which is not processed at the target shard (alternatively an merkletree root of these transactions could be used). Each cross-link acts as a heartbeat. If a receipt is not processed within `x` heartbeats, both the target shard as source shard committee validators will be punished. This way, we economically garuantee reliable transactions. A finalisation, i.e. cross-linking block also allows to prune every blockheader of the non-validating shards. 

Moreover, instead of waiting epoch (6.4 minutes) before a receipt can be processed on the target shard, one can 
use the bock headers of each shard and an extended fork-choice rule of to invalidate and validate inconsistent blocks. This way, one can minimalise the waiting time to include cross-shard transactions without risking inconsistencies.


## Using the Visualiser
The visualiser simulates an abstracted version of above protocol. For every shard the block headers are plotted as circles over time. The color of the circle indicates the status and the lines between circles a parent-child relation, with the parent always on earlier in time on the left side. Clicking on a circle shows the `txOut` and  `txIn`  transaction list of the related block header. Moreover, the beacon chain finalised blocks in the background.

Status colors selected shard (full node):
* **Green** -> Finalised
* **Red** -> Invalid (block includes inconsistent txIn which is not part of the canonical chain of the sourceshard)
* **Yelllow** -> Canonical chain
* **Purple** -> Stale block
* **Blue** -> Genisis block

Status colors other shards (light client):
* **White** -> Pruned canonical chain
* **Yellow** - >Canonical chain
* **Blue** -> Last finalised block
 
  
A detailed description about the visualiser can be found in my thesis document: Smart-Shard.


IMPORTANT: The visualiser is used to get insights in the Smart-Shard protocol and is not formally proven. 


## Gallery

![screenshot](https://raw.githubusercontent.com/sjoerdwels/Smart-Shard/master/assets/demo.gif)

![screenshot2](https://raw.githubusercontent.com/sjoerdwels/Smart-Shard/master/assets/demo2.png)
