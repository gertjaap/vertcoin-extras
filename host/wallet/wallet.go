package wallet

import (
	"bytes"
	"fmt"
	"os"
	"path"

	"github.com/gertjaap/vertcoin/host/coinparams"

	"github.com/btcsuite/btcd/rpcclient"
	"github.com/btcsuite/btcd/wire"
	"github.com/gertjaap/vertcoin/logging"
	"github.com/gertjaap/vertcoin/util"
	"github.com/tidwall/buntdb"
)

type Wallet struct {
	rpcClient    *rpcclient.Client
	coinNetwork  coinparams.CoinNetwork
	coin         coinparams.Coin
	key          *Key
	utxos        []Utxo
	stealthUtxos []StealthUtxo
	assetUtxos   []OpenAssetUtxo
	assets       []*OpenAsset
	db           *buntdb.DB
}

func NewWallet(c *rpcclient.Client, coinNetwork coinparams.CoinNetwork, coin coinparams.Coin, key *Key) (*Wallet, error) {
	var err error
	w := new(Wallet)
	w.rpcClient = c
	w.coinNetwork = coinNetwork
	w.coin = coin
	w.key = key
	os.MkdirAll(path.Join(util.DataDirectory(), coin.Id, coinNetwork.Id), 0700)
	w.db, err = buntdb.Open(path.Join(util.DataDirectory(), coin.Id, coinNetwork.Id, "wallet.db"))
	if err != nil {
		return nil, err
	}
	w.loadStuff()
	return w, nil
}

func (w *Wallet) loadStuff() {
	w.db.View(func(tx *buntdb.Tx) error {
		tx.AscendRange("", "utxo-", "utxp-", func(key, value string) bool {
			logging.Debugf("Loading key %s\n", key)
			w.utxos = append(w.utxos, UtxoFromBytes([]byte(value)))
			return true
		})
		return nil
	})

	w.db.View(func(tx *buntdb.Tx) error {
		tx.AscendRange("", "sutxo-", "sutxp-", func(key, value string) bool {
			logging.Debugf("Loading key %s\n", key)
			w.stealthUtxos = append(w.stealthUtxos, StealthUtxoFromBytes([]byte(value)))
			return true
		})
		return nil
	})

	w.db.View(func(tx *buntdb.Tx) error {
		tx.AscendRange("", "autxo-", "autxp-", func(key, value string) bool {
			logging.Debugf("Loading key %s\n", key)
			w.assetUtxos = append(w.assetUtxos, OpenAssetUtxoFromBytes([]byte(value)))
			return true
		})
		return nil
	})

	w.db.View(func(tx *buntdb.Tx) error {
		tx.AscendRange("", "asset-", "asseu-", func(key, value string) bool {
			logging.Debugf("Loading key %s\n", key)
			w.assets = append(w.assets, OpenAssetFromBytes([]byte(value)))
			return true
		})
		return nil
	})
}

func (w *Wallet) UpdateClient(c *rpcclient.Client) {
	w.rpcClient = &(*c)
}

/*
func (w *Wallet) VertcoinAddress() (string, error) {
	return bech32.SegWitV0Encode(w.coinNetwork.Bech32Prefix, w.pubKeyHash[:])
}

func (w *Wallet) AssetsAddress() (string, error) {
	return bech32.SegWitV0Encode(w.config.Network.AssetAddressPrefix, w.pubKeyHash[:])
}

func (w *Wallet) StealthAddress() (string, error) {
	return bech32.Encode(w.config.Network.StealthAddressPrefix, w.pubKey.SerializeCompressed()), nil
}
*/

func (w *Wallet) Balance() uint64 {
	value := uint64(0)
	for _, u := range w.utxos {
		value += u.Value
	}
	return value
}

func (w *Wallet) StealthBalance() uint64 {
	value := uint64(0)
	for _, u := range w.stealthUtxos {
		value += u.Utxo.Value
	}
	return value
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
			w.updateAsset(a)
		}
	}
}

func (w *Wallet) UnfollowAsset(assetID []byte) {
	for _, a := range w.assets {
		if bytes.Equal(a.AssetID, assetID) {
			a.Follow = true
			w.updateAsset(a)
		}
	}
}

func (w *Wallet) ProcessTransaction(tx *wire.MsgTx) {
	if IsOpenAssetTransaction(tx) {
		w.processOpenAssetTransaction(tx)
	} else if IsStealthTransaction(tx) {
		w.processStealthTransaction(tx)
	} else {
		w.processNormalTransaction(tx)
	}
}

func (w *Wallet) processNormalTransaction(tx *wire.MsgTx) {
	/*for i, out := range tx.TxOut {
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
	*/
	w.markTxStealthInputsAsSpent(tx)
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
			w.db.Update(func(dtx *buntdb.Tx) error {
				key := fmt.Sprintf("utxo-%s-%d", w.utxos[removeIndex].TxHash.String(), w.utxos[removeIndex].Outpoint)
				_, err := dtx.Delete(key)
				return err
			})
			w.utxos = append(w.utxos[:removeIndex], w.utxos[removeIndex+1:]...)

		}
	}
}

