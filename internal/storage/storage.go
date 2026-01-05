package storage

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

// Service handles encrypted file operations
type Service struct {
	key     []byte
	dataDir string
	gcm     cipher.AEAD
}

// NewService creates a new storage service with the given encryption key
func NewService(key string, dataDir string) (*Service, error) {
	keyBytes := []byte(key)

	// Create the cipher block
	block, err := aes.NewCipher(keyBytes)
	if err != nil {
		return nil, fmt.Errorf("failed to create cipher: %w", err)
	}

	// Create GCM wrapper for authenticated encryption
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, fmt.Errorf("failed to create GCM: %w", err)
	}

	return &Service{
		key:     keyBytes,
		dataDir: dataDir,
		gcm:     gcm,
	}, nil
}

// validateFilename checks for path traversal and other dangerous patterns
func (s *Service) validateFilename(filename string) error {
	// Check for empty filename
	if filename == "" {
		return errors.New("filename cannot be empty")
	}

	// Check for path traversal attempts
	if strings.Contains(filename, "..") {
		return errors.New("filename contains path traversal attempt")
	}

	// Check for absolute paths or directory separators
	if strings.ContainsAny(filename, "/\\") {
		return errors.New("filename cannot contain path separators")
	}

	// Only allow alphanumeric, dash, underscore, and dots
	validPattern := regexp.MustCompile(`^[a-zA-Z0-9_\-\.]+$`)
	if !validPattern.MatchString(filename) {
		return errors.New("filename contains invalid characters")
	}

	// Make sure filename isnt just dots
	if filename == "." || filename == ".." {
		return errors.New("invalid filename")
	}

	return nil
}

// SaveFile encrypts and saves a file to disk
func (s *Service) SaveFile(filename string, content io.Reader) error {
	// Validate the filename first
	if err := s.validateFilename(filename); err != nil {
		return fmt.Errorf("invalid filename: %w", err)
	}

	// 1. Setup Hasher
	hasher := sha256.New()

	// 2. "Tee" the stream: Read from content -> Write to Hasher AND Encryption
	// This allows us to hash the PLAINTEXT while reading
	teeReader := io.TeeReader(content, hasher)

	// Read all data from the tee reader to get the byte slice for encryption
	data, err := io.ReadAll(teeReader)
	if err != nil {
		return fmt.Errorf("failed to read file content: %w", err)
	}

	// Generate random nonce for GCM
	nonce := make([]byte, s.gcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return fmt.Errorf("failed to generate nonce: %w", err)
	}

	// Encrypt the data
	// The nonce is prepended to the ciphertext
	encrypted := s.gcm.Seal(nonce, nonce, data, nil)

	// Build the file path
	filePath := filepath.Join(s.dataDir, filename+".enc")

	// Write to file
	if err := os.WriteFile(filePath, encrypted, 0600); err != nil {
		return fmt.Errorf("failed to write encrypted file: %w", err)
	}

	// 4. Get the Checksum
	checksum := hex.EncodeToString(hasher.Sum(nil))

	// 5. Save Checksum (e.g., filename.sha256)
	checksumPath := filepath.Join(s.dataDir, filename+".sha256")
	if err := os.WriteFile(checksumPath, []byte(checksum), 0644); err != nil {
		return fmt.Errorf("failed to write checksum file: %w", err)
	}

	return nil
}

// LoadFile reads and decrypts a file from disk
func (s *Service) LoadFile(filename string) ([]byte, error) {
	// Validate filename
	if err := s.validateFilename(filename); err != nil {
		return nil, fmt.Errorf("invalid filename: %w", err)
	}

	// Build the file path
	filePath := filepath.Join(s.dataDir, filename+".enc")

	// Read the encrypted file
	encrypted, err := os.ReadFile(filePath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, errors.New("file not found")
		}
		return nil, fmt.Errorf("failed to read file: %w", err)
	}

	// Check if we have at least the nonce
	nonceSize := s.gcm.NonceSize()
	if len(encrypted) < nonceSize {
		return nil, errors.New("encrypted file is too small")
	}

	// Split nonce and ciphertext
	nonce := encrypted[:nonceSize]
	ciphertext := encrypted[nonceSize:]

	// Decrypt
	plaintext, err := s.gcm.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to decrypt file: %w", err)
	}

	// Verify Checksum
	checksumPath := filepath.Join(s.dataDir, filename+".sha256")
	storedChecksumBytes, err := os.ReadFile(checksumPath)
	if err != nil {
		// Log/Warning? Fail for security mode.
		if os.IsNotExist(err) {
			return nil, errors.New("integrity check failed: checksum file missing")
		}
		return nil, fmt.Errorf("failed to read checksum: %w", err)
	}
	storedChecksum := string(storedChecksumBytes)

	// Hash the decrypted content
	hasher := sha256.New()
	hasher.Write(plaintext)
	calculatedChecksum := hex.EncodeToString(hasher.Sum(nil))

	if calculatedChecksum != storedChecksum {
		return nil, fmt.Errorf("integrity check failed: hash mismatch")
	}

	return plaintext, nil
}

// FileExists checks if an encrypted file exists
func (s *Service) FileExists(filename string) bool {
	if err := s.validateFilename(filename); err != nil {
		return false
	}
	filePath := filepath.Join(s.dataDir, filename+".enc")
	_, err := os.Stat(filePath)
	return err == nil
}
