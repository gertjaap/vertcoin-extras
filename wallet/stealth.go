package wallet

import (
	"bufio"
	"bytes"
	"crypto/rand"
	"fmt"

	"github.com/btcsuite/btcd/btcec"
	"github.com/btcsuite/btcd/wire"
	"github.com/btcsuite/btcutil"
	"github.com/gertjaap/vertcoin-extras/util"
	"github.com/mit-dci/lit/lnutil"
	"github.com/tidwall/buntdb"
)

func isStealthData(b []byte) bool {
	if b[0] == 0x5f && b[1] == 0x51 && b[2] == 0x01 && b[3] == 0x00 { // Stealth 1.0
		return true
	}
	return false
}

func (w *Wallet) processStealthTransaction(tx *wire.MsgTx) {
	// first find the marker output and extract the amounts
	stealthData := extractStealthData(tx)

	if stealthData == nil || len(stealthData) == 0 {
		return
	}

	// Try decrypting
	priv, err := w.GetStealthPrivateKey(stealthData[4:])
	if err == nil {
		keyHash := util.KeyHashFromPkScript(tx.TxOut[0].PkScript)
		stealthPKH := btcutil.Hash160(priv.PubKey().SerializeCompressed())
		if bytes.Equal(keyHash, stealthPKH) {
			// the public key of the decrypted key combined with our private key
			// matches the PKH in the script. So yeah, we want this!
			var stxo StealthUtxo
			stxo.Utxo = Utxo{
				TxHash:   tx.TxHash(),
				Outpoint: 0,
				Value:    uint64(tx.TxOut[0].Value),
				PkScript: tx.TxOut[0].PkScript,
			}
			// Store the encrypted one-time key. We'll decrypt it when we need to
			stxo.EncOTK = stealthData[4:]
			w.registerStealthUtxo(stxo)
		} else {
			fmt.Printf("Unexpected stealth pubkey. PKScript is [%x], calculated is [%x]\n", keyHash, stealthPKH)
		}

	}

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

	w.markTxStealthInputsAsSpent(tx)
	w.markTxInputsAsSpent(tx)
}

func extractStealthData(tx *wire.MsgTx) []byte {
	foundStealth := false
	b := make([]byte, 0)
	buf := bytes.NewBuffer(b)
	for _, txo := range tx.TxOut {
		if len(txo.PkScript) > 0 && txo.PkScript[0] == 0x6a {
			idx := 1
			for {
				opCode := txo.PkScript[idx]
				if opCode >= 0x05 && opCode <= 0x4b { // pushdata
					if isStealthData(txo.PkScript[idx+1:idx+1+int(opCode)]) || foundStealth {
						foundStealth = true
						buf.Write(txo.PkScript[idx+1 : idx+1+int(opCode)])
					}
					idx += 1 + int(opCode)
				} else {
					idx++
				}
				if len(txo.PkScript) <= idx {
					break
				}
			}

		}
		if foundStealth {
			return buf.Bytes()
		}
	}
	return nil
}

func (w *Wallet) markTxStealthInputsAsSpent(tx *wire.MsgTx) {
	for _, in := range tx.TxIn {
		removeIndex := -1
		for j, out := range w.stealthUtxos {
			if in.PreviousOutPoint.Hash.IsEqual(&out.Utxo.TxHash) && in.PreviousOutPoint.Index == out.Utxo.Outpoint {
				// Spent!
				removeIndex = j
				break
			}
		}
		if removeIndex >= 0 {
			w.db.Update(func(dtx *buntdb.Tx) error {
				key := fmt.Sprintf("sutxo-%s-%d", w.stealthUtxos[removeIndex].Utxo.TxHash.String(), w.stealthUtxos[removeIndex].Utxo.Outpoint)
				_, err := dtx.Delete(key)
				return err
			})
			w.stealthUtxos = append(w.stealthUtxos[:removeIndex], w.stealthUtxos[removeIndex+1:]...)

		}
	}
}

func IsStealthTransaction(tx *wire.MsgTx) bool {
	b := extractStealthData(tx)

	return b != nil && len(b) > 0
}

func (w *Wallet) GenerateStealthTx(tx StealthTransaction, stealthInputs bool) (*wire.MsgTx, error) {
	stx := wire.NewMsgTx(1)
	neededInputs := uint64(100000) // minfee

	// generate one-time key
	priv := [32]byte{}
	n, err := rand.Read(priv[:])
	if err != nil {
		return nil, err
	}
	if n != 32 {
		return nil, fmt.Errorf("Unable to read 32 bytes for one-time key")
	}

	_, pub := btcec.PrivKeyFromBytes(btcec.S256(), priv[:])
	pub33 := [33]byte{}
	copy(pub33[:], pub.SerializeCompressed())

	// generate combined pubkey
	cpk := lnutil.CombinePubs(tx.RecipientPubKey, pub33)
	cpkh := [20]byte{}
	copy(cpkh[:], btcutil.Hash160(cpk[:]))

	stx.AddTxOut(wire.NewTxOut(int64(tx.Amount), util.DirectWPKHScriptFromPKH(cpkh)))
	neededInputs += tx.Amount

	// write encrypted priv key into OP_RETURN
	enc, err := util.EncryptECIES(tx.RecipientPubKey, priv[:])
	if err != nil {
		return nil, err
	}
	var buf bytes.Buffer
	writer := bufio.NewWriter(&buf)
	writer.Write([]byte{0x5f, 0x51}) // Stealth marker
	writer.Write([]byte{0x01, 0x00}) // Stealth Version 1
	writer.Write(enc)
	writer.Flush()

	b := buf.Bytes()
	startIdx := 0
	var scriptBuf bytes.Buffer
	scriptBuf.WriteByte(0x6A) // OP_RETURN

	for len(b) > startIdx {
		lenToWrite := len(b) - startIdx
		if lenToWrite > 75 {
			lenToWrite = 75
		}
		scriptBuf.WriteByte(byte(lenToWrite))
		scriptBuf.Write(b[startIdx : startIdx+lenToWrite])
		startIdx += lenToWrite

	}
	stx.AddTxOut(wire.NewTxOut(0, scriptBuf.Bytes()))

	if stealthInputs {
		err = w.AddStealthInputsAndChange(stx, uint64(neededInputs))
		if err != nil {
			return nil, err
		}
	} else {
		err = w.AddInputsAndChange(stx, uint64(neededInputs))
		if err != nil {
			return nil, err
		}
	}

	return stx, nil
}

func (w *Wallet) FindStealthUtxoFromTxIn(txi *wire.TxIn) (Utxo, *btcec.PrivateKey, error) {
	for _, sout := range w.stealthUtxos {
		if txi.PreviousOutPoint.Hash.IsEqual(&sout.Utxo.TxHash) && txi.PreviousOutPoint.Index == sout.Utxo.Outpoint {
			cpriv, err := w.GetStealthPrivateKey(sout.EncOTK)
			if err == nil {
				return sout.Utxo, cpriv, nil
			}
		}
	}
	return Utxo{}, nil, fmt.Errorf("Stealth utxo not found")
}

func (w *Wallet) GetStealthPrivateKey(encOTK []byte) (*btcec.PrivateKey, error) {
	priv, err := util.DecryptECIES(w.privateKey, encOTK)
	if err != nil {
		return nil, err
	}

	fmt.Printf("Decrypted OTK: %x\n", priv)

	pk, _ := btcec.PrivKeyFromBytes(btcec.S256(), priv)
	cpriv := util.CombinePrivateKeys(w.privateKey, pk)
	return cpriv, nil
}
