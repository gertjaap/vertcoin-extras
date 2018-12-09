package server

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/gertjaap/vertcoin-openassets/wallet"
)

type NewAssetParameters struct {
	TotalSupply uint64
	Decimals    uint8
	Ticker      string
}

type NewAssetResult struct {
	TxID string
}

func (h *HttpServer) NewAsset(w http.ResponseWriter, r *http.Request) {
	decoder := json.NewDecoder(r.Body)
	var params NewAssetParameters
	err := decoder.Decode(&params)
	if err != nil {
		h.writeError(w, fmt.Errorf("Error decoding json: %s", err.Error()))
		return
	}

	var tx wallet.OpenAssetTransaction
	tx.Metadata.Decimals = params.Decimals
	tx.Metadata.Ticker = params.Ticker
	tx.Issuances = append(tx.Issuances, wallet.OpenAssetIssuanceOutput{
		Value:        params.TotalSupply,
		RecipientPkh: h.wallet.MyPKH(),
	})

	wireTx, err := h.wallet.GenerateOpenAssetTx(tx)
	if err != nil {
		h.writeError(w, fmt.Errorf("Error decoding json: %s", err.Error()))
		return
	}

	err = h.wallet.SignMyInputs(wireTx)
	if err != nil {
		h.writeError(w, fmt.Errorf("Error decoding json: %s", err.Error()))
		return
	}

	txid, err := h.wallet.SendTransaction(wireTx)
	if err != nil {
		h.writeError(w, fmt.Errorf("Error decoding json: %s", err.Error()))
		return
	}

	var res NewAssetResult
	res.TxID = txid.String()
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(200)
	json.NewEncoder(w).Encode(res)
}
