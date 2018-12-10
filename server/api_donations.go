package server

import (
	"encoding/json"
	"fmt"
	"net/http"
)

func (h *HttpServer) DisableDonations(w http.ResponseWriter, r *http.Request) {
	err := h.config.DisableDonations()
	if err != nil {
		h.writeError(w, fmt.Errorf("Error disabling donations: %s", err.Error()))
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(200)
	json.NewEncoder(w).Encode(true)
}

func (h *HttpServer) EnableDonations(w http.ResponseWriter, r *http.Request) {
	err := h.config.EnableDonations()
	if err != nil {
		h.writeError(w, fmt.Errorf("Error enabling donations: %s", err.Error()))
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(200)
	json.NewEncoder(w).Encode(true)
}

func (h *HttpServer) DonationStatus(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(200)
	json.NewEncoder(w).Encode(h.config.Donate)
}
