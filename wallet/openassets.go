package wallet

import (
	"bufio"
	"bytes"
	"fmt"

	"github.com/btcsuite/btcd/wire"
	"github.com/btcsuite/btcutil"
	"github.com/gertjaap/vertcoin-openassets/leb128"
	"github.com/gertjaap/vertcoin-openassets/util"
)

const MINOUTPUT uint64 = 500

func isOpenAssetMarkerData(b []byte) bool {
	if b[0] == 0x4f && b[1] == 0x41 && b[2] == 0x01 && b[3] == 0x00 { // Open Asset 1.0
		return true
	}
	return false
}

func (w *Wallet) processOpenAssetTransaction(tx *wire.MsgTx) {
	// first find the marker output and extract the amounts
	openAssetData, txoIndex := extractOpenAssetMarkerData(tx)

	fmt.Printf("Found openasset data: %x\n", openAssetData)

	buf := bytes.NewBuffer(openAssetData[4:]) // Skip marker and version
	reader := bufio.NewReader(buf)
	numAmounts, _ := wire.ReadVarInt(reader, 0)
	amounts := []uint64{}
	for i := uint64(0); i < numAmounts; i++ {
		amounts = append(amounts, leb128.MustReadVarUint64(reader))
	}

	fmt.Printf("Extracted %d amounts\n", len(amounts))

	// TODO: Verify if total inputs match total transfer outputs.

	for i, txo := range tx.TxOut {
		amountIdx := i
		if i == txoIndex { // Skip marker output
			continue
		}
		if i > txoIndex {
			amountIdx--
		}
		amount := uint64(0)
		if len(amounts) > amountIdx {
			amount = amounts[amountIdx]
		}

		utxo := Utxo{
			TxHash:   tx.TxHash(),
			Outpoint: uint32(i),
			Value:    uint64(txo.Value),
			PkScript: txo.PkScript,
		}
		if amount > 0 {
			var oatxo OpenAssetUtxo
			keyHash := util.KeyHashFromPkScript(txo.PkScript)
			oatxo.Ours = bytes.Equal(keyHash, w.pubKeyHash[:])
			oatxo.Utxo = utxo
			oatxo.AssetValue = amount
			txHash := tx.TxHash()
			oatxo.AssetID = btcutil.Hash160(txHash[:])
			w.registerAssetUtxo(oatxo)
		} else {
			w.registerUtxo(utxo)
		}
	}

	w.markTxInputsAsSpent(tx)

}

func extractOpenAssetMarkerData(tx *wire.MsgTx) ([]byte, int) {
	for i, txo := range tx.TxOut {
		if len(txo.PkScript) > 0 && txo.PkScript[0] == 0x6a {
			idx := 1
			for {
				opCode := txo.PkScript[idx]
				if opCode >= 0x05 && opCode <= 0x4b { // pushdata
					if isOpenAssetMarkerData(txo.PkScript[idx+1 : idx+1+int(opCode)]) {
						return txo.PkScript[idx+1 : idx+1+int(opCode)], i
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
	}
	return []byte{}, -1
}

func IsOpenAssetTransaction(tx *wire.MsgTx) bool {
	_, pos := extractOpenAssetMarkerData(tx)
	return pos >= 0
}

func GenerateOpenAssetTx(w *Wallet, tx OpenAssetTransaction) (*wire.MsgTx, error) {
	oatx := wire.NewMsgTx(1)
	neededInputs := uint64(100000) // minfee

	// first we add the OA inputs
	for _, oai := range tx.AssetInputs {
		oatx.AddTxIn(wire.NewTxIn(&wire.OutPoint{Hash: oai.Utxo.TxHash, Index: oai.Utxo.Outpoint}, nil, nil))
		neededInputs -= oai.Utxo.Value
	}

	// for each issuance output we add an output with above dust value, otherwise we can't
	// use it as input next time.
	for _, oai := range tx.Issuances {
		oatx.AddTxOut(wire.NewTxOut(int64(MINOUTPUT), util.DirectWPKHScriptFromPKH(oai.RecipientPkh)))
		neededInputs += MINOUTPUT
	}

	// Here we add the marker output
	var buf bytes.Buffer
	writer := bufio.NewWriter(&buf)
	writer.Write([]byte{0x4f, 0x41})                                         // Open Asset Marker
	writer.Write([]byte{0x01, 0x00})                                         // Open Asset Version 1
	wire.WriteVarInt(writer, 0, uint64(len(tx.Issuances)+len(tx.Transfers))) // Number of amounts

	for _, oai := range tx.Issuances {
		leb128.WriteVarUint64(writer, oai.Value)
	}
	for _, oat := range tx.Transfers {
		leb128.WriteVarUint64(writer, oat.Value)
	}

	wire.WriteVarInt(writer, 0, uint64(len(tx.Metadata)))
	writer.Write(tx.Metadata)
	writer.Flush()
	var scriptBuf bytes.Buffer
	scriptBuf.WriteByte(0x6A) // OP_RETURN
	scriptBuf.WriteByte(byte(buf.Len()))
	scriptBuf.Write(buf.Bytes())
	oatx.AddTxOut(wire.NewTxOut(0, scriptBuf.Bytes()))

	// for each transfer output we add an output with above dust value, otherwise we can't
	// use it as input next time.
	for _, oat := range tx.Transfers {
		oatx.AddTxOut(wire.NewTxOut(int64(MINOUTPUT), util.DirectWPKHScriptFromPKH(oat.RecipientPkh)))
		neededInputs += MINOUTPUT
	}

	err := w.AddInputsAndChange(oatx, uint64(neededInputs))
	if err != nil {
		return nil, err
	}

	return oatx, nil
}
