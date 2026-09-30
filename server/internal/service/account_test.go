package service

import (
	"os"
	"path/filepath"
	"testing"
)

func TestRemoveStoredObjectsStaysInsideUploadRoot(t *testing.T) {
	base := t.TempDir()
	root := filepath.Join(base, "uploads")
	stored := filepath.Join(root, "public", "records", "photo.jpg")
	outside := filepath.Join(base, "outside.txt")
	if err := os.MkdirAll(filepath.Dir(stored), 0o750); err != nil {
		t.Fatalf("MkdirAll() error = %v", err)
	}
	if err := os.WriteFile(stored, []byte("image"), 0o600); err != nil {
		t.Fatalf("WriteFile(stored) error = %v", err)
	}
	if err := os.WriteFile(outside, []byte("keep"), 0o600); err != nil {
		t.Fatalf("WriteFile(outside) error = %v", err)
	}

	service := &AccountService{uploadRoot: root}
	service.removeStoredObjects([]string{"public/records/photo.jpg", "../outside.txt"})

	if _, err := os.Stat(stored); !os.IsNotExist(err) {
		t.Fatalf("stored file still exists or stat failed unexpectedly: %v", err)
	}
	if _, err := os.Stat(outside); err != nil {
		t.Fatalf("outside file was removed: %v", err)
	}
}
