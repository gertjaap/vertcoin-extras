package server

import (
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/gorilla/mux"
)

type AssetsResponse struct {
	Assets []AssetResponseAsset
}

type AssetResponseAsset struct {
	AssetID  string
	Ticker   string
	Decimals uint8
	Balance  uint64
}

func (h *HttpServer) AllAssets(w http.ResponseWriter, r *http.Request) {
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

func (h *HttpServer) MyAssets(w http.ResponseWriter, r *http.Request) {
	resp := AssetsResponse{}

	assets := h.wallet.Assets()
	for _, a := range assets {
		if a.Follow {

			resp.Assets = append(resp.Assets, AssetResponseAsset{
				AssetID:  hex.EncodeToString(a.AssetID),
				Ticker:   a.Metadata.Ticker,
				Decimals: a.Metadata.Decimals,
				Balance:  h.wallet.AssetBalance(a.AssetID),
			})
		}
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(200)
	json.NewEncoder(w).Encode(resp)
}

func (h *HttpServer) FollowAsset(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	assetID, err := hex.DecodeString(vars["assetID"])
	if err != nil {
		h.writeError(w, fmt.Errorf("Error decoding json: %s", err.Error()))
		return
	}
	h.wallet.FollowAsset(assetID)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(200)
	json.NewEncoder(w).Encode(true)
}

func (h *HttpServer) UnfollowAsset(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	assetID, err := hex.DecodeString(vars["assetID"])
	if err != nil {
		h.writeError(w, fmt.Errorf("Error decoding json: %s", err.Error()))
		return
	}
	h.wallet.UnfollowAsset(assetID)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(200)
	json.NewEncoder(w).Encode(true)
}
