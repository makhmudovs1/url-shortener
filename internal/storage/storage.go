package storage

import (
	"context"
	"time"
)

type Storage interface {
	Create(ctx context.Context, longURL string, expireAt *time.Time) (id int64, err error)
	SetCode(ctx context.Context, id int64, code string) error
	GetByCode(ctx context.Context, code string) (string, error)
}
