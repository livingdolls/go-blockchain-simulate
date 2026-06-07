package entity

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSentinelErrors_NotNil(t *testing.T) {
	// Memastikan semua sentinel error terdefinisi (tidak nil) dan memiliki
	// pesan yang non-kosong. Penting untuk logging yang bermakna.
	sentinels := map[string]error{
		"ErrNoPendingTransactions":    ErrNoPendingTransactions,
		"ErrUserBalanceNotFound":      ErrUserBalanceNotFound,
		"ErrAddressNotFound":          ErrAddressNotFound,
		"ErrAmountMustBePositive":     ErrAmountMustBePositive,
		"ErrUserWalletNotFound":       ErrUserWalletNotFound,
		"ErrInsufficientWalletBalance": ErrInsufficientWalletBalance,
		"ErrUserNotFound":             ErrUserNotFound,
		"ErrUsernameAlreadyExists":    ErrUsernameAlreadyExists,
		"ErrAddressAlreadyRegistered": ErrAddressAlreadyRegistered,
		"ErrTransactionNotFound":      ErrTransactionNotFound,
		"ErrInvalidTransactionType":   ErrInvalidTransactionType,
		"ErrSignatureVerificationFailed": ErrSignatureVerificationFailed,
		"ErrBlockNotFound":            ErrBlockNotFound,
		"ErrInvalidBlockData":         ErrInvalidBlockData,
		"ErrUnauthorized":             ErrUnauthorized,
		"ErrInvalidToken":             ErrInvalidToken,
		"ErrTokenExpired":             ErrTokenExpired,
		"ErrDatabase":                 ErrDatabase,
		"ErrInternalServer":           ErrInternalServer,
		"ErrInvalidInput":             ErrInvalidInput,
		"ErrOperationFailed":          ErrOperationFailed,
		"ErrConflict":                 ErrConflict,
		"ErrDependenciesNotInitialized": ErrDependenciesNotInitialized,
		"ErrAdminNotFound":            ErrAdminNotFound,
		"ErrDuplicateAdminUsername":   ErrDuplicateAdminUsername,
		"ErrAdminContextMissing":      ErrAdminContextMissing,
		"ErrInvalidAdminContext":      ErrInvalidAdminContext,
	}

	for name, err := range sentinels {
		assert.NotNil(t, err, "sentinel %s harus terdefinisi", name)
		assert.NotEmpty(t, err.Error(), "sentinel %s harus punya pesan", name)
	}
}

func TestSentinelErrors_Unique(t *testing.T) {
	// Memastikan dua sentinel berbeda tidak menggunakan pointer/pesan yang sama.
	// Kalau tidak unik, errors.Is bisa match yang salah.
	sentinels := map[string]error{
		"ErrNoPendingTransactions":    ErrNoPendingTransactions,
		"ErrUserNotFound":             ErrUserNotFound,
		"ErrAdminNotFound":            ErrAdminNotFound,
		"ErrBlockNotFound":            ErrBlockNotFound,
		"ErrTransactionNotFound":      ErrTransactionNotFound,
		"ErrInvalidInput":             ErrInvalidInput,
		"ErrOperationFailed":          ErrOperationFailed,
		"ErrConflict":                 ErrConflict,
		"ErrUnauthorized":             ErrUnauthorized,
		"ErrInvalidToken":             ErrInvalidToken,
		"ErrAdminContextMissing":      ErrAdminContextMissing,
		"ErrInvalidAdminContext":      ErrInvalidAdminContext,
		"ErrDuplicateAdminUsername":   ErrDuplicateAdminUsername,
	}

	seen := make(map[string]string)
	for name, err := range sentinels {
		msg := err.Error()
		if other, ok := seen[msg]; ok {
			t.Errorf("sentinel %s dan %s punya pesan identik: %q", other, name, msg)
			continue
		}
		seen[msg] = name
	}
}

func TestErrorsIs_Wrapping(t *testing.T) {
	// Memastikan errors.Is bekerja setelah wrap dengan %w.
	wrapped := errors.Join(errors.New("context"), ErrUserNotFound)
	assert.True(t, errors.Is(wrapped, ErrUserNotFound))

	wrappedFmt := wrapWithFmt(ErrInvalidInput, "extra")
	assert.True(t, errors.Is(wrappedFmt, ErrInvalidInput))
}

func wrapWithFmt(err error, msg string) error {
	return &wrappedErr{msg: msg, err: err}
}

type wrappedErr struct {
	msg string
	err error
}

func (w *wrappedErr) Error() string { return w.msg + ": " + w.err.Error() }
func (w *wrappedErr) Unwrap() error { return w.err }
