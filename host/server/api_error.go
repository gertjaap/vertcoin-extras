package server

import (
	"encoding/json"
	"net/http"
)

type ErrorResponse struct {
	Error        bool
	ErrorDetails string
}

func (h *HttpServer) writeError(w http.ResponseWriter, err error) {
	resp := ErrorResponse{Error: true, ErrorDetails: err.Error()}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(500)
	json.NewEncoder(w).Encode(resp)
}
