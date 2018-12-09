package server

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/gertjaap/vertcoin-openassets/wallet"
)

type NewAssetParameters struct {
	Metadata    string
	TotalSupply uint64
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
	tx.Metadata = []byte(params.Metadata)
	tx.Issuances = append(tx.Issuances, wallet.OpenAssetIssuanceOutput{
		Value:        params.TotalSupply,
		RecipientPkh: h.wallet.MyPKH(),
	})

	wireTx, err := wallet.GenerateOpenAssetTx(h.wallet, tx)
	if err != nil {
		h.writeError(w, fmt.Errorf("Error decoding json: %s", err.Error()))
		return
	}

	err = h.wallet.SignMyInputs(wireTx)
	if err != nil {
		h.writeError(w, fmt.Errorf("Error decoding json: %s", err.Error()))
		return
	}

	var buf bytes.Buffer
	writer := bufio.NewWriter(&buf)
	wireTx.Serialize(writer)
	writer.Flush()
	fmt.Printf("Hex transaction:\n%x\n", buf.Bytes())

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
