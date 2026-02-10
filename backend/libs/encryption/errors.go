package encryption

import "errors"

var (
	// ErrInvalidKEKLength is returned when the KEK length is not 32 bytes (AES-256).
	ErrInvalidKEKLength = errors.New("invalid KEK length: must be 32 bytes for AES-256")

	// ErrUnsupportedDEKLength is returned for lengths other than 16, 24, or 32 bytes.
	ErrUnsupportedDEKLength = errors.New("unsupported DEK length: must be 16, 24, or 32 bytes")

	// ErrNilEncryptor is returned when calling methods on a nil encryptor instance.
	ErrNilEncryptor = errors.New("encryptor instance is nil")

	// ErrEntropyFailure is returned when the random number generator fails.
	ErrEntropyFailure = errors.New("failed to generate random bytes")

	// ErrEmptyDEK is returned when trying to encrypt an empty DEK.
	ErrEmptyDEK = errors.New("DEK cannot be empty")

	// ErrCipherInit is returned when the AES block or GCM instance fails to initialize.
	ErrCipherInit = errors.New("failed to initialize cipher")

	// ErrNilEnvelope is returned when a nil envelope is provided for decryption.
	ErrNilEnvelope = errors.New("envelope is nil")

	// ErrMismatchedKeyID is returned when the envelope's KeyID doesn't match the encryptor's KeyID.
	ErrMismatchedKeyID = errors.New("mismatched key ID")

	// ErrInvalidEncoding is returned when base64/hex decoding fails.
	ErrInvalidEncoding = errors.New("invalid string encoding")

	// ErrInvalidNonce is returned when the nonce size is incorrect.
	ErrInvalidNonce = errors.New("invalid nonce size")

	// ErrAuthenticationFailed is returned when GCM decryption fails (e.g., tampered data or wrong KEK).
	ErrAuthenticationFailed = errors.New("authentication failed during decryption")
)
