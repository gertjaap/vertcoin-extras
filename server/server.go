package server

import (
	"net/http"

	"github.com/gertjaap/vertcoin-openassets/server/routes"
)

func RunHttpServer() error {
	http.HandleFunc("/", routes.Home)
	return http.ListenAndServe(":8080", nil)
}
