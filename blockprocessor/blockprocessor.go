package blockprocessor

import (
	"fmt"
	"log"
	"time"

	"github.com/btcsuite/btcd/chaincfg/chainhash"
	"github.com/btcsuite/btcd/rpcclient"
	"github.com/btcsuite/btcd/wire"
	"github.com/gertjaap/vertcoin-openassets/openassets"
)

func Loop() {
	lastHash := chainhash.Hash{}
	connCfg := &rpcclient.ConnConfig{
		Host:         "localhost:18443",
		User:         "vtc",
		Pass:         "vtc",
		HTTPPostMode: true, // Bitcoin core only supports HTTP POST mode
		DisableTLS:   true, // Bitcoin core does not provide TLS by default
	}
	client, err := rpcclient.New(connCfg, nil)
	if err != nil {
		log.Fatal(err)
	}
	defer client.Shutdown()

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

			err = processBlock(block)
			if err != nil {
				fmt.Printf("Error processing block: %s\n", err.Error())
				continue
			}
		}

		lastHash.SetBytes(bestHash.CloneBytes())
	}
}

func processBlock(block *wire.MsgBlock) error {
	for _, tx := range block.Transactions {
		if openassets.IsOpenAssetTransaction(tx) {
			fmt.Printf("Found open asset transaction\n")
		}
	}

	fmt.Printf("Processed block %s", block.BlockHash())
	return nil
}
