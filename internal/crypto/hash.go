package crypto

import (
	"crypto/sha256"
	"encoding/hex"

	"golang.org/x/crypto/ripemd160"
)

// DoubleSHA256 computes SHA256(SHA256(data))
func DoubleSHA256(data []byte) [32]byte {
	first := sha256.Sum256(data)
	return sha256.Sum256(first[:])
}

// Hash160 computes RIPEMD160(SHA256(data))
func Hash160(data []byte) []byte {
	sha := sha256.Sum256(data)
	r := ripemd160.New()
	r.Write(sha[:])
	return r.Sum(nil)
}

// TxID computes the transaction ID (double SHA-256 of legacy serialization, reversed)
func TxID(legacySerialization []byte) string {
	hash := DoubleSHA256(legacySerialization)
	reversed := ReverseBytes(hash[:])
	return hex.EncodeToString(reversed)
}

// WTxID computes the witness transaction ID
func WTxID(fullSerialization []byte) string {
	hash := DoubleSHA256(fullSerialization)
	reversed := ReverseBytes(hash[:])
	return hex.EncodeToString(reversed)
}

// ReverseBytes returns a reversed copy of the byte slice
func ReverseBytes(data []byte) []byte {
	result := make([]byte, len(data))
	for i, b := range data {
		result[len(data)-1-i] = b
	}
	return result
}
