package handler

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/jackc/pgx/v5"
	"github.com/makhmudovs1/url-shortener/internal/shortener"
	"github.com/makhmudovs1/url-shortener/internal/storage"
	"log/slog"
	"net/http"
	"time"
)

type shortenRequest struct {
	URL      string     `json:"url"`
	ExpireAt *time.Time `json:"expire_at"`
}

type shortenResponse struct {
	ShortURL string `json:"short_url"`
}

func ShortenHandler(repo storage.Storage) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req shortenRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "invalid request", http.StatusBadRequest)
			return
		}
		id, err := repo.Create(r.Context(), req.URL, req.ExpireAt)
		if err != nil {
			slog.Error("create url failed", "err", err)
			http.Error(w, "db error", http.StatusInternalServerError)
			return
		}
		code := shortener.ShortenURL(id)

		if err := repo.SetCode(r.Context(), id, code); err != nil {
			slog.Error("SetCode failed", "err", err)
			http.Error(w, "db error", http.StatusInternalServerError)
			return
		}
		shortURL := fmt.Sprintf("http://%s/%s", r.Host, code)

		resp := shortenResponse{ShortURL: shortURL}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	}
}

func RedirectHandler(repo storage.Storage) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		code := r.URL.Path[1:]
		if code == "" {
			http.Error(w, "code is required", http.StatusBadRequest)
			return
		}

		longURL, err := repo.GetByCode(r.Context(), code)
		if err != nil {
			if errors.Is(err, pgx.ErrNoRows) {
				http.NotFound(w, r)
				return
			}
			slog.Error("GetByCode failed", "err", err)
			http.Error(w, "db error", http.StatusInternalServerError)
			return
		}

		http.Redirect(w, r, longURL, http.StatusFound)
	}
}
