package sncli

import (
	"archive/zip"
	"encoding/base64"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/jonhadfield/gosn-v2/common"
	"github.com/jonhadfield/gosn-v2/items"
	"github.com/stretchr/testify/require"
)

// writeBackupArchive builds an archive the way BackupConfig.Run does, without
// needing a session: manifest in the clear, payloads sealed when encrypting.
func writeBackupArchive(t *testing.T, path, password string, encrypt bool) {
	t.Helper()
	writeBackupArchiveWithItems(t, path, password, encrypt, []byte("[]"), []byte("[]"))
}

func writeBackupArchiveWithItems(t *testing.T, path, password string, encrypt bool, notesJSON, tagsJSON []byte) {
	t.Helper()

	b := &BackupConfig{OutputFile: path, Encrypt: encrypt, Password: password}

	// #nosec G304 -- path is a file inside the test's own t.TempDir().
	f, err := os.Create(path)
	require.NoError(t, err)
	defer f.Close()

	zw := zip.NewWriter(f)

	manifest := BackupManifest{
		ItemCounts: map[string]int{"notes": 0, "tags": 0},
		Encrypted:  encrypt,
		Version:    backupFormatVersion,
	}

	if encrypt {
		salt := make([]byte, backupSaltSize)
		for i := range salt {
			salt[i] = byte(i)
		}

		manifest.KDF = backupKDFName
		manifest.Salt = base64.StdEncoding.EncodeToString(salt)
		manifest.Iterations = backupKDFIterations

		aead, err := newGCM(deriveBackupKey(password, salt, backupKDFIterations))
		require.NoError(t, err)

		require.NoError(t, b.writeZipFile(zw, "notes.json", notesJSON, aead))
		require.NoError(t, b.writeZipFile(zw, "tags.json", tagsJSON, aead))
	} else {
		require.NoError(t, b.writeZipFile(zw, "notes.json", notesJSON, nil))
		require.NoError(t, b.writeZipFile(zw, "tags.json", tagsJSON, nil))
	}

	manifestData, err := json.MarshalIndent(manifest, "", "  ")
	require.NoError(t, err)
	require.NoError(t, b.writeZipFile(zw, "manifest.json", manifestData, nil))
	require.NoError(t, zw.Close())
}

// An encrypted backup must be restorable. Version 1.0 sealed the manifest as
// well, so restore could never parse it and the archive was unrecoverable.
func TestRestoreEncryptedBackupRoundTrip(t *testing.T) {
	const password = "correct horse battery staple"

	path := filepath.Join(t.TempDir(), "backup.zip")
	writeBackupArchive(t, path, password, true)

	r := RestoreConfig{InputFile: path, Password: password, DryRun: true}

	result, err := r.Run()
	require.NoError(t, err)
	require.True(t, result.Manifest.Encrypted)
	require.Equal(t, backupFormatVersion, result.Manifest.Version)
	require.Equal(t, backupKDFIterations, result.Manifest.Iterations)
}

func TestRestoreUnencryptedBackupRoundTrip(t *testing.T) {
	path := filepath.Join(t.TempDir(), "backup.zip")
	writeBackupArchive(t, path, "", false)

	r := RestoreConfig{InputFile: path, DryRun: true}

	result, err := r.Run()
	require.NoError(t, err)
	require.False(t, result.Manifest.Encrypted)
}

func TestRestoreEncryptedBackupWrongPassword(t *testing.T) {
	path := filepath.Join(t.TempDir(), "backup.zip")
	writeBackupArchive(t, path, "right-password", true)

	r := RestoreConfig{InputFile: path, Password: "wrong-password", DryRun: true}

	_, err := r.Run()
	require.Error(t, err)
	require.Contains(t, err.Error(), "decrypt")
}

func TestRestoreEncryptedBackupRequiresPassword(t *testing.T) {
	path := filepath.Join(t.TempDir(), "backup.zip")
	writeBackupArchive(t, path, "hunter2", true)

	r := RestoreConfig{InputFile: path, DryRun: true}

	_, err := r.Run()
	require.ErrorContains(t, err, "password required")
}

// Each backup must use its own salt, so one precomputed table cannot attack
// every archive sn-cli has ever written.
func TestBackupSaltsAreUnique(t *testing.T) {
	const runs = 20

	seen := make(map[string]bool, runs)

	for range runs {
		salt, err := newBackupSalt()
		require.NoError(t, err)
		require.Len(t, salt, backupSaltSize)

		encoded := base64.StdEncoding.EncodeToString(salt)
		require.False(t, seen[encoded], "salt repeated across backups")

		seen[encoded] = true
	}

	require.Len(t, seen, runs)
}

// PBKDF2 work factor should meet current OWASP guidance.
func TestBackupKDFParameters(t *testing.T) {
	require.GreaterOrEqual(t, backupKDFIterations, 600000)
	require.GreaterOrEqual(t, backupSaltSize, 16)
}

// GetBackupInfo must work on an encrypted backup: the manifest is not secret.
func TestGetBackupInfoOnEncryptedBackup(t *testing.T) {
	path := filepath.Join(t.TempDir(), "backup.zip")
	writeBackupArchive(t, path, "hunter2", true)

	info, err := GetBackupInfo(path)
	require.NoError(t, err)
	require.True(t, info.Encrypted)
	require.Equal(t, backupKDFName, info.KDF)
}

