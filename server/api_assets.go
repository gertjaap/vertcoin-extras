package server

import (
	"encoding/hex"
	"encoding/json"
	"net/http"
)

type AssetsResponse struct {
	Assets []AssetResponseAsset
}

type AssetResponseAsset struct {
	AssetID  string
	Ticker   string
	Decimals uint8
}

func (h *HttpServer) Assets(w http.ResponseWriter, r *http.Request) {
	resp := AssetsResponse{}

	assets := h.wallet.Assets()
	for _, a := range assets {
		resp.Assets = append(resp.Assets, AssetResponseAsset{
			AssetID:  hex.EncodeToString(a.AssetID),
			Ticker:   a.Metadata.Ticker,
			Decimals: a.Metadata.Decimals,
		})
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(200)
	json.NewEncoder(w).Encode(resp)
}
