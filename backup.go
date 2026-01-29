package sncli

import (
	"archive/zip"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"time"

	"github.com/jonhadfield/gosn-v2/cache"
	"github.com/jonhadfield/gosn-v2/common"
	"github.com/jonhadfield/gosn-v2/items"
	"golang.org/x/crypto/pbkdf2"
)

// BackupConfig holds backup configuration
type BackupConfig struct {
	Session       *cache.Session
	OutputFile    string
	Incremental   bool
	LastBackupTime string
	Encrypt       bool
	Password      string
	Debug         bool
}

// RestoreConfig holds restore configuration
type RestoreConfig struct {
	Session    *cache.Session
	InputFile  string
	DryRun     bool
	Password   string
	Debug      bool
}

// BackupManifest contains metadata about the backup
type BackupManifest struct {
	Timestamp   string `json:"timestamp"`
	ItemCounts  map[string]int `json:"item_counts"`
	Incremental bool   `json:"incremental"`
	Encrypted   bool   `json:"encrypted"`
	Version     string `json:"version"`
}

// BackupItem represents an item in the backup
type BackupItem struct {
	UUID      string `json:"uuid"`
	Type      string `json:"type"`
	Content   string `json:"content"`
	CreatedAt string `json:"created_at"`
	UpdatedAt string `json:"updated_at"`
}

// Run executes the backup
func (b *BackupConfig) Run() error {
	// Create zip file
	zipFile, err := os.Create(b.OutputFile)
	if err != nil {
		return fmt.Errorf("failed to create backup file: %w", err)
	}
	defer zipFile.Close()

	var writer io.Writer = zipFile
	var gcm cipher.AEAD

	// Set up encryption if requested
	if b.Encrypt {
		if b.Password == "" {
			return fmt.Errorf("password required for encrypted backup")
		}

		// Derive key from password
		key := pbkdf2.Key([]byte(b.Password), []byte("sn-cli-backup-salt"), 100000, 32, sha256.New)

		// Create AES cipher
		block, err := aes.NewCipher(key)
		if err != nil {
			return fmt.Errorf("failed to create cipher: %w", err)
		}

		gcm, err = cipher.NewGCM(block)
		if err != nil {
			return fmt.Errorf("failed to create GCM: %w", err)
		}
	}

	zipWriter := zip.NewWriter(writer)
	defer zipWriter.Close()

	// Create manifest
	manifest := BackupManifest{
		Timestamp:   time.Now().Format(time.RFC3339),
		ItemCounts:  make(map[string]int),
		Incremental: b.Incremental,
		Encrypted:   b.Encrypt,
		Version:     "1.0",
	}

	// Backup notes
	if err := b.backupNotes(zipWriter, &manifest, gcm); err != nil {
		return fmt.Errorf("failed to backup notes: %w", err)
	}

	// Backup tags
	if err := b.backupTags(zipWriter, &manifest, gcm); err != nil {
		return fmt.Errorf("failed to backup tags: %w", err)
	}

	// Write manifest
	manifestData, err := json.MarshalIndent(manifest, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal manifest: %w", err)
	}

	if err := b.writeZipFile(zipWriter, "manifest.json", manifestData, gcm); err != nil {
		return fmt.Errorf("failed to write manifest: %w", err)
	}

	return nil
}

// backupNotes backs up all notes
func (b *BackupConfig) backupNotes(zipWriter *zip.Writer, manifest *BackupManifest, gcm cipher.AEAD) error {
	noteFilter := items.Filter{
		Type: common.SNItemTypeNote,
	}

	getNoteConfig := GetNoteConfig{
		Session: b.Session,
		Filters: items.ItemFilters{
			MatchAny: false,
			Filters:  []items.Filter{noteFilter},
		},
		Debug: b.Debug,
	}

	rawNotes, err := getNoteConfig.Run()
	if err != nil {
		return err
	}

	var notesToBackup []BackupItem

	for _, item := range rawNotes {
		note := item.(*items.Note)

		// Skip if incremental and not modified since last backup
		if b.Incremental && b.LastBackupTime != "" {
			if note.UpdatedAt < b.LastBackupTime {
				continue
			}
		}

		// Marshal note content
		contentData, err := json.Marshal(note.Content)
		if err != nil {
			return fmt.Errorf("failed to marshal note content: %w", err)
		}

		notesToBackup = append(notesToBackup, BackupItem{
			UUID:      note.UUID,
			Type:      common.SNItemTypeNote,
			Content:   string(contentData),
			CreatedAt: note.CreatedAt,
			UpdatedAt: note.UpdatedAt,
		})
	}

	// Write notes to zip
	notesData, err := json.MarshalIndent(notesToBackup, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal notes: %w", err)
	}

	if err := b.writeZipFile(zipWriter, "notes.json", notesData, gcm); err != nil {
		return err
	}

	manifest.ItemCounts["notes"] = len(notesToBackup)
	return nil
}

