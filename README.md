# Guaranteed-TX Simulator
The Guaranteed-TX simulator is a graphical interface written in Golang to simulate visualise the working of Guaranteed-TX, a  guaranteed cross-shard transaction execution protocol for Ethereum 2.0.

It was created as a graduation project to create a guaranteed cross-shard transaction execution protocol for [Ethereum 2.0](https://github.com/ethereum/eth2.0-specs).

## Compiling the Guaranteed-TX source code
The graphic interface provides some features to modify the simulation, however, many more variables can be modified in the source code. 

For building the Guaranteed-TX Simulator the source code needs to be available on the build system and Golang version >1.4+. Guaranteed-TX  uses Go bindings for nuklear.h — a small ANSI C gui library and requires a GNU Compiler Collection to build nuklear. Windows users can use MinGW. An extended installation description for nuklear can be found in the [Nuklear Go binding](https://github.com/golang-ui/nuklear) repository.

Subsequently, one can compile the code with `go build`.

## About Guaranteed-TX
Guaranteed-TX is a guaranteed cross-shard transaction execution protocol for Ethereum 2.0. Guaranteed-TX allows shards to process cross-shard transactions before being finalised in the block it was created - a property called optimistic execution - which significantly improves cross-shard transaction latencies. In addition, it provides economic guarantees that all cross-shard transaction will eventually be processed. In order to achieve both Guaranteed-TX intro- duces a messaging layer which records the created and processed cross-shard transactions and is shared with every shard. The messaging layer is used to finalise consistent blocks and punish valida- tors in a shard committee for not processing cross-shard transactions. Consequently, cross-shard transactions are either processed or slowly drain the stake of the validators within the addressed shard.

## Using the simulator
The simulator simulates an abstracted version of above protocol. For every shard the block headers are plotted as circles over time. The color of the circle indicates the status and the lines between circles a parent-child relation, with the parent always on earlier in time on the left side. Clicking on a circle shows the `txOut` and  `txIn`  transaction list of the related block header. Moreover, the beacon chain finalises blocks in the background.

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
 
  
A detailed description about the visualiser can be found in my thesis document.


IMPORTANT: The simulator is used to get insights in Guaranteed-TX and is not formally proven. 


## Gallery

![screenshot](https://raw.githubusercontent.com/sjoerdwels/Guaranteed-TX/master/assets/demo.gif)

![screenshot2](https://raw.githubusercontent.com/sjoerdwels/Guaranteed-TX/master/assets/demo2.png)
