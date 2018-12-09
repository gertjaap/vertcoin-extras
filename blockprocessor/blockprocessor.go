package blockprocessor

import (
	"fmt"
	"time"

	"github.com/btcsuite/btcd/chaincfg/chainhash"
	"github.com/btcsuite/btcd/rpcclient"
	"github.com/btcsuite/btcd/wire"
	"github.com/gertjaap/vertcoin-openassets/config"
	"github.com/gertjaap/vertcoin-openassets/wallet"
)

type BlockProcessor struct {
	wallet    *wallet.Wallet
	config    *config.Config
	rpcClient *rpcclient.Client
}

func NewBlockProcessor(w *wallet.Wallet, c *rpcclient.Client, conf *config.Config) *BlockProcessor {
	bp := new(BlockProcessor)
	bp.wallet = w
	bp.rpcClient = c
	bp.config = conf
	return bp
}

func (bp *BlockProcessor) Loop() {
	client := bp.rpcClient
	lastHash := bp.config.Network.StartHash

	for {
		time.Sleep(time.Second)
		bestHash, err := client.GetBestBlockHash()
		if err != nil {
			fmt.Printf("Error getting best blockhash: %s\n", err.Error())
			continue
		}

		if bestHash.IsEqual(lastHash) {
			continue
		}

		hash := chainhash.Hash{}
		hash.SetBytes(bestHash.CloneBytes())

		pendingBlockHashes := make([]chainhash.Hash, 0)

		for {
			header, err := client.GetBlockHeader(&hash)
			if err != nil {
				fmt.Printf("Error getting block: %s\n", err.Error())
				continue
			}

			pendingBlockHashes = append([]chainhash.Hash{hash}, pendingBlockHashes...)

			hash = header.PrevBlock
			if lastHash.IsEqual(&header.PrevBlock) {
				break
			}
		}

		for _, hash := range pendingBlockHashes {
			block, err := client.GetBlock(&hash)
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

		lastHash.SetBytes(bestHash.CloneBytes())

	
	}
}

func (bp *BlockProcessor) processBlock(block *wire.MsgBlock) error {
	for _, tx := range block.Transactions {
		bp.wallet.ProcessTransaction(tx)
	}

	return nil
}
