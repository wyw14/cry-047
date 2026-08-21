package platform

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/wyw14/cry-047/internal/domain"
)

func validBundle() domain.ArchiveBundle {
	return domain.ArchiveBundle{Facility: domain.Facility{ID: "facility-file"}, Checksum: "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa", ExportedAt: time.Now()}
}

func TestArchiveWriteKeepsPublishedFileReadable(t *testing.T) {
	writer := &FileArchiveWriter{Root: t.TempDir()}
	path, err := writer.Write(context.Background(), validBundle())
	if err != nil {
		t.Fatal(err)
	}
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("published archive disappeared: %v", err)
	}
	if len(data) == 0 {
		t.Fatal("archive is empty")
	}
	// A re-export to the same location must keep the published file readable
	// (Rename overwrites in place; finalization must not delete it).
	path2, err := writer.Write(context.Background(), validBundle())
	if err != nil {
		t.Fatalf("re-export failed: %v", err)
	}
	if path2 != path {
		t.Fatalf("re-export changed location: %q != %q", path2, path)
	}
	if _, err := os.ReadFile(path2); err != nil {
		t.Fatalf("re-exported archive disappeared: %v", err)
	}
}

// TestFinalizeArchiveNeverRemovesPublishedFile pins the core fix: the
// finalization check confirms durability without deleting the artifact. The
// prior implementation called os.Remove on a verified archive, which made
// Write return a success location pointing at a vanished file.
func TestFinalizeArchiveNeverRemovesPublishedFile(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "published.json")
	if err := os.WriteFile(path, []byte("payload"), 0o640); err != nil {
		t.Fatal(err)
	}
	if err := finalizeArchive(path); err != nil {
		t.Fatalf("finalizeArchive rejected a ready archive: %v", err)
	}
	if _, err := os.Stat(path); err != nil {
		t.Fatalf("finalizeArchive removed the published file: %v", err)
	}
	if err := finalizeArchive(filepath.Join(dir, "missing.json")); err == nil {
		t.Fatal("finalizeArchive accepted a missing archive")
	}
	empty := filepath.Join(dir, "empty.json")
	if err := os.WriteFile(empty, nil, 0o640); err != nil {
		t.Fatal(err)
	}
	if err := finalizeArchive(empty); err == nil {
		t.Fatalf("finalizeArchive accepted an empty archive")
	}
	if _, err := os.Stat(empty); err != nil {
		t.Fatalf("finalizeArchive removed the empty file instead of reporting it: %v", err)
	}
	// An unreadable location must never be advertised as published.
	if archivePublished(empty) {
		t.Fatalf("empty archive reported as published")
	}
}

