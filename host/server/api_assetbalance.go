package server

import (
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/gorilla/mux"
)

type AssetBalanceResponse struct {
	TotalBalance uint64
}

func (h *HttpServer) AssetBalance(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	assetID, err := hex.DecodeString(vars["assetID"])
	if err != nil {
		h.writeError(w, fmt.Errorf("Error decoding json: %s", err.Error()))
		return
	}

	resp := AssetBalanceResponse{TotalBalance: h.wallet.AssetBalance(assetID)}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(200)
	json.NewEncoder(w).Encode(resp)
}
