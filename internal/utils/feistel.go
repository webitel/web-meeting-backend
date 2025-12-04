package utils

import (
	"errors"
	"fmt"
	"math"
	"strings"
)

const Base62Charset = "0123456789abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"
const Base = 62

const FeistelKey1 uint32 = 0x5a4f7831
const FeistelKey2 uint32 = 0xd8e9c4b7
const FeistelKey3 uint32 = 0x9a8b7c6d

func F(r uint32, key uint32) uint32 {
	return ((r << 13) | (r >> 19)) + (r ^ key)
}

func ObfuscateID(id int64) int64 {

	idU64 := uint64(id)
	L := uint32(idU64 >> 32)
	R := uint32(idU64)

	// 1: Swap L, R
	L, R = R, L^F(R, FeistelKey1)

	// 2: Swap L, R
	L, R = R, L^F(R, FeistelKey2)

	// 3: Swap L, R
	L, R = R, L^F(R, FeistelKey3)

	return int64((uint64(L) << 32) | uint64(R))
}

func DeobfuscateID(cipher int64) int64 {
	cipherU64 := uint64(cipher)
	L := uint32(cipherU64 >> 32)
	R := uint32(cipherU64)

	L, R = R^F(L, FeistelKey3), L
	L, R = R^F(L, FeistelKey2), L
	L, R = R^F(L, FeistelKey1), L

	return int64((uint64(L) << 32) | uint64(R))
}

func EncodeID(id uint64) string {
	if id == 0 {
		return string(Base62Charset[0])
	}

	var encoded string
	for id > 0 {
		remainder := id % Base
		encoded = string(Base62Charset[remainder]) + encoded
		id /= Base
	}

	return encoded
}

func DecodeCode(code string) (uint64, error) {
	var id uint64 = 0

	for _, char := range code {
		index := strings.IndexRune(Base62Charset, char)
		if index == -1 {
			return 0, errors.New("invalid character in Base62 code")
		}
		// Check for overflow before multiplication
		if id > (math.MaxUint64-uint64(index))/Base {
			return 0, fmt.Errorf("decoding will cause overflow")
		}
		id = id*Base + uint64(index)
	}

	return id, nil
}

func ShortCode(id int64) string {
	obfuscatedID := ObfuscateID(id)
	return EncodeID(uint64(obfuscatedID))
}

func DecodeShortCode(code string) (int64, error) {
	decodedObfuscatedID, err := DecodeCode(code)
	if err != nil {
		return 0, err
	}
	return DeobfuscateID(int64(decodedObfuscatedID)), nil
}
