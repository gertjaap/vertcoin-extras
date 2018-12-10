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

const MINOUTPUT uint64 = 1000

func isOpenAssetMarkerData(b []byte) bool {
	if b[0] == 0x4f && b[1] == 0x41 && b[2] == 0x01 && b[3] == 0x00 { // Open Asset 1.0
		return true
	}
	return false
}

func (w *Wallet) processOpenAssetTransaction(tx *wire.MsgTx) {
	// first find the marker output and extract the amounts
	openAssetData, txoIndex := extractOpenAssetMarkerData(tx)

	buf := bytes.NewBuffer(openAssetData[4:]) // Skip marker and version
	reader := bufio.NewReader(buf)
	numAmounts, _ := wire.ReadVarInt(reader, 0)
	amounts := []uint64{}
	for i := uint64(0); i < numAmounts; i++ {
		amounts = append(amounts, leb128.MustReadVarUint64(reader))
	}

	metadata, _ := wire.ReadVarBytes(reader, 0, 150, "")
	assetTicker := ""
	assetDecimals := uint64(0)
	if len(metadata) > 0 {
		mdbuf := bytes.NewBuffer(metadata)
		mdreader := bufio.NewReader(mdbuf)

		assetTicker, _ = wire.ReadVarString(mdreader, 0)
		assetDecimals, _ = wire.ReadVarInt(mdreader, 0)
	}
	// Fetch the assetID and total amount of inputs
	inputAssetId, totalInputAmount := w.getAssetIDAndTotalAmount(tx)

	totalTransferOutputs := uint64(0)
	// Verify if total inputs matches the total
	for i, _ := range tx.TxOut {
		if i > txoIndex { // transfer
			amount := uint64(0)
			if len(amounts) > i-1 {
				amount = amounts[i-1]
			}
			totalTransferOutputs += amount
		}
	}

	if totalTransferOutputs > totalInputAmount {
		// Don't process - invalid TX
		fmt.Printf("Invalid Transaction : Insufficient inputs [%d/%d] - ignoring\n", totalTransferOutputs, totalInputAmount)
		return
	}

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
		keyHash := util.KeyHashFromPkScript(txo.PkScript)

		if amount > 0 {
			var oatxo OpenAssetUtxo
			oatxo.Ours = bytes.Equal(keyHash, w.pubKeyHash[:])
			oatxo.Utxo = utxo
			oatxo.AssetValue = amount
			if i < txoIndex { // Issuance
				// Register the new asset
				txHash := tx.TxHash()
				oatxo.AssetID = btcutil.Hash160(txHash[:])

				w.registerAsset(OpenAsset{
					AssetID: oatxo.AssetID,
					Follow:  oatxo.Ours,
					Metadata: OpenAssetMetadata{
						Decimals: uint8(assetDecimals),
						Ticker:   assetTicker,
					},
				})
			} else {
				oatxo.AssetID = inputAssetId
			}
			w.registerAssetUtxo(oatxo)
		} else {
			if bytes.Equal(keyHash, w.pubKeyHash[:]) {
				w.registerUtxo(utxo)
			}
		}
	}

	w.markTxInputsAsSpent(tx)
	w.markOpenAssetTxInputsAsSpent(tx)
}

func (w *Wallet) getAssetIDAndTotalAmount(tx *wire.MsgTx) ([]byte, uint64) {
	assetID := []byte{}
	var totalAmount uint64

	for _, in := range tx.TxIn {
		for _, out := range w.assetUtxos {
			if in.PreviousOutPoint.Hash.IsEqual(&out.Utxo.TxHash) && in.PreviousOutPoint.Index == out.Utxo.Outpoint {
				if bytes.Equal([]byte{}, assetID) {
					assetID = out.AssetID
				}
				if bytes.Equal(out.AssetID, assetID) {
					totalAmount += out.AssetValue
				}
				break
			}
		}
	}

	return assetID, totalAmount
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

func (w *Wallet) GenerateOpenAssetTx(tx OpenAssetTransaction) (*wire.MsgTx, error) {
	oatx := wire.NewMsgTx(1)
	neededInputs := uint64(100000) // minfee

	if len(tx.Transfers) > 0 {
		err := w.addOpenAssetInputsAndChange(&tx, tx.Transfers[0].AssetID)
		if err != nil {
			return nil, err
		}
	}

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

	var metadataBuf bytes.Buffer
	if (len(tx.Metadata.Ticker) + int(tx.Metadata.Decimals)) > 0 {
		mdw := bufio.NewWriter(&metadataBuf)
		wire.WriteVarString(mdw, 0, tx.Metadata.Ticker)
		wire.WriteVarInt(mdw, 0, uint64(tx.Metadata.Decimals))
		mdw.Flush()
	}

	wire.WriteVarBytes(writer, 0, metadataBuf.Bytes())

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

func (w *Wallet) addOpenAssetInputsAndChange(tx *OpenAssetTransaction, assetID []byte) error {
	totalNeeded := uint64(0)
	for _, oat := range tx.Transfers {
		totalNeeded += oat.Value
	}

	totalAdded := uint64(0)
	assetUtxosToAdd := []OpenAssetUtxo{}
	for _, autxo := range w.assetUtxos {
		if autxo.Ours && bytes.Equal(autxo.AssetID, assetID) {
			totalAdded += autxo.AssetValue
			assetUtxosToAdd = append(assetUtxosToAdd, autxo)
		}
	}

	if totalAdded < totalNeeded {
		return fmt.Errorf("Insufficient asset balance. Wanted %d got %d", totalNeeded, totalAdded)
	}

	for _, autxo := range assetUtxosToAdd {
		tx.AssetInputs = append(tx.AssetInputs, autxo)
	}

	if totalAdded > totalNeeded {
		// Change output
		tx.Transfers = append(tx.Transfers, OpenAssetTransferOutput{
			AssetID:      assetID,
			RecipientPkh: w.MyPKH(),
			Value:        totalAdded - totalNeeded,
		})
	}

	return nil
}

func (w *Wallet) markOpenAssetTxInputsAsSpent(tx *wire.MsgTx) {
	for _, in := range tx.TxIn {
		removeIndex := -1
		for j, out := range w.assetUtxos {
			if in.PreviousOutPoint.Hash.IsEqual(&out.Utxo.TxHash) && in.PreviousOutPoint.Index == out.Utxo.Outpoint {
				// Spent!
				removeIndex = j
				break
			}
		}
		if removeIndex >= 0 {
			w.assetUtxos = append(w.assetUtxos[:removeIndex], w.assetUtxos[removeIndex+1:]...)
		}
	}
}