// backupTags backs up all tags
func (b *BackupConfig) backupTags(zipWriter *zip.Writer, manifest *BackupManifest, gcm cipher.AEAD) error {
	tagFilter := items.Filter{
		Type: common.SNItemTypeTag,
	}

	getTagConfig := GetTagConfig{
		Session: b.Session,
		Filters: items.ItemFilters{
			MatchAny: false,
			Filters:  []items.Filter{tagFilter},
		},
		Debug: b.Debug,
	}

	rawTags, err := getTagConfig.Run()
	if err != nil {
		return err
	}

	var tagsToBackup []BackupItem

	for _, item := range rawTags {
		tag := item.(*items.Tag)

		// Skip if incremental and not modified since last backup
		if b.Incremental && b.LastBackupTime != "" {
			if tag.UpdatedAt < b.LastBackupTime {
				continue
			}
		}

		// Marshal tag content
		contentData, err := json.Marshal(tag.Content)
		if err != nil {
			return fmt.Errorf("failed to marshal tag content: %w", err)
		}

		tagsToBackup = append(tagsToBackup, BackupItem{
			UUID:      tag.UUID,
			Type:      common.SNItemTypeTag,
			Content:   string(contentData),
			CreatedAt: tag.CreatedAt,
			UpdatedAt: tag.UpdatedAt,
		})
	}

	// Write tags to zip
	tagsData, err := json.MarshalIndent(tagsToBackup, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal tags: %w", err)
	}

	if err := b.writeZipFile(zipWriter, "tags.json", tagsData, gcm); err != nil {
		return err
	}

	manifest.ItemCounts["tags"] = len(tagsToBackup)
	return nil
}

// writeZipFile writes data to zip file with optional encryption
func (b *BackupConfig) writeZipFile(zipWriter *zip.Writer, filename string, data []byte, gcm cipher.AEAD) error {
	fileWriter, err := zipWriter.Create(filename)
	if err != nil {
		return fmt.Errorf("failed to create zip entry: %w", err)
	}

	if b.Encrypt && gcm != nil {
		// Generate nonce
		nonce := make([]byte, gcm.NonceSize())
		if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
			return fmt.Errorf("failed to generate nonce: %w", err)
		}

		// Encrypt data
		encrypted := gcm.Seal(nonce, nonce, data, nil)

		if _, err := fileWriter.Write(encrypted); err != nil {
			return fmt.Errorf("failed to write encrypted data: %w", err)
		}
	} else {
		if _, err := fileWriter.Write(data); err != nil {
			return fmt.Errorf("failed to write data: %w", err)
		}
	}

	return nil
}

