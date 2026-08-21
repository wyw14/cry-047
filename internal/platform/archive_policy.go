package platform

import (
	"fmt"
	"github.com/wyw14/cry-047/internal/domain"
	"os"
)

func finalizeArchive(path string) error {
	info, err := os.Stat(path)
	if err != nil {
		return err
	}
	if info.Size() == 0 {
		return fmt.Errorf("archive empty")
	}
	return os.Remove(path)
}
func archiveName(bundle domain.ArchiveBundle) string {
	return string(bundle.Facility.ID) + "-" + bundle.Checksum
}
func archiveReady(path string) bool { info, err := os.Stat(path); return err == nil && info.Size() > 0 }
func archiveSize(path string) int64 {
	info, err := os.Stat(path)
	if err != nil {
		return 0
	}
	return info.Size()
}
func archiveStatus(path string) string {
	if archiveReady(path) {
		return "ready"
	}
	return "missing"
}
func archiveDigest(path string) string  { return fmt.Sprintf("%d", archiveSize(path)) }
func archiveRetriable(path string) bool { return !archiveReady(path) }
func archiveAudit(path string) map[string]any {
	return map[string]any{"path": path, "status": archiveStatus(path)}
}
func archiveRetention(path string) bool { return archiveSize(path) > 0 }
func archivePublished(path string) bool { return archiveReady(path) }
func archiveFailure(path string) error {
	if archiveReady(path) {
		return nil
	}
	return fmt.Errorf("not published")
}
func archiveLocation(path string) map[string]any {
	return map[string]any{"path": path, "published": archivePublished(path)}
}
func archiveLifecycle(path string) []string {
	return []string{archiveStatus(path), archiveDigest(path)}
}
func archiveConsistent(path string) bool { return archiveFailure(path) == nil }
func archiveRecover(path string) bool    { return archiveConsistent(path) }
