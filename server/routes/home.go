package routes

import (
	"fmt"
	"net/http"
)

func Home(w http.ResponseWriter, r *http.Request) {
	fmt.Fprint(w, "Welcome to Vertcoin OpenAssets v0.1 beta")
}
