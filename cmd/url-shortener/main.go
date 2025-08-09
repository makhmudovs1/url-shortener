package main

import (
	"fmt"
	"log/slog"
	"net/http"
	"os"
)

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	slog.SetDefault(logger)

	http.HandleFunc("/healthz", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, "OK")
	})
	addr := ":" + port
	slog.Info("Server is running on port", port)
	if err := http.ListenAndServe(addr, nil); err != nil {
		slog.Error("server failed", "err", err)
	}
}
