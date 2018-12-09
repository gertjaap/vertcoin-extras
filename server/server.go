package server

import (
	"net/http"

	"github.com/gertjaap/vertcoin-openassets/wallet"
	"github.com/gorilla/mux"
)

type HttpServer struct {
	wallet *wallet.Wallet
}

func NewHttpServer(w *wallet.Wallet) *HttpServer {
	h := new(HttpServer)
	h.wallet = w
	return h
}

func (h *HttpServer) Run() error {
	r := mux.NewRouter()

	r.HandleFunc("/", h.Home)
	r.HandleFunc("/api/newAsset", h.NewAsset)
	r.HandleFunc("/api/assets", h.Assets)
	r.HandleFunc("/api/assetBalance/{assetID}", h.AssetBalance)
	r.HandleFunc("/api/balance", h.Balance)
	http.Handle("/", r)
	return http.ListenAndServe(":8080", nil)
}
