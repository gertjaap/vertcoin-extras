package wallet

import (
	"github.com/btcsuite/btcd/chaincfg/chainhash"
	"github.com/btcsuite/btcd/txscript"
	"github.com/btcsuite/btcd/wire"
)

func (w *Wallet) SignMyInputs(tx *wire.MsgTx) error {
	// generate tx-wide hashCache for segwit stuff
	// might not be needed (non-witness) but make it anyway
	hCache := txscript.NewTxSigHashes(tx)
	witStash := make([][][]byte, len(tx.TxIn))

	for i, txi := range tx.TxIn {
		utxo, err := w.FindUtxoFromTxIn(txi)
		if err != nil {
			continue
		}

		witStash[i], err = txscript.WitnessSignature(tx, hCache, i,
			int64(utxo.Value), utxo.PkScript, txscript.SigHashAll, w.privateKey, true)
		if err != nil {
			return err
		}
	}

	for i, txi := range tx.TxIn {
		utxo, priv, err := w.FindStealthUtxoFromTxIn(txi)
		if err != nil {
			continue
		}

		witStash[i], err = txscript.WitnessSignature(tx, hCache, i,
			int64(utxo.Value), utxo.PkScript, txscript.SigHashAll, priv, true)
		if err != nil {
			return err
		}
	}

	// swap sigs into sigScripts in txins
	for i, txin := range tx.TxIn {
		if witStash[i] != nil {
			txin.Witness = witStash[i]
			txin.SignatureScript = nil
		}
	}

	return nil
}

func (w *Wallet) SendTransaction(tx *wire.MsgTx) (*chainhash.Hash, error) {
	return w.rpcClient.SendRawTransaction(tx, false)
}
