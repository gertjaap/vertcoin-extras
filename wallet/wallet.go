package wallet

import (
	"bytes"
	"crypto/rand"
	"fmt"
	"io/ioutil"
	"os"

	"github.com/btcsuite/btcd/btcec"
	"github.com/btcsuite/btcd/rpcclient"
	"github.com/btcsuite/btcd/wire"
	"github.com/btcsuite/btcutil"
	"github.com/gertjaap/vertcoin-openassets/config"
	"github.com/gertjaap/vertcoin-openassets/util"
	"github.com/mit-dci/lit/bech32"
)

type Wallet struct {
	rpcClient  *rpcclient.Client
	config     *config.Config
	utxos      []Utxo
	assetUtxos []OpenAssetUtxo
	assets     []*OpenAsset
	privateKey *btcec.PrivateKey
	pubKey     *btcec.PublicKey
	pubKeyHash [20]byte
}

func NewWallet(c *rpcclient.Client, conf *config.Config) *Wallet {
	w := new(Wallet)
	w.rpcClient = c
	w.config = conf
	return w
}

func (w *Wallet) UpdateClient(c *rpcclient.Client) {
	w.rpcClient = &(*c)
}

func (w *Wallet) InitKey() error {
	var err error
	var privKey [32]byte
	if _, err := os.Stat("privkey.hex"); os.IsNotExist(err) {
		rand.Read(privKey[:])
		err := ioutil.WriteFile("privkey.hex", privKey[:], 0600)
		if err != nil {
			return err
		}
	} else {
		bytes, err := ioutil.ReadFile("privkey.hex")
		if err != nil {
			return err
		}
		copy(privKey[:], bytes[:])
	}

	w.privateKey, w.pubKey = btcec.PrivKeyFromBytes(btcec.S256(), privKey[:])
	copy(w.pubKeyHash[:], btcutil.Hash160(w.pubKey.SerializeCompressed()))
	return err
}

func (w *Wallet) VertcoinAddress() (string, error) {
	return bech32.SegWitV0Encode(w.config.Network.VtcAddressPrefix, w.pubKeyHash[:])
}

func (w *Wallet) AssetsAddress() (string, error) {
	return bech32.SegWitV0Encode(w.config.Network.AssetAddressPrefix, w.pubKeyHash[:])
}

func (w *Wallet) Balance() uint64 {
	value := uint64(0)
	for _, u := range w.utxos {
		value += u.Value
	}
	return value
}

func (w *Wallet) MyPKH() [20]byte {
	return w.pubKeyHash
}

func (w *Wallet) AssetBalance(assetID []byte) uint64 {
	value := uint64(0)
	for _, au := range w.assetUtxos {
		if bytes.Equal(au.AssetID, assetID) && au.Ours {
			value += au.AssetValue
		}
	}
	return value
}

func (w *Wallet) Assets() []*OpenAsset {
	return w.assets
}

func (w *Wallet) FollowAsset(assetID []byte) {
	for _, a := range w.assets {
		if bytes.Equal(a.AssetID, assetID) {
			a.Follow = true
		}
	}
}

func (w *Wallet) UnfollowAsset(assetID []byte) {
	for _, a := range w.assets {
		if bytes.Equal(a.AssetID, assetID) {
			a.Follow = true
		}
	}
}

func (w *Wallet) ProcessTransaction(tx *wire.MsgTx) {
	if IsOpenAssetTransaction(tx) {
		w.processOpenAssetTransaction(tx)
	} else {
		w.processNormalTransaction(tx)
	}
}

func (w *Wallet) processNormalTransaction(tx *wire.MsgTx) {
	for i, out := range tx.TxOut {
		keyHash := util.KeyHashFromPkScript(out.PkScript)
		if bytes.Equal(keyHash, w.pubKeyHash[:]) {
			w.registerUtxo(Utxo{
				TxHash:   tx.TxHash(),
				Outpoint: uint32(i),
				Value:    uint64(out.Value),
				PkScript: out.PkScript,
			})
		}
	}

	w.markTxInputsAsSpent(tx)
}

func (w *Wallet) markTxInputsAsSpent(tx *wire.MsgTx) {
	for _, in := range tx.TxIn {
		removeIndex := -1
		for j, out := range w.utxos {
			if in.PreviousOutPoint.Hash.IsEqual(&out.TxHash) && in.PreviousOutPoint.Index == out.Outpoint {
				// Spent!
				removeIndex = j
				break
			}
		}
		if removeIndex >= 0 {
			w.utxos = append(w.utxos[:removeIndex], w.utxos[removeIndex+1:]...)
		}
	}
}

func (w *Wallet) registerUtxo(utxo Utxo) {
	w.utxos = append(w.utxos, utxo)
}

func (w *Wallet) registerAssetUtxo(utxo OpenAssetUtxo) {
	w.assetUtxos = append(w.assetUtxos, utxo)
}

func (w *Wallet) registerAsset(asset OpenAsset) {
	w.assets = append(w.assets, &asset)
}

func (w *Wallet) FindUtxoFromTxIn(txi *wire.TxIn) (Utxo, error) {
	for _, out := range w.utxos {
		if txi.PreviousOutPoint.Hash.IsEqual(&out.TxHash) && txi.PreviousOutPoint.Index == out.Outpoint {
			return out, nil
		}
	}
	for _, aout := range w.assetUtxos {
		if txi.PreviousOutPoint.Hash.IsEqual(&aout.Utxo.TxHash) && txi.PreviousOutPoint.Index == aout.Utxo.Outpoint {
			return aout.Utxo, nil
		}
	}
	return Utxo{}, fmt.Errorf("Utxo not found")
}

func (w *Wallet) AddInputsAndChange(tx *wire.MsgTx, totalValueNeeded uint64) error {
	valueAdded := uint64(0)
	utxosToAdd := []Utxo{}
	for _, utxo := range w.utxos {
		utxosToAdd = append(utxosToAdd, utxo)
		valueAdded += utxo.Value
		if valueAdded > totalValueNeeded {
			break
		}

	}
	if valueAdded < totalValueNeeded {
		return fmt.Errorf("Insufficient balance")
	}

	for _, utxo := range utxosToAdd {
		tx.AddTxIn(wire.NewTxIn(&wire.OutPoint{utxo.TxHash, utxo.Outpoint}, nil, nil))
	}

	// Add change output when there's more than dust left, otherwise give to miners
	if valueAdded-totalValueNeeded > MINOUTPUT {
		tx.AddTxOut(wire.NewTxOut(int64(valueAdded-totalValueNeeded), util.DirectWPKHScriptFromPKH(w.MyPKH())))
	}

	return nil
}
