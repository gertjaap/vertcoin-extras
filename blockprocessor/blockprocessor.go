package blockprocessor

import (
	"fmt"
	"time"

	"github.com/btcsuite/btcd/chaincfg/chainhash"
	"github.com/btcsuite/btcd/rpcclient"
	"github.com/btcsuite/btcd/wire"
	"github.com/gertjaap/vertcoin-openassets/wallet"
)

type BlockProcessor struct {
	wallet    *wallet.Wallet
	rpcClient *rpcclient.Client
}

func NewBlockProcessor(w *wallet.Wallet, c *rpcclient.Client) *BlockProcessor {
	bp := new(BlockProcessor)
	bp.wallet = w
	bp.rpcClient = c
	return bp
}

func (bp *BlockProcessor) Loop() {
	client := bp.rpcClient
	lastHash := chainhash.Hash{}

	for {
		time.Sleep(time.Second)
		bestHash, err := client.GetBestBlockHash()
		if err != nil {
			fmt.Printf("Error getting best blockhash: %s\n", err.Error())
			continue
		}

		if bestHash.IsEqual(&lastHash) {
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

		fmt.Printf("Processed %d blocks, best hash %s\n", len(pendingBlockHashes), lastHash)

	}
}

func (bp *BlockProcessor) processBlock(block *wire.MsgBlock) error {
	for _, tx := range block.Transactions {
		bp.wallet.ProcessTransaction(tx)
	}

	return nil
}
