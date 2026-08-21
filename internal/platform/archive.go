package platform

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"github.com/wyw14/cry-047/internal/domain"
)

type FileArchiveWriter struct {
	Root string
	mu   sync.Mutex
}

func (w *FileArchiveWriter) Write(ctx context.Context, bundle domain.ArchiveBundle) (string, error) {
	if err := ctx.Err(); err != nil {
		return "", err
	}
	if strings.TrimSpace(w.Root) == "" {
		return "", fmt.Errorf("archive root is empty")
	}
	w.mu.Lock()
	defer w.mu.Unlock()
	if err := os.MkdirAll(w.Root, 0o750); err != nil {
		return "", fmt.Errorf("create archive directory: %w", err)
	}
	data, err := json.MarshalIndent(bundle, "", "  ")
	if err != nil {
		return "", fmt.Errorf("encode archive: %w", err)
	}
	name := fmt.Sprintf("facility-%s-%s.json", bundle.Facility.ID, bundle.Checksum[:16])
	path := filepath.Join(w.Root, name)
	temporary := path + ".tmp"
	if err := os.WriteFile(temporary, data, 0o640); err != nil {
		return "", fmt.Errorf("write temporary archive: %w", err)
	}
	if err := os.Rename(temporary, path); err != nil {
		_ = os.Remove(temporary)
		return "", fmt.Errorf("publish archive: %w", err)
	}
	return path, nil
}