func (w *Wallet) registerUtxo(utxo Utxo) {
	w.utxos = append(w.utxos, utxo)
	err := w.db.Update(func(dtx *buntdb.Tx) error {
		key := fmt.Sprintf("utxo-%s-%d", utxo.TxHash.String(), utxo.Outpoint)
		logging.Debugf("Saving key %s\n", key)
		_, _, err := dtx.Set(key, string(utxo.Bytes()), nil)
		return err
	})
	if err != nil {
		logging.Errorf("[Wallet] Error registering utxo: %s", err.Error())
	}
}

func (w *Wallet) registerStealthUtxo(utxo StealthUtxo) {
	w.stealthUtxos = append(w.stealthUtxos, utxo)
	err := w.db.Update(func(dtx *buntdb.Tx) error {
		key := fmt.Sprintf("sutxo-%s-%d", utxo.Utxo.TxHash.String(), utxo.Utxo.Outpoint)
		logging.Debugf("Saving key %s\n", key)
		_, _, err := dtx.Set(key, string(utxo.Bytes()), nil)
		return err
	})
	if err != nil {
		logging.Errorf("[Wallet] Error registering stealth utxo: %s", err.Error())
	}
}

func (w *Wallet) registerAssetUtxo(utxo OpenAssetUtxo) {
	w.assetUtxos = append(w.assetUtxos, utxo)
	err := w.db.Update(func(dtx *buntdb.Tx) error {
		key := fmt.Sprintf("autxo-%s-%d", utxo.Utxo.TxHash.String(), utxo.Utxo.Outpoint)
		logging.Debugf("Saving key %s\n", key)
		_, _, err := dtx.Set(key, string(utxo.Bytes()), nil)
		return err
	})
	if err != nil {
		logging.Errorf("[Wallet] Error registering asset utxo: %s", err.Error())
	}
}

func (w *Wallet) registerAsset(asset OpenAsset) {
	w.assets = append(w.assets, &asset)
	err := w.db.Update(func(dtx *buntdb.Tx) error {
		key := fmt.Sprintf("asset-%x", asset.AssetID)
		logging.Debugf("Saving key %s\n", key)
		_, _, err := dtx.Set(key, string((&asset).Bytes()), nil)
		return err
	})
	if err != nil {
		logging.Errorf("[Wallet] Error registering asset: %s", err.Error())
	}
}

func (w *Wallet) updateAsset(asset *OpenAsset) {
	w.db.Update(func(dtx *buntdb.Tx) error {
		key := fmt.Sprintf("asset-%x", asset.AssetID)
		_, _, err := dtx.Set(key, string(asset.Bytes()), nil)
		return err
	})
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
		// TODO Derive proper change address and use that
		// tx.AddTxOut(wire.NewTxOut(int64(valueAdded-totalValueNeeded), util.DirectWPKHScriptFromPKH(w.MyPKH())))

	}

	return nil
}

func (w *Wallet) AddStealthInputsAndChange(tx *wire.MsgTx, totalValueNeeded uint64) error {
	valueAdded := uint64(0)
	utxosToAdd := []StealthUtxo{}
	for _, utxo := range w.stealthUtxos {
		utxosToAdd = append(utxosToAdd, utxo)
		valueAdded += utxo.Utxo.Value
		if valueAdded > totalValueNeeded {
			break
		}

	}
	if valueAdded < totalValueNeeded {
		return fmt.Errorf("Insufficient stealth balance")
	}

	for _, utxo := range utxosToAdd {
		tx.AddTxIn(wire.NewTxIn(&wire.OutPoint{utxo.Utxo.TxHash, utxo.Utxo.Outpoint}, nil, nil))
	}

	// Add change output when there's more than dust left, otherwise give to miners
	if valueAdded-totalValueNeeded > MINOUTPUT {
		// TODO Derive proper change address and use that
		// tx.AddTxOut(wire.NewTxOut(int64(valueAdded-totalValueNeeded), util.DirectWPKHScriptFromPKH(w.MyPKH())))
	}

	return nil
}

func (w *Wallet) GenerateNormalSendTx(tx SendTransaction, stealthInputs bool) (*wire.MsgTx, error) {
	stx := wire.NewMsgTx(1)
	neededInputs := uint64(100000) // minfee

	stx.AddTxOut(wire.NewTxOut(int64(tx.Amount), util.DirectWPKHScriptFromPKH(tx.RecipientPkh)))
	neededInputs += tx.Amount

	if stealthInputs {
		err := w.AddStealthInputsAndChange(stx, uint64(neededInputs))
		if err != nil {
			return nil, err
		}
	} else {
		err := w.AddInputsAndChange(stx, uint64(neededInputs))
		if err != nil {
			return nil, err
		}
	}

	return stx, nil
}
