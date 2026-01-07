package storage

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"
)

func TestValidateFilename(t *testing.T) {
	svc, _ := NewService("12345678901234567890123456789012", os.TempDir())

	tests := []struct {
		name     string
		filename string
		wantErr  bool
	}{
		{"valid simple", "file.txt", false},
		{"valid with dash", "my-file.txt", false},
		{"valid with underscore", "my_file.txt", false},
		{"empty filename", "", true},
		{"path traversal", "../etc/passwd", true},
		{"path traversal 2", "..\\windows\\system32", true},
		{"has forward slash", "some/file.txt", true},
		{"has backslash", "some\\file.txt", true},
		{"just dots", "..", true},
		{"single dot", ".", true},
		{"special chars", "file<>.txt", true},
		{"valid numbers", "file123.txt", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := svc.validateFilename(tt.filename)
			if (err != nil) != tt.wantErr {
				t.Errorf("validateFilename(%q) error = %v, wantErr %v", tt.filename, err, tt.wantErr)
			}
		})
	}
}

func TestSaveAndLoadFile(t *testing.T) {
	// Create temp directory for test
	tmpDir, err := os.MkdirTemp("", "storage_test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	svc, err := NewService("12345678901234567890123456789012", tmpDir)
	if err != nil {
		t.Fatalf("Failed to create service: %v", err)
	}

	testData := []byte("Hello, this is secret data that should be encrypted!")
	filename := "test.txt"

	// Test saving
	if err := svc.SaveFile(filename, bytes.NewReader(testData)); err != nil {
		t.Fatalf("SaveFile failed: %v", err)
	}

	// Verify the encrypted file exists
	encPath := filepath.Join(tmpDir, filename+".enc")
	if _, err := os.Stat(encPath); os.IsNotExist(err) {
		t.Fatalf("Encrypted file was not created")
	}

	// Verify the checksum file exists
	checksumPath := filepath.Join(tmpDir, filename+".sha256")
	if _, err := os.Stat(checksumPath); os.IsNotExist(err) {
		t.Fatalf("Checksum file was not created")
	}

	// Verify stored data is encrypted (not plaintext)
	storedData, _ := os.ReadFile(encPath)
	if string(storedData) == string(testData) {
		t.Error("Data was stored in plaintext, not encrypted!")
	}

	// Test loading
	loadedData, err := svc.LoadFile(filename)
	if err != nil {
		t.Fatalf("LoadFile failed: %v", err)
	}

	// Verify decrypted data matches original
	if string(loadedData) != string(testData) {
		t.Errorf("Loaded data does not match original.\nGot: %s\nWant: %s", loadedData, testData)
	}
}

func TestLoadNonExistentFile(t *testing.T) {
	tmpDir, _ := os.MkdirTemp("", "storage_test")
	defer os.RemoveAll(tmpDir)

	svc, _ := NewService("12345678901234567890123456789012", tmpDir)

	_, err := svc.LoadFile("nonexistent.txt")
	if err == nil {
		t.Error("Expected error for non-existent file, got nil")
	}
}

func TestFileExists(t *testing.T) {
	tmpDir, _ := os.MkdirTemp("", "storage_test")
	defer os.RemoveAll(tmpDir)

	svc, _ := NewService("12345678901234567890123456789012", tmpDir)

	// Before saving
	if svc.FileExists("myfile.txt") {
		t.Error("FileExists returned true for non-existent file")
	}

	// Save a file
	svc.SaveFile("myfile.txt", bytes.NewReader([]byte("test")))

	// After saving
	if !svc.FileExists("myfile.txt") {
		t.Error("FileExists returned false for existing file")
	}
}