func TestNoteFromBackupItemPreservesUUIDAndContent(t *testing.T) {
	created := time.Date(2024, 3, 1, 12, 0, 0, 0, time.UTC)
	updated := created.Add(2 * time.Hour)

	content := items.NewNoteContent()
	content.Title = "Restored note"
	content.Text = "body text"
	content.SetUpdateTime(updated)

	contentJSON, err := json.Marshal(content)
	require.NoError(t, err)

	bi := BackupItem{
		UUID:      "note-uuid-1",
		Type:      common.SNItemTypeNote,
		Content:   string(contentJSON),
		CreatedAt: created.Format(common.TimeLayout),
		UpdatedAt: updated.Format(common.TimeLayout),
	}

	note, err := noteFromBackupItem(bi)
	require.NoError(t, err)
	require.Equal(t, "note-uuid-1", note.UUID)
	require.Equal(t, common.SNItemTypeNote, note.ContentType)
	require.Equal(t, "Restored note", note.Content.Title)
	require.Equal(t, "body text", note.Content.Text)
	require.Equal(t, created.UnixMicro(), note.CreatedAtTimestamp)
	require.Equal(t, updated.UnixMicro(), note.UpdatedAtTimestamp)

	gotUpdated, err := note.Content.GetUpdateTime()
	require.NoError(t, err)
	require.Equal(t, updated.Format(common.TimeLayout), gotUpdated.Format(common.TimeLayout))
}

func TestTagFromBackupItemFillsMissingUpdateTime(t *testing.T) {
	created := time.Date(2025, 1, 15, 9, 30, 0, 0, time.UTC)

	// Minimal archived content without ClientUpdatedAt.
	contentJSON := `{"title":"work","references":[]}`

	bi := BackupItem{
		UUID:      "tag-uuid-1",
		Type:      common.SNItemTypeTag,
		Content:   contentJSON,
		CreatedAt: created.Format(common.TimeLayout),
		UpdatedAt: created.Format(common.TimeLayout),
	}

	tag, err := tagFromBackupItem(bi)
	require.NoError(t, err)
	require.Equal(t, "tag-uuid-1", tag.UUID)
	require.Equal(t, "work", tag.Content.Title)

	gotUpdated, err := tag.Content.GetUpdateTime()
	require.NoError(t, err)
	require.Equal(t, created.Format(common.TimeLayout), gotUpdated.Format(common.TimeLayout))
}

func TestRestoreRequiresSessionWhenNotDryRun(t *testing.T) {
	path := filepath.Join(t.TempDir(), "backup.zip")

	content := items.NewNoteContent()
	content.Title = "A note"
	content.Text = "text"
	content.SetUpdateTime(time.Now().UTC())

	contentJSON, err := json.Marshal(content)
	require.NoError(t, err)

	notes := []BackupItem{{
		UUID:      "n1",
		Type:      common.SNItemTypeNote,
		Content:   string(contentJSON),
		CreatedAt: time.Now().UTC().Format(common.TimeLayout),
		UpdatedAt: time.Now().UTC().Format(common.TimeLayout),
	}}
	notesJSON, err := json.Marshal(notes)
	require.NoError(t, err)

	writeBackupArchiveWithItems(t, path, "", false, notesJSON, []byte("[]"))

	r := RestoreConfig{InputFile: path, DryRun: false}

	_, err = r.Run()
	require.ErrorContains(t, err, "session required")
}

func TestRestoreDryRunCountsItemsWithoutWriting(t *testing.T) {
	path := filepath.Join(t.TempDir(), "backup.zip")

	noteContent := items.NewNoteContent()
	noteContent.Title = "Note"
	noteContent.Text = "text"
	noteContent.SetUpdateTime(time.Now().UTC())
	noteJSON, err := json.Marshal(noteContent)
	require.NoError(t, err)

	tagContent := items.NewTagContent()
	tagContent.Title = "Tag"
	tagJSON, err := json.Marshal(tagContent)
	require.NoError(t, err)

	now := time.Now().UTC().Format(common.TimeLayout)
	notesPayload, err := json.Marshal([]BackupItem{{
		UUID: "n1", Type: common.SNItemTypeNote, Content: string(noteJSON),
		CreatedAt: now, UpdatedAt: now,
	}})
	require.NoError(t, err)
	tagsPayload, err := json.Marshal([]BackupItem{{
		UUID: "t1", Type: common.SNItemTypeTag, Content: string(tagJSON),
		CreatedAt: now, UpdatedAt: now,
	}})
	require.NoError(t, err)

	writeBackupArchiveWithItems(t, path, "", false, notesPayload, tagsPayload)

	r := RestoreConfig{InputFile: path, DryRun: true}
	result, err := r.Run()
	require.NoError(t, err)
	require.Equal(t, 1, result.NotesCount)
	require.Equal(t, 1, result.TagsCount)
	require.Zero(t, result.NotesRestored)
	require.Zero(t, result.TagsRestored)
}
