package utils

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestFeistel(t *testing.T) {
	t.Run("ObfuscateID and DeobfuscateID round trip", func(t *testing.T) {
		tests := []int64{
			1,
			12345,
			9876543210,
			111,
			0, // Edge case if applicable, though usually IDs are > 0
		}

		for _, originalID := range tests {
			obfuscated := ObfuscateID(originalID)
			deobfuscated := DeobfuscateID(obfuscated)
			assert.Equal(t, originalID, deobfuscated, "Deobfuscated ID should match original ID for input %d", originalID)
			assert.NotEqual(t, originalID, obfuscated, "Obfuscated ID should generally not be equal to original ID for input %d", originalID)
		}
	})
}

func TestBase62(t *testing.T) {
	t.Run("EncodeID and DecodeCode round trip", func(t *testing.T) {
		tests := []uint64{
			0,
			1,
			61,
			62,
			123456789,
			18446744073709551615, // Max Uint64
		}

		for _, originalID := range tests {
			encoded := EncodeID(originalID)
			decoded, err := DecodeCode(encoded)
			require.NoError(t, err)
			assert.Equal(t, originalID, decoded, "Decoded ID should match original ID for input %d", originalID)
		}
	})

	t.Run("DecodeCode invalid input", func(t *testing.T) {
		invalidCodes := []string{
			"Invalid#Char",
			" ",
			"-",
		}
		for _, code := range invalidCodes {
			_, err := DecodeCode(code)
			assert.Error(t, err, "Should detect invalid characters in code %s", code)
		}
	})
}

func TestShortCode(t *testing.T) {
	t.Run("ShortCode and DecodeShortCode round trip", func(t *testing.T) {
		tests := []int64{
			1,
			100,
			123456,
			99999999,
		}

		for _, originalID := range tests {
			code := ShortCode(originalID)
			assert.NotEmpty(t, code)

			decoded, err := DecodeShortCode(code)
			require.NoError(t, err)
			assert.Equal(t, originalID, decoded, "Decoded ShortCode should match original ID for input %d", originalID)
		}
	})
}
