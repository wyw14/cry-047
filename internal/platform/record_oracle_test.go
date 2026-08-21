package platform

import (
	"context"
	"github.com/wyw14/cry-047/internal/domain"
	"os"
	"testing"
	"time"
)

func TestArchiveWriteKeepsPublishedFileReadable(t *testing.T) {
	writer := &FileArchiveWriter{Root: t.TempDir()}
	bundle := domain.ArchiveBundle{Facility: domain.Facility{ID: "facility-file"}, Checksum: "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa", ExportedAt: time.Now()}
	path, err := writer.Write(context.Background(), bundle)
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
}
