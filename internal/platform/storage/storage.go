package storage

import (
	"errors"
	"io"
	"mime"
	"os"
	"path/filepath"
	"strings"
)

type Local struct {
	Root     string
	MaxBytes int64
	Allowed  map[string]bool
}

func (s Local) Save(name string, r io.Reader, size int64) (string, error) {
	if size <= 0 || size > s.MaxBytes {
		return "", errors.New("invalid attachment size")
	}
	if filepath.Base(name) != name || strings.ContainsAny(name, `/\\`) {
		return "", errors.New("unsafe attachment name")
	}
	typ := mime.TypeByExtension(filepath.Ext(name))
	if !s.Allowed[typ] {
		return "", errors.New("attachment type not allowed")
	}
	if err := os.MkdirAll(s.Root, 0750); err != nil {
		return "", err
	}
	path := filepath.Join(s.Root, name)
	f, err := os.OpenFile(path, os.O_CREATE|os.O_WRONLY|os.O_EXCL, 0600)
	if err != nil {
		return "", err
	}
	defer f.Close()
	if _, err = io.CopyN(f, r, size); err != nil {
		return "", err
	}
	return path, nil
}
