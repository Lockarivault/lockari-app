package encryption

import (
	"context"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"fmt"
	"io"
)

// Encryptor exposes the public API for DEK lifecycle management.
type Encryptor interface {
	GenerateDEK(ctx context.Context) ([]byte, error)
	EncryptDEKWithKEK(ctx context.Context, dek []byte) (*Envelope, error)
	DecryptDEK(ctx context.Context, envelope *Envelope) ([]byte, error)
}

// Option allows configuring the encryptor behaviour.
type Option func(*encryptor)

// WithRandReader overrides the randomness source used to generate DEKs and nonces.
// Intended for tests; production uses crypto/rand.Reader.
func WithRandReader(r io.Reader) Option {
	return func(e *encryptor) {
		if r != nil {
			e.rand = r
		}
	}
}

// WithDEKLength sets the desired DEK length. Valid values are 16, 24 and 32 bytes.
func WithDEKLength(length int) Option {
	return func(e *encryptor) {
		if length > 0 {
			e.dekLength = length
		}
	}
}

type encryptor struct {
	kek       []byte
	keyID     string
	rand      io.Reader
	dekLength int
}

// NewEncryptor constructs an Encryptor from the provided KEK material and options.
func NewEncryptor(kek []byte, keyID string, opts ...Option) (Encryptor, error) {
	if len(kek) != 32 {
		return nil, fmt.Errorf("%w: got %d", ErrInvalidKEKLength, len(kek))
	}
	e := &encryptor{
		kek:       append([]byte(nil), kek...), // defensive copy
		keyID:     keyID,
		rand:      rand.Reader,
		dekLength: 32,
	}
	for _, opt := range opts {
		opt(e)
	}
	if e.dekLength != 16 && e.dekLength != 24 && e.dekLength != 32 {
		return nil, fmt.Errorf("%w: %d", ErrUnsupportedDEKLength, e.dekLength)
	}
	return e, nil
}

// Envelope stores the metadata required to reconstruct a DEK.
type Envelope struct {
	Ciphertext string `json:"ciphertext" yaml:"ciphertext"`
	Nonce      string `json:"nonce" yaml:"nonce"`
	KeyID      string `json:"key_id" yaml:"key_id"`
}

func (e *encryptor) GenerateDEK(_ context.Context) ([]byte, error) {
	if e == nil {
		return nil, ErrNilEncryptor
	}
	buf := make([]byte, e.dekLength)
	if _, err := io.ReadFull(e.rand, buf); err != nil {
		return nil, fmt.Errorf("%w: %v", ErrEntropyFailure, err)
	}
	return buf, nil
}

func (e *encryptor) EncryptDEKWithKEK(_ context.Context, dek []byte) (*Envelope, error) {
	if e == nil {
		return nil, ErrNilEncryptor
	}
	if len(dek) == 0 {
		return nil, ErrEmptyDEK
	}
	block, err := aes.NewCipher(e.kek)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrCipherInit, err)
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrCipherInit, err)
	}
	nonce := make([]byte, gcm.NonceSize())
	if _, err := io.ReadFull(e.rand, nonce); err != nil {
		return nil, fmt.Errorf("%w: %v", ErrEntropyFailure, err)
	}
	ciphertext := gcm.Seal(nil, nonce, dek, nil)
	return &Envelope{
		Ciphertext: EncodeBase64(ciphertext),
		Nonce:      EncodeBase64(nonce),
		KeyID:      e.keyID,
	}, nil
}

func (e *encryptor) DecryptDEK(_ context.Context, envelope *Envelope) ([]byte, error) {
	if e == nil {
		return nil, ErrNilEncryptor
	}
	if envelope == nil {
		return nil, ErrNilEnvelope
	}
	if envelope.KeyID != "" && envelope.KeyID != e.keyID {
		return nil, fmt.Errorf("%w: got %s expected %s", ErrMismatchedKeyID, envelope.KeyID, e.keyID)
	}
	block, err := aes.NewCipher(e.kek)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrCipherInit, err)
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrCipherInit, err)
	}
	nonce, err := DecodeBase64(envelope.Nonce)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrInvalidEncoding, err)
	}
	ciphertext, err := DecodeBase64(envelope.Ciphertext)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrInvalidEncoding, err)
	}
	if len(nonce) != gcm.NonceSize() {
		return nil, fmt.Errorf("%w: nonce size %d", ErrInvalidNonce, len(nonce))
	}
	dek, err := gcm.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrAuthenticationFailed, err)
	}
	return dek, nil
}

// GenerateRandomKey generates a cryptographically secure random key of the specified length.
func GenerateRandomKey(length int) ([]byte, error) {
	key := make([]byte, length)
	if _, err := io.ReadFull(rand.Reader, key); err != nil {
		return nil, fmt.Errorf("%w: %v", ErrEntropyFailure, err)
	}
	return key, nil
}
