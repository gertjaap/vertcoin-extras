package server

import (
	"fmt"
	"net/http"

	"github.com/gertjaap/vertcoin-openassets/config"
	"github.com/gertjaap/vertcoin-openassets/wallet"
	"github.com/gobuffalo/packr"
	"github.com/gorilla/mux"
	"github.com/rs/cors"
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
	box := packr.NewBox("./static")

	r.HandleFunc("/api/addresses", h.Addresses)
	r.HandleFunc("/api/newAsset", h.NewAsset)
	r.HandleFunc("/api/network", h.Network)
	r.HandleFunc("/api/transferAsset", h.TransferAsset)
	r.HandleFunc("/api/assets/all", h.AllAssets)
	r.HandleFunc("/api/assets/follow/{assetID}", h.FollowAsset)
	r.HandleFunc("/api/assets/unfollow/{assetID}", h.UnfollowAsset)

	r.HandleFunc("/api/assets/mine", h.MyAssets)
	r.HandleFunc("/api/assetBalance/{assetID}", h.AssetBalance)
	r.HandleFunc("/api/balance", h.Balance)

	r.PathPrefix("/").Handler(http.FileServer(box))

	if h.config.Cors {
		http.Handle("/", cors.Default().Handler(r))
	} else {
		http.Handle("/", r)
	}

	return http.ListenAndServe(fmt.Sprintf(":%d", h.config.Port), nil)
}
