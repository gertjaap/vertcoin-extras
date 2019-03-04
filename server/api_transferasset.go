package server

import (
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/gertjaap/vertcoin-extras/wallet"
	"github.com/mit-dci/lit/bech32"
)

type TransferAssetParameters struct {
	AssetID          string
	Amount           uint64
	RecipientAddress string
}

type TransferAssetResult struct {
	TxID string
}

func (h *HttpServer) TransferAsset(w http.ResponseWriter, r *http.Request) {
	decoder := json.NewDecoder(r.Body)
	var params TransferAssetParameters
	err := decoder.Decode(&params)
	if err != nil {
		h.writeError(w, fmt.Errorf("Error decoding json: %s", err.Error()))
		return
	}

	assetID, err := hex.DecodeString(params.AssetID)
	if err != nil {
		h.writeError(w, fmt.Errorf("Error decoding asset ID: %s", err.Error()))
		return
	}

	var recipientPkh [20]byte
	decoded, err := bech32.SegWitAddressDecode(params.RecipientAddress)
	if err != nil {
		h.writeError(w, fmt.Errorf("Error decoding recipient address: %s", err.Error()))
		return
	}
	copy(recipientPkh[:], decoded[2:]) // skip version and pushdata byte returned by SegWitAddressDecode

	var tx wallet.OpenAssetTransaction
	tx.Transfers = append(tx.Transfers, wallet.OpenAssetTransferOutput{
		AssetID:      assetID,
		Value:        params.Amount,
		RecipientPkh: recipientPkh,
	})

	wireTx, err := h.wallet.GenerateOpenAssetTx(tx)
	if err != nil {
		h.writeError(w, fmt.Errorf("Error generating transaction: %s", err.Error()))
		return
	}

	err = h.wallet.SignMyInputs(wireTx)
	if err != nil {
		h.writeError(w, fmt.Errorf("Error signing: %s", err.Error()))
		return
	}

	txid, err := h.wallet.SendTransaction(wireTx)
	if err != nil {
		h.writeError(w, fmt.Errorf("Error sending: %s", err.Error()))
		return
	}

	var res NewAssetResult
	res.TxID = txid.String()
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(200)
	json.NewEncoder(w).Encode(res)
}
