package sncli

import (
	"archive/zip"
	"encoding/base64"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
)

// writeBackupArchive builds an archive the way BackupConfig.Run does, without
// needing a session: manifest in the clear, payloads sealed when encrypting.
func writeBackupArchive(t *testing.T, path, password string, encrypt bool) {
	t.Helper()

	b := &BackupConfig{OutputFile: path, Encrypt: encrypt, Password: password}

	f, err := os.Create(path)
	require.NoError(t, err)
	defer f.Close()

	zw := zip.NewWriter(f)

	manifest := BackupManifest{
		ItemCounts: map[string]int{"notes": 0, "tags": 0},
		Encrypted:  encrypt,
		Version:    backupFormatVersion,
	}

	var gcm = func() (g interface {
		Seal([]byte, []byte, []byte, []byte) []byte
	}) {
		return nil
	}()
	_ = gcm

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

		for _, name := range []string{"notes.json", "tags.json"} {
			require.NoError(t, b.writeZipFile(zw, name, []byte("[]"), aead))
		}
	} else {
		for _, name := range []string{"notes.json", "tags.json"} {
			require.NoError(t, b.writeZipFile(zw, name, []byte("[]"), nil))
		}
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
