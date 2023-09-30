package main

import (
	"fmt"
	"net/http"

	"github.com/Jourloy/go-metrics-collector/cmd/server/handlers"
)

func liveHandler(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("Live"))
}

func main() {
	// Prepare for .env
	port := `8080`

	// Initiate handlers
	mux := http.NewServeMux()
	handlers.RegisterLiveHandler(mux)
	handlers.RegisterCollectorHandler(mux)

	fmt.Println(`Server started on port`, port)

	// Starting server
	if err := http.ListenAndServe(fmt.Sprintf(`:%s`, port), mux); err != nil {
		if err == http.ErrServerClosed {
			fmt.Println(`Server closed`)
		} else {
			panic(err)
		}
	}
}
