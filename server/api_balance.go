package server

import (
	"encoding/json"
	"net/http"
)

type BalanceResponse struct {
	TotalBalance uint64
}

func (h *HttpServer) Balance(w http.ResponseWriter, r *http.Request) {
	resp := BalanceResponse{TotalBalance: h.wallet.Balance()}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(200)
	json.NewEncoder(w).Encode(resp)
}
