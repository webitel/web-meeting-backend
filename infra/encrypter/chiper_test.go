package encrypter_test

import (
	"bytes"
	"testing"

	"github.com/webitel/web-meeting-backend/infra/encrypter" // Імпортуємо пакет, який тестуємо
)

func TestNewDataEncrypter(t *testing.T) {
	t.Run("Valid key of standard length", func(t *testing.T) {
		key := bytes.Repeat([]byte("a"), 32) // A 32-byte key
		de, err := encrypter.New(key)
		if err != nil {
			t.Fatalf("New() failed with a valid 32-byte key: %v", err)
		}
		if de == nil {
			t.Fatal("New() returned nil DataEncrypter")
		}
	})

	t.Run("Key with non-standard length should now be valid due to hashing", func(t *testing.T) {
		// Цей тест перевіряє, що хешування працює як очікувалося.
		// Він замінює старий тест "Invalid key size".
		shortKey := []byte("short")
		de, err := encrypter.New(shortKey)
		if err != nil {
			t.Fatalf("New() failed with a short key, but it should have been hashed to a valid length. Error: %v", err)
		}
		if de == nil {
			t.Fatal("New() returned nil DataEncrypter for a short key")
		}

		longKey := bytes.Repeat([]byte("long-key-that-is-definitely-not-32-bytes"), 3)
		de, err = encrypter.New(longKey)
		if err != nil {
			t.Fatalf("New() failed with a long key, but it should have been hashed to a valid length. Error: %v", err)
		}
		if de == nil {
			t.Fatal("New() returned nil DataEncrypter for a long key")
		}
	})
}

func TestDataEncrypter_EncryptDecrypt(t *testing.T) {
	key := []byte("my-secret-key-of-any-length")
	de, err := encrypter.New(key)
	if err != nil {
		t.Fatalf("Failed to create DataEncrypter: %v", err)
	}

	// Стандартні розміри для AES-GCM
	const gcmNonceSize = 12
	const gcmOverhead = 16 // GCM tag size

	testCases := []struct {
		name      string
		plaintext []byte
	}{
		{"Empty plaintext", []byte("")},
		{"Short plaintext", []byte("hello world")},
		{"Long plaintext", bytes.Repeat([]byte("This is a longer test string for encryption and decryption."), 10)},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ciphertext, err := de.Encrypt(tc.plaintext)
			if err != nil {
				t.Fatalf("Encrypt() failed: %v", err)
			}

			// Перевіряємо, що зашифрований текст не порожній
			if len(ciphertext) == 0 && len(tc.plaintext) > 0 {
				t.Error("Encrypt() returned empty ciphertext for non-empty plaintext")
			}

			// Перевіряємо мінімальну довжину: nonce + GCM tag
			expectedMinLen := gcmNonceSize + len(tc.plaintext) + gcmOverhead
			if len(ciphertext) != expectedMinLen {
				t.Errorf("Encrypt() returned ciphertext with unexpected length: got %d, want %d", len(ciphertext), expectedMinLen)
			}

			// Дешифруємо та перевіряємо
			decryptedText, err := de.Decrypt(ciphertext)
			if err != nil {
				t.Fatalf("Decrypt() failed: %v", err)
			}

			if !bytes.Equal(decryptedText, tc.plaintext) {
				t.Errorf("Decrypted text mismatch:\nGot:  %q\nWant: %q", decryptedText, tc.plaintext)
			}
		})
	}
}

func TestDataEncrypter_DecryptInvalidInput(t *testing.T) {
	key := []byte("another-secret")
	de, err := encrypter.New(key)
	if err != nil {
		t.Fatalf("Failed to create DataEncrypter: %v", err)
	}

	const gcmNonceSize = 12 // Стандартний розмір nonce для GCM

	t.Run("Ciphertext too short (less than nonce size)", func(t *testing.T) {
		shortCiphertext := []byte("too_short") // 9 байт
		_, err := de.Decrypt(shortCiphertext)
		if err == nil {
			t.Fatal("Decrypt() succeeded with too short ciphertext, expected an error")
		}
		expectedErr := "ciphertext is too short: expected at least 12 bytes for nonce, got 9"
		if err.Error() != expectedErr {
			t.Errorf("Decrypt() returned unexpected error message:\nGot:  %q\nWant: %q", err.Error(), expectedErr)
		}
	})

	t.Run("Tampered ciphertext", func(t *testing.T) {
		plaintext := []byte("original message")
		ciphertext, err := de.Encrypt(plaintext)
		if err != nil {
			t.Fatalf("Encrypt() failed for tampered test: %v", err)
		}

		// Змінюємо байт у частині зашифрованого тексту (після nonce)
		tamperedCiphertext := make([]byte, len(ciphertext))
		copy(tamperedCiphertext, ciphertext)
		tamperedCiphertext[gcmNonceSize] ^= 0x01 // Перевертаємо біт

		_, err = de.Decrypt(tamperedCiphertext)
		if err == nil {
			t.Fatal("Decrypt() succeeded with tampered ciphertext, expected an error")
		}
		// Повідомлення про помилку від gcm.Open зазвичай "cipher: message authentication failed"
		expectedErr := "decryption failed: cipher: message authentication failed"
		if err.Error() != expectedErr {
			t.Errorf("Decrypt() returned unexpected error message for tampered data:\nGot:  %q\nWant: %q", err.Error(), expectedErr)
		}
	})
}
