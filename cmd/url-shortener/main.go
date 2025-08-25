package main

import (
	"context"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/makhmudovs1/url-shortener/internal/handler"
	"github.com/makhmudovs1/url-shortener/internal/shortener"
	"github.com/makhmudovs1/url-shortener/internal/storage"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	slog.SetDefault(logger)

	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		slog.Error("DATABASE_URL is empty")
		os.Exit(1)
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	pool, err := pgxpool.New(ctx, dsn)
	if err != nil {
		slog.Error("pgxpool.New failed", "err", err)
		os.Exit(1)
	}
	err = pool.Ping(ctx)
	if err != nil {
		slog.Error("db ping failed", "err", err)
		os.Exit(1)
	}
	defer pool.Close()
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	repo := storage.New(pool)
	id, _ := repo.Create(ctx, "https://example.com", nil)
	code := shortener.ShortenURL(id)
	_ = repo.SetCode(ctx, id, code)

	mux := http.NewServeMux()
	mux.HandleFunc("/healthz", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("ok"))
	})
	mux.HandleFunc("/readyz", func(w http.ResponseWriter, r *http.Request) {
		ctx, cancel := context.WithTimeout(r.Context(), 500*time.Millisecond)
		defer cancel()
		if _, err := pool.Exec(ctx, "SELECT 1"); err != nil {
			slog.Error("readyz: db not ready", "err", err)
			w.WriteHeader(http.StatusServiceUnavailable)
			w.Write([]byte("db not ready"))
			return
		}
		w.Write([]byte("ready"))
	})
	mux.HandleFunc("/shorten", handler.ShortenHandler(repo))
	mux.HandleFunc("/", handler.RedirectHandler(repo))

	srv := &http.Server{
		Addr:         ":" + port,
		Handler:      mux,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  60 * time.Second,
	}
	// Running server in another goroutine
	go func() {
		slog.Info("starting http server", "port", port)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			slog.Error("server failed", "err", err)
			os.Exit(1)
		}
	}()
	// chanel for signals
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)
	<-stop
	slog.Info("shutting down server...")

	shCtx, shCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer shCancel()

	if err := srv.Shutdown(shCtx); err != nil {
		slog.Error("graceful shutdown failed", "err", err)
	} else {
		slog.Info("server stopped")
	}
}