// Run executes the restore
func (r *RestoreConfig) Run() (RestoreResult, error) {
	result := RestoreResult{
		DryRun: r.DryRun,
	}

	// Open zip file
	zipReader, err := zip.OpenReader(r.InputFile)
	if err != nil {
		return result, fmt.Errorf("failed to open backup file: %w", err)
	}
	defer zipReader.Close()

	// Read manifest
	manifestData, encrypted, err := r.readZipFile(&zipReader.Reader, "manifest.json")
	if err != nil {
		return result, fmt.Errorf("failed to read manifest: %w", err)
	}

	var manifest BackupManifest
	if err := json.Unmarshal(manifestData, &manifest); err != nil {
		return result, fmt.Errorf("failed to parse manifest: %w", err)
	}

	result.Manifest = manifest

	// Verify encryption
	if manifest.Encrypted && r.Password == "" {
		return result, fmt.Errorf("password required for encrypted backup")
	}

	var gcm cipher.AEAD
	if manifest.Encrypted {
		// Derive key from password
		key := pbkdf2.Key([]byte(r.Password), []byte("sn-cli-backup-salt"), 100000, 32, sha256.New)

		// Create AES cipher
		block, err := aes.NewCipher(key)
		if err != nil {
			return result, fmt.Errorf("failed to create cipher: %w", err)
		}

		gcm, err = cipher.NewGCM(block)
		if err != nil {
			return result, fmt.Errorf("failed to create GCM: %w", err)
		}
	}

	// Read notes
	notesData, _, err := r.readZipFile(&zipReader.Reader, "notes.json")
	if err != nil {
		return result, fmt.Errorf("failed to read notes: %w", err)
	}

	// Decrypt if needed
	if encrypted && gcm != nil {
		notesData, err = r.decrypt(notesData, gcm)
		if err != nil {
			return result, fmt.Errorf("failed to decrypt notes: %w", err)
		}
	}

	var notes []BackupItem
	if err := json.Unmarshal(notesData, &notes); err != nil {
		return result, fmt.Errorf("failed to parse notes: %w", err)
	}

	result.NotesCount = len(notes)

	// Read tags
	tagsData, _, err := r.readZipFile(&zipReader.Reader, "tags.json")
	if err != nil {
		return result, fmt.Errorf("failed to read tags: %w", err)
	}

	// Decrypt if needed
	if encrypted && gcm != nil {
		tagsData, err = r.decrypt(tagsData, gcm)
		if err != nil {
			return result, fmt.Errorf("failed to decrypt tags: %w", err)
		}
	}

	var tags []BackupItem
	if err := json.Unmarshal(tagsData, &tags); err != nil {
		return result, fmt.Errorf("failed to parse tags: %w", err)
	}

	result.TagsCount = len(tags)

	// TODO: Implement actual restore logic (not in dry-run mode)
	// This would involve creating notes and tags in the session

	return result, nil
}

// readZipFile reads a file from the zip archive
func (r *RestoreConfig) readZipFile(zipReader *zip.Reader, filename string) ([]byte, bool, error) {
	for _, file := range zipReader.File {
		if file.Name == filename {
			rc, err := file.Open()
			if err != nil {
				return nil, false, err
			}
			defer rc.Close()

			data, err := io.ReadAll(rc)
			if err != nil {
				return nil, false, err
			}

			// Check if encrypted (has nonce prefix)
			encrypted := len(data) > 12 // Minimal size for nonce + ciphertext
			return data, encrypted, nil
		}
	}

	return nil, false, fmt.Errorf("file %s not found in backup", filename)
}

// decrypt decrypts data using GCM
func (r *RestoreConfig) decrypt(data []byte, gcm cipher.AEAD) ([]byte, error) {
	nonceSize := gcm.NonceSize()
	if len(data) < nonceSize {
		return nil, fmt.Errorf("ciphertext too short")
	}

	nonce, ciphertext := data[:nonceSize], data[nonceSize:]
	plaintext, err := gcm.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return nil, fmt.Errorf("decryption failed: %w", err)
	}

	return plaintext, nil
}

// RestoreResult contains the result of a restore operation
type RestoreResult struct {
	DryRun     bool
	Manifest   BackupManifest
	NotesCount int
	TagsCount  int
}

// GetBackupInfo reads backup metadata without restoring
func GetBackupInfo(filename string, password string) (BackupManifest, error) {
	zipReader, err := zip.OpenReader(filename)
	if err != nil {
		return BackupManifest{}, fmt.Errorf("failed to open backup file: %w", err)
	}
	defer zipReader.Close()

	// Find manifest file
	var manifestFile *zip.File
	for _, file := range zipReader.File {
		if file.Name == "manifest.json" {
			manifestFile = file
			break
		}
	}

	if manifestFile == nil {
		return BackupManifest{}, fmt.Errorf("manifest not found in backup")
	}

	rc, err := manifestFile.Open()
	if err != nil {
		return BackupManifest{}, err
	}
	defer rc.Close()

	data, err := io.ReadAll(rc)
	if err != nil {
		return BackupManifest{}, err
	}

	var manifest BackupManifest
	if err := json.Unmarshal(data, &manifest); err != nil {
		return BackupManifest{}, fmt.Errorf("failed to parse manifest: %w", err)
	}

	return manifest, nil
}
