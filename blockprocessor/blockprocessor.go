package blockprocessor

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"time"

	"github.com/btcsuite/btcd/chaincfg/chainhash"
	"github.com/btcsuite/btcd/rpcclient"
	"github.com/btcsuite/btcd/wire"
	"github.com/gertjaap/vertcoin-extras/config"
	"github.com/gertjaap/vertcoin-extras/util"
	"github.com/gertjaap/vertcoin-extras/wallet"
)

type ChainIndex []*chainhash.Hash

type BlockProcessor struct {
	wallet      *wallet.Wallet
	config      *config.Config
	rpcClient   *rpcclient.Client
	activeChain ChainIndex
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
	bp.activeChain = ChainIndex{bp.config.Network.StartHash}
	bp.tipHeight = bp.config.Network.StartHeight - 1 + len(bp.activeChain)
	bp.ReadChainState()
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

func (bp *BlockProcessor) PersistChainState() {
	var buf bytes.Buffer
	for _, h := range bp.activeChain {
		buf.Write(h[:])
	}
	ioutil.WriteFile("chainstate.hex", buf.Bytes(), 0644)
}

func (bp *BlockProcessor) ReadChainState() {
	b, err := ioutil.ReadFile("chainstate.hex")
	if err != nil {
		return
	}
	readIndex := ChainIndex{}
	buf := bytes.NewBuffer(b)
	hash := make([]byte, 32)
	for {
		i, err := buf.Read(hash)
		if i == 32 && err == nil {
			ch, err := chainhash.NewHash(hash)
			if err == nil {
				readIndex = append(readIndex, ch)
			} else {
				break
			}
		} else {
			break
		}
	}
	bp.activeChain = readIndex
}

func (bp *BlockProcessor) UpdateClient(c *rpcclient.Client) {
	bp.rpcClient = c
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
	bp.synced = false
	for {
		bp.syncing = false
		time.Sleep(time.Second)
		bp.syncing = true
		bp.syncHeight = bp.config.Network.StartHeight - 1 + len(bp.activeChain)

		bestHash, err := bp.rpcClient.GetBestBlockHash()
		if err != nil {
			if util.IsConnectionError(err) {
				bp.connected = false
			} else {
				fmt.Printf("Error getting best blockhash: %s\n", err.Error())
			}
			continue
		}
		bp.connected = true

		if bestHash.IsEqual(bp.activeChain[len(bp.activeChain)-1]) {
			bp.synced = true
			continue
		}

		hash, _ := chainhash.NewHash(bestHash.CloneBytes())
		pendingBlockHashes := make([]*chainhash.Hash, 0)
		startIndex := 0
		for {
			header, err := bp.rpcClient.GetBlockHeader(hash)
			if err != nil {
				fmt.Printf("Error getting block: %s\n", err.Error())
				continue
			}

			bp.headerQueue = len(pendingBlockHashes)
			newHash, _ := chainhash.NewHash(hash.CloneBytes())
			pendingBlockHashes = append([]*chainhash.Hash{newHash}, pendingBlockHashes...)
			hash = &header.PrevBlock
			idx := bp.activeChain.FindBlock(&header.PrevBlock)
			if idx > -1 {
				// We found a way to connect to our activeChain
				// Remove all blocks after idx, if any
				newChain := bp.activeChain[:idx]
				newChain = append(newChain, pendingBlockHashes...)
				bp.activeChain = newChain
				startIndex = idx
				break
			}
		}

		if len(bp.activeChain)-startIndex > 5 {
			bp.synced = false
		}

		bp.tipHeight = bp.config.Network.StartHeight - 1 + len(bp.activeChain)

		for i, hash := range bp.activeChain[startIndex:] {
			bp.headerQueue = len(bp.activeChain) - i - 1
			block, err := bp.rpcClient.GetBlock(hash)
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
		bp.PersistChainState()

	}
}

func (bp *BlockProcessor) processBlock(block *wire.MsgBlock) error {
	for _, tx := range block.Transactions {
		bp.wallet.ProcessTransaction(tx)
	}

	return nil
}
