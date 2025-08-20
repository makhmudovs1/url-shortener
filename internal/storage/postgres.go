package storage

import (
	"context"
	"fmt"
	"github.com/jackc/pgx/v5/pgxpool"
	"time"
)

type Repo struct {
	db *pgxpool.Pool
}

func New(db *pgxpool.Pool) *Repo {
	return &Repo{
		db: db,
	}
}

func (r *Repo) Create(ctx context.Context, longURL string, expireAt *time.Time) (id int64, err error) {
	ctx, cancel := context.WithTimeout(ctx, 2*time.Second)
	defer cancel()

	var ID int64
	err = r.db.QueryRow(ctx,
		"INSERT INTO urls (long_url, expire_at) VALUES ($1, $2) RETURNING id", longURL, expireAt).Scan(&ID)
	if err != nil {
		return 0, err
	}
	return ID, nil
}

func (r *Repo) SetCode(ctx context.Context, id int64, code string) error {
	ctx, cancel := context.WithTimeout(ctx, 2*time.Second)
	defer cancel()

	ct, err := r.db.Exec(ctx, `UPDATE urls SET code = $1 WHERE id = $2`, code, id)
	if err != nil {
		return err
	}
	if ct.RowsAffected() != 1 {
		return fmt.Errorf("no rows updated for id=%d", id)
	}
	return nil
}

func (r *Repo) GetByCode(ctx context.Context, code string) (string, error) {
	var longURL string
	err := r.db.QueryRow(ctx,
		`SELECT long_url FROM urls WHERE code = $1 LIMIT 1`, code,
	).Scan(&longURL)
	if err != nil {
		return "", err
	}
	return longURL, nil
}
