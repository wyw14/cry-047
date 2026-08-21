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
	// finalizeArchive confirms the published file is durable and readable
	// before we report its location. On failure we must not advertise a
	// broken location (谎报成功): remove the unusable artifact and surface the
	// error so the caller knows the publish did not land.
	if err := finalizeArchive(path); err != nil {
		_ = os.Remove(path)
		return "", fmt.Errorf("finalize archive: %w", err)
	}
	return path, nil
}
