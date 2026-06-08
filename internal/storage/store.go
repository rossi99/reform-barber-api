package storage

import (
	"context"
	"io"
)

type Store interface {
	Upload(ctx context.Context, key string, r io.Reader, contentType string) (publicURL string, err error)
	Delete(ctx context.Context, key string) error
}
