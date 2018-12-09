package server

import (
	"encoding/hex"
	"encoding/json"
	"net/http"
)

type AssetsResponse struct {
	Assets []string
}

func (h *HttpServer) Assets(w http.ResponseWriter, r *http.Request) {
	resp := AssetsResponse{Assets: []string{}}

	assets := h.wallet.Assets()
	for _, asset := range assets {
		resp.Assets = append(resp.Assets, hex.EncodeToString(asset))
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(200)
	json.NewEncoder(w).Encode(resp)
}
