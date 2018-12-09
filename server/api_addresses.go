package server

import (
	"encoding/json"
	"fmt"
	"net/http"
)

type AddressesResponse struct {
	VertcoinAddress string
	AssetAddress    string
}

func (h *HttpServer) Addresses(w http.ResponseWriter, r *http.Request) {
	vtcAdr, err := h.wallet.VertcoinAddress()
	if err != nil {
		h.writeError(w, fmt.Errorf("Error fetching VTC Address: %s", err.Error()))
		return
	}

	assAdr, err := h.wallet.AssetsAddress()
	if err != nil {
		h.writeError(w, fmt.Errorf("Error fetching Asset Address: %s", err.Error()))
		return
	}

	resp := AddressesResponse{
		VertcoinAddress: vtcAdr,
		AssetAddress:    assAdr,
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(200)
	json.NewEncoder(w).Encode(resp)
}
