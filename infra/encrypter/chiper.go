package encrypter

import (
	"crypto"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"fmt"
	"io"

	_ "golang.org/x/crypto/sha3"
)

type DataEncrypter struct {
	gcm cipher.AEAD
}

func New(key []byte) (*DataEncrypter, error) {
	// Використовуємо SHA3-256 для отримання 32-байтного ключа з будь-якого вхідного ключа.
	sh := crypto.SHA3_256.New()
	sh.Write(key)
	hash := sh.Sum(nil)

	block, err := aes.NewCipher(hash)
	if err != nil {
		// Ця помилка теоретично не повинна виникати, оскільки хеш завжди правильного розміру.
		return nil, err
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}

	return &DataEncrypter{
		gcm: gcm,
	}, nil
}

// Encrypt шифрує дані за допомогою AES-GCM.
func (de *DataEncrypter) Encrypt(plaintext []byte) ([]byte, error) {
	nonce := make([]byte, de.gcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return nil, fmt.Errorf("failed to generate nonce: %w", err)
	}

	// Додаємо nonce на початок зашифрованого тексту.
	return de.gcm.Seal(nonce, nonce, plaintext, nil), nil
}

// Decrypt дешифрує дані, зашифровані за допомогою AES-GCM.
func (de *DataEncrypter) Decrypt(text []byte) ([]byte, error) {
	nonceSize := de.gcm.NonceSize()
	if len(text) < nonceSize {
		return nil, fmt.Errorf("ciphertext is too short: expected at least %d bytes for nonce, got %d", nonceSize, len(text))
	}

	nonce, ciphertext := text[:nonceSize], text[nonceSize:]
	plaintext, err := de.gcm.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return nil, fmt.Errorf("decryption failed: %w", err)
	}

	return plaintext, nil
}
