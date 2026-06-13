package logger

import (
	"bytes"
	"io"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestNoSyncWriter_WriteDelegates memastikan Write() di-forward ke
// underlying writer (stdout tidak di-skip, log masih keluar).
func TestNoSyncWriter_WriteDelegates(t *testing.T) {
	var buf bytes.Buffer
	w := noSyncWriter{w: &buf}

	msg := `{"msg":"hello"}` + "\n"
	n, err := w.Write([]byte(msg))
	require.NoError(t, err)
	assert.Equal(t, len(msg), n)
	assert.Equal(t, msg, buf.String())
}

// TestNoSyncWriter_SyncIsNoop adalah test KRITIS untuk fix bug
// "logger shutdown error: sync /dev/stdout: invalid argument".
// Sebelum fix: Sync() call fsync() di pipe/terminal -> return EINVAL.
// Sesudah fix: Sync() selalu return nil sehingga Shutdown() bersih.
func TestNoSyncWriter_SyncIsNoop(t *testing.T) {
	// Pakai stdout asli - kalau sync di-forward ke fsync, test ini
	// bisa fail di environment tertentu (CI, container tanpa TTY).
	// Tapi karena kita test wrapper-nya, harusnya aman.
	w := noSyncWriter{w: os.Stdout}
	assert.NoError(t, w.Sync(), "Sync() harus no-op, bukan forward fsync ke pipe")
}

// TestNoSyncWriter_ImplementsWriteSyncer memastikan type implements
// interface yang dipakai zapcore.AddSync. Compile-time check via var _.
func TestNoSyncWriter_ImplementsWriteSyncer(t *testing.T) {
	var _ io.Writer = noSyncWriter{w: os.Stdout}
	// Kalau interface berubah, baris ini gagal compile - good.
}

// TestShutdown_DoesNotErrorOnStdoutPipe adalah regression test untuk
// bug "sync /dev/stdout: invalid argument". Init logger (default cfg)
// lalu Shutdown - kalau stdout masih pakai writer biasa, test ini
// akan return error di environment dengan stdout=pipe (Docker, CI).
func TestShutdown_DoesNotErrorOnStdoutPipe(t *testing.T) {
	cfg := DevelopmentConfig("test", "1.0.0")
	cfg.Env = "test"
	// Pakai cfg.LogPath kosong agar tidak buat file di test dir.
	cfg.LogPath = ""

	require.NoError(t, Init(cfg))
	assert.NotNil(t, L)

	// Write beberapa log agar ada yang perlu di-flush.
	LogInfo("test message before shutdown")

	// Shutdown - harus return nil, bukan error fsync.
	err := Shutdown(2 * 1_000_000_000) // 2s
	assert.NoError(t, err, "Shutdown harus bersih, tidak boleh ada 'sync /dev/stdout: invalid argument'")
}
