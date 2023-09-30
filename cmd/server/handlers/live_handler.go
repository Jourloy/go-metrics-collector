package handlers

import (
	"fmt"
	"net/http"
)

func live(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("Live"))
}

func RegisterLiveHandler(mux *http.ServeMux) {
	mux.HandleFunc("/live", live)
	fmt.Println(`Live handler registered on /live`)
}
