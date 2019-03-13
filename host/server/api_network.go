package server

import (
	"encoding/json"
	"net/http"
)

type NetworkResponse struct {
	NetworkName string
}

func (h *HttpServer) Network(w http.ResponseWriter, r *http.Request) {
	resp := NetworkResponse{NetworkName: h.config.Network.Name}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(200)
	json.NewEncoder(w).Encode(resp)
}
