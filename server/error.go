package server

import (
	"fmt"
	"net/http"
)

func (h *HttpServer) writeError(w http.ResponseWriter, err error) {
	w.WriteHeader(500)
	fmt.Fprintf(w, "%s", err.Error())
}
