package server

import (
	"encoding/json"
	"net/http"
)

func (h *HttpServer) SyncStatus(w http.ResponseWriter, r *http.Request) {
	resp := h.blockProcessor.GetSyncStatus()
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(200)
	json.NewEncoder(w).Encode(resp)
}
