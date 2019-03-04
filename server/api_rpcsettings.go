package server

import (
	"encoding/json"
	"fmt"
	"net/http"
)

type RpcSettingsResponse struct {
	RpcHost     string
	RpcUser     string
	RpcPassword string
}

func (h *HttpServer) RpcSettings(w http.ResponseWriter, r *http.Request) {
	resp := RpcSettingsResponse{h.config.RpcHost, h.config.RpcUser, h.config.RpcPassword}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(200)
	json.NewEncoder(w).Encode(resp)
}

func (h *HttpServer) ChangeRpcSettings(w http.ResponseWriter, r *http.Request) {
	decoder := json.NewDecoder(r.Body)
	var params RpcSettingsResponse
	err := decoder.Decode(&params)
	if err != nil {
		h.writeError(w, fmt.Errorf("Error decoding json: %s", err.Error()))
		return
	}

	err = h.config.SetRpcCredentials(params.RpcHost, params.RpcUser, params.RpcPassword)
	if err != nil {
		h.writeError(w, fmt.Errorf("Error changing RPC: %s", err.Error()))
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(200)
	json.NewEncoder(w).Encode(true)
}
