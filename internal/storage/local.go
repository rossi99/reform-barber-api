package storage

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
)

// LocalStore writes files to disk. Used in development.
// Files are served by the HTTP server at /uploads/*.
type LocalStore struct {
	dir     string // absolute path to uploads directory
	baseURL string // e.g. http://localhost:8080/uploads
}

func NewLocalStore(dir, baseURL string) *LocalStore {
	return &LocalStore{dir: dir, baseURL: baseURL}
}

func (s *LocalStore) Upload(_ context.Context, key string, r io.Reader, _ string) (string, error) {
	dest := filepath.Join(s.dir, filepath.FromSlash(key))
	if err := os.MkdirAll(filepath.Dir(dest), 0o755); err != nil {
		return "", err
	}
	f, err := os.Create(dest)
	if err != nil {
		return "", err
	}
	defer f.Close()
	if _, err = io.Copy(f, r); err != nil {
		return "", err
	}
	return fmt.Sprintf("%s/%s", s.baseURL, key), nil
}

func (s *LocalStore) Delete(_ context.Context, key string) error {
	return os.Remove(filepath.Join(s.dir, filepath.FromSlash(key)))
}
