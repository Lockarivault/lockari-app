package encryption

import (
	"bytes"
	"context"
	"crypto/rand"
	"errors"
	"io"
	"testing"
)

func TestEncryptionLifecycle(t *testing.T) {
	ctx := context.Background()

	// Generate a dummy KEK (32 bytes for AES-256)
	kek := make([]byte, 32)
	if _, err := io.ReadFull(rand.Reader, kek); err != nil {
		t.Fatal(err)
	}

	keyID := "test-kek-id"

	encryptor, err := NewEncryptor(kek, keyID)
	if err != nil {
		t.Fatalf("failed to create encryptor: %v", err)
	}

	// 1. Generate DEK
	dek, err := encryptor.GenerateDEK(ctx)
	if err != nil {
		t.Fatalf("failed to generate DEK: %v", err)
	}
	if len(dek) != 32 {
		t.Errorf("expected DEK length 32, got %d", len(dek))
	}

	// 2. Encrypt DEK (Envelope)
	envelope, err := encryptor.EncryptDEKWithKEK(ctx, dek)
	if err != nil {
		t.Fatalf("failed to encrypt DEK: %v", err)
	}
	if envelope.KeyID != keyID {
		t.Errorf("expected KeyID %s in envelope, got %s", keyID, envelope.KeyID)
	}
	if envelope.Ciphertext == "" {
		t.Error("envelope ciphertext is empty")
	}

	// 3. Decrypt DEK
	decryptedDEK, err := encryptor.DecryptDEK(ctx, envelope)
	if err != nil {
		t.Fatalf("failed to decrypt DEK: %v", err)
	}

	// 4. Verify
	if !bytes.Equal(dek, decryptedDEK) {
		t.Error("decrypted DEK mismatch original DEK")
	}
}

func TestNewEncryptor_Validation(t *testing.T) {
	t.Run("Invalid KEK Length", func(t *testing.T) {
		_, err := NewEncryptor([]byte("too-short"), "id")
		if err == nil {
			t.Error("expected error for invalid KEK length")
		}
	})

	t.Run("Invalid DEK Length Option", func(t *testing.T) {
		kek := make([]byte, 32)
		_, err := NewEncryptor(kek, "id", WithDEKLength(10))
		if err == nil {
			t.Error("expected error for unsupported DEK length")
		}
	})
}

func TestDecryptDEK_Errors(t *testing.T) {
	ctx := context.Background()
	kek := make([]byte, 32)
	keyID := "key-1"
	encryptor, _ := NewEncryptor(kek, keyID)

	t.Run("Mismatched KeyID", func(t *testing.T) {
		envelope := &Envelope{KeyID: "wrong-id"}
		_, err := encryptor.DecryptDEK(ctx, envelope)
		if !errors.Is(err, ErrMismatchedKeyID) {
			t.Errorf("expected mismatched key id error, got %v", err)
		}
	})

	t.Run("Nil Envelope", func(t *testing.T) {
		_, err := encryptor.DecryptDEK(ctx, nil)
		if !errors.Is(err, ErrNilEnvelope) {
			t.Errorf("expected nil envelope error, got %v", err)
		}
	})
}
