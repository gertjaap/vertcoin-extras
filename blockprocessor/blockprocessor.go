package blockprocessor

import (
	"fmt"
	"time"

	"github.com/btcsuite/btcd/chaincfg/chainhash"
	"github.com/btcsuite/btcd/rpcclient"
	"github.com/btcsuite/btcd/wire"
	"github.com/gertjaap/vertcoin-openassets/config"
	"github.com/gertjaap/vertcoin-openassets/util"
	"github.com/gertjaap/vertcoin-openassets/wallet"
)

type ChainIndex []*chainhash.Hash

type BlockProcessor struct {
	wallet      *wallet.Wallet
	config      *config.Config
	rpcClient   *rpcclient.Client
	synced      bool
	syncing     bool
	connected   bool
	headerQueue int
	syncHeight  int
	tipHeight   int
}

func NewBlockProcessor(w *wallet.Wallet, c *rpcclient.Client, conf *config.Config) *BlockProcessor {
	bp := new(BlockProcessor)
	bp.wallet = w
	bp.rpcClient = c
	bp.config = conf
	return bp
}

type SyncStatus struct {
	Connected   bool
	Synced      bool
	Syncing     bool
	SyncHeight  int
	TipHeight   int
	HeaderQueue int
}

func (bp *BlockProcessor) UpdateClient(c *rpcclient.Client) {
	bp.rpcClient = &(*c)
	bp.connected = true
}

func (bp *BlockProcessor) GetSyncStatus() SyncStatus {
	return SyncStatus{bp.connected, bp.synced, bp.syncing, bp.syncHeight, bp.tipHeight, bp.headerQueue}
}

func (activeChain ChainIndex) FindBlock(hash *chainhash.Hash) int {
	for i, b := range activeChain {
		if b.IsEqual(hash) {
			return i
		}
	}

	return -1
}

func (bp *BlockProcessor) Loop() {
	client := bp.rpcClient
	activeChain := ChainIndex{bp.config.Network.StartHash}
	bp.tipHeight = bp.config.Network.StartHeight - 1 + len(activeChain)

	bp.synced = false
	for {
		bp.syncing = false
		time.Sleep(time.Second)
		bp.syncing = true
		bp.syncHeight = bp.config.Network.StartHeight - 1 + len(activeChain)

		bestHash, err := client.GetBestBlockHash()
		if err != nil {
			if util.IsConnectionError(err) {
				bp.connected = false
			} else {
				fmt.Printf("Error getting best blockhash: %s\n", err.Error())
			}
			continue
		}
		bp.connected = true

		if bestHash.IsEqual(activeChain[len(activeChain)-1]) {
			continue
		}

		hash, _ := chainhash.NewHash(bestHash.CloneBytes())
		pendingBlockHashes := make([]*chainhash.Hash, 0)
		startIndex := 0
		for {
			header, err := client.GetBlockHeader(hash)
			if err != nil {
				fmt.Printf("Error getting block: %s\n", err.Error())
				continue
			}
			bp.headerQueue = len(pendingBlockHashes)
			newHash, _ := chainhash.NewHash(hash.CloneBytes())
			pendingBlockHashes = append([]*chainhash.Hash{newHash}, pendingBlockHashes...)
			hash = &header.PrevBlock
			idx := activeChain.FindBlock(&header.PrevBlock)
			fmt.Printf("Found %d blocks to process\n", len(pendingBlockHashes))
			if idx > -1 {
				fmt.Printf("Found prevBlock at index %d...\n", idx)
				// We found a way to connect to our activeChain
				// Remove all blocks after idx, if any
				newChain := activeChain[:idx]
				newChain = append(newChain, pendingBlockHashes...)
				activeChain = newChain
				startIndex = idx
				break
			}
		}

		if len(activeChain)-startIndex > 5 {
			bp.synced = false
		}

		bp.tipHeight = bp.config.Network.StartHeight - 1 + len(activeChain)

		for i, hash := range activeChain[startIndex:] {
			bp.headerQueue = len(activeChain) - i - 1
			block, err := client.GetBlock(hash)
			if err != nil {
				fmt.Printf("Error getting block: %s\n", err.Error())
				continue
			}

			err = bp.processBlock(block)
			if err != nil {
				fmt.Printf("Error processing block: %s\n", err.Error())
				continue
			}
		}

		bp.synced = true
	}
}

func (bp *BlockProcessor) processBlock(block *wire.MsgBlock) error {
	for _, tx := range block.Transactions {
		bp.wallet.ProcessTransaction(tx)
	}

	return nil
}
