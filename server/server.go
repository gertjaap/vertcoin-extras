package server

import (
	"fmt"
	"net/http"

	"github.com/gertjaap/vertcoin-openassets/config"
	"github.com/gertjaap/vertcoin-openassets/wallet"
	"github.com/gorilla/mux"
)

type HttpServer struct {
	wallet *wallet.Wallet
	config *config.Config
}

func NewHttpServer(w *wallet.Wallet, c *config.Config) *HttpServer {
	h := new(HttpServer)
	h.wallet = w
	h.config = c
	return h
}

func (h *HttpServer) Run() error {
	r := mux.NewRouter()

	r.HandleFunc("/", h.Home)
	r.HandleFunc("/api/addresses", h.Addresses)
	r.HandleFunc("/api/newAsset", h.NewAsset)
	r.HandleFunc("/api/transferAsset", h.TransferAsset)
	r.HandleFunc("/api/assets", h.Assets)
	r.HandleFunc("/api/assetBalance/{assetID}", h.AssetBalance)
	r.HandleFunc("/api/balance", h.Balance)
	http.Handle("/", r)

	return http.ListenAndServe(fmt.Sprintf(":%d", h.config.Port), nil)
}
