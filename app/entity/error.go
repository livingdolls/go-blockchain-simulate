package entity

import "errors"

var ErrNoPendingTransactions = errors.New("no pending transactions")
var ErrUserBalanceNotFound = errors.New("user balance not found")
var ErrAddressNotFound = errors.New("address not found")
var ErrAmountMustBePositive = errors.New("amount must be greater than zero")

// WALLET ERRORS
var ErrUserWalletNotFound = errors.New("user wallet not found")
var ErrInsufficientWalletBalance = errors.New("insufficient wallet balance")

// USER ERRORS
var ErrUserNotFound = errors.New("user not found")
var ErrUsernameAlreadyExists = errors.New("username already exists")
var ErrAddressAlreadyRegistered = errors.New("address already registered")

// TRANSACTION ERRORS
var ErrTransactionNotFound = errors.New("transaction not found")
var ErrInvalidTransactionType = errors.New("invalid transaction type")
var ErrSignatureVerificationFailed = errors.New("signature verification failed")

// BLOCK ERRORS
var ErrBlockNotFound = errors.New("block not found")
var ErrInvalidBlockData = errors.New("invalid block data")

// AUTHENTICATION ERRORS
var ErrUnauthorized = errors.New("unauthorized access")
var ErrInvalidToken = errors.New("invalid token")
var ErrTokenExpired = errors.New("token has expired")

// GENERAL ERRORS
var ErrDatabase = errors.New("database error")
var ErrInternalServer = errors.New("internal server error")
var ErrInvalidInput = errors.New("invalid input")
var ErrOperationFailed = errors.New("operation failed")
var ErrConflict = errors.New("resource conflict")
