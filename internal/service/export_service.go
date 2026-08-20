package service

import (
	"archive/zip"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/wyw14/cry-046/internal/domain"
)

type LocalPackageWriter struct{ Root string }

func (w LocalPackageWriter) WritePackage(_ context.Context, p domain.Palette, assets []domain.Asset, format string) (string, error) {
	if err := os.MkdirAll(w.Root, 0750); err != nil {
		return "", err
	}
	safe := strings.ReplaceAll(p.ID, "/", "_")
	path := filepath.Join(w.Root, safe+"."+format)
	var data []byte
	var err error
	switch format {
	case "json":
		data, err = json.MarshalIndent(map[string]any{"palette": p, "assets": assets}, "", "  ")
	case "csv":
		data = []byte("name,hex,source\n")
		for _, e := range p.Entries {
			data = append(data, []byte(fmt.Sprintf("%s,%s,%s\n", e.Name, e.Hex, e.Source))...)
		}
	case "zip":
		var b bytes.Buffer
		z := zip.NewWriter(&b)
		f, _ := z.Create("palette.json")
		enc, _ := json.Marshal(p)
		_, _ = f.Write(enc)
		_ = z.Close()
		data = b.Bytes()
	default:
		return "", fmt.Errorf("format %s unsupported", format)
	}
	if err != nil {
		return "", err
	}
	if err := os.WriteFile(path, data, 0600); err != nil {
		return "", err
	}
	return path, nil
}
func CopyTo(w io.Writer, r io.Reader) error { _, err := io.Copy(w, r); return err }
