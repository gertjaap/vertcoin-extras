package server

import (
	"encoding/json"
	"net/http"
)

type BalanceResponse struct {
	TotalBalance   uint64
	StealthBalance uint64
}

func (h *HttpServer) Balance(w http.ResponseWriter, r *http.Request) {
	balance := h.wallet.Balance()
	stealthBalance := h.wallet.StealthBalance()
	resp := BalanceResponse{TotalBalance: balance + stealthBalance, StealthBalance: stealthBalance}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(200)
	json.NewEncoder(w).Encode(resp)
}
