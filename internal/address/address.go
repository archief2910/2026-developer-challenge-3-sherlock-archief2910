package address

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"math/big"
	"strings"
)

// Base58 alphabet
const base58Alphabet = "123456789ABCDEFGHJKLMNPQRSTUVWXYZabcdefghijkmnopqrstuvwxyz"

// Base58Check encodes data with version byte
func Base58Check(version byte, payload []byte) string {
	data := append([]byte{version}, payload...)
	checksum := doubleSHA256Checksum(data)
	full := append(data, checksum[:4]...)
	return base58Encode(full)
}

func doubleSHA256Checksum(data []byte) [32]byte {
	first := sha256.Sum256(data)
	return sha256.Sum256(first[:])
}

func base58Encode(data []byte) string {
	leadingZeros := 0
	for _, b := range data {
		if b == 0 {
			leadingZeros++
		} else {
			break
		}
	}

	num := new(big.Int).SetBytes(data)
	zero := big.NewInt(0)
	base := big.NewInt(58)
	mod := new(big.Int)

	var result []byte
	for num.Cmp(zero) > 0 {
		num.DivMod(num, base, mod)
		result = append(result, base58Alphabet[mod.Int64()])
	}

	for i := 0; i < leadingZeros; i++ {
		result = append(result, '1')
	}

	for i, j := 0, len(result)-1; i < j; i, j = i+1, j-1 {
		result[i], result[j] = result[j], result[i]
	}

	return string(result)
}

const bech32Charset = "qpzry9x8gf2tvdw0s3jn54khce6mua7l"

func bech32Polymod(values []int) int {
	generator := []int{0x3b6a57b2, 0x26508e6d, 0x1ea119fa, 0x3d4233dd, 0x2a1462b3}
	chk := 1
	for _, v := range values {
		top := chk >> 25
		chk = (chk&0x1ffffff)<<5 ^ v
		for i := 0; i < 5; i++ {
			if (top>>uint(i))&1 == 1 {
				chk ^= generator[i]
			}
		}
	}
	return chk
}

func bech32HRPExpand(hrp string) []int {
	result := make([]int, len(hrp)*2+1)
	for i, c := range hrp {
		result[i] = int(c) >> 5
	}
	result[len(hrp)] = 0
	for i, c := range hrp {
		result[len(hrp)+1+i] = int(c) & 31
	}
	return result
}

func bech32CreateChecksum(hrp string, data []int, spec int) []int {
	values := append(bech32HRPExpand(hrp), data...)
	values = append(values, 0, 0, 0, 0, 0, 0)
	polymod := bech32Polymod(values) ^ spec
	result := make([]int, 6)
	for i := 0; i < 6; i++ {
		result[i] = (polymod >> uint(5*(5-i))) & 31
	}
	return result
}

func bech32Encode(hrp string, data []int, spec int) string {
	checksum := bech32CreateChecksum(hrp, data, spec)
	combined := append(data, checksum...)
	var sb strings.Builder
	sb.WriteString(hrp)
	sb.WriteByte('1')
	for _, d := range combined {
		sb.WriteByte(bech32Charset[d])
	}
	return sb.String()
}

func convertBits(data []byte, fromBits, toBits int, pad bool) ([]int, error) {
	acc := 0
	bits := 0
	maxv := (1 << uint(toBits)) - 1
	var result []int

	for _, value := range data {
		acc = (acc << uint(fromBits)) | int(value)
		bits += fromBits
		for bits >= toBits {
			bits -= toBits
			result = append(result, (acc>>uint(bits))&maxv)
		}
	}

	if pad {
		if bits > 0 {
			result = append(result, (acc<<uint(toBits-bits))&maxv)
		}
	} else if bits >= fromBits {
		return nil, fmt.Errorf("excess bits")
	} else if (acc<<uint(toBits-bits))&maxv != 0 {
		return nil, fmt.Errorf("non-zero padding")
	}

	return result, nil
}

// SegwitAddress creates a bech32/bech32m segwit address
func SegwitAddress(hrp string, witnessVersion int, witnessProgram []byte) (string, error) {
	data5bit, err := convertBits(witnessProgram, 8, 5, true)
	if err != nil {
		return "", err
	}

	enc := append([]int{witnessVersion}, data5bit...)

	spec := 1 // bech32
	if witnessVersion > 0 {
		spec = 0x2bc830a3 // bech32m
	}

	return bech32Encode(hrp, enc, spec), nil
}

// DeriveAddress derives a Bitcoin mainnet address from a scriptPubKey hex string
func DeriveAddress(scriptPubKeyHex string) *string {
	script, err := hex.DecodeString(scriptPubKeyHex)
	if err != nil {
		return nil
	}
	return DeriveAddressFromBytes(script)
}

// DeriveAddressFromBytes derives a Bitcoin mainnet address from scriptPubKey bytes
func DeriveAddressFromBytes(script []byte) *string {
	n := len(script)

	// P2PKH
	if n == 25 && script[0] == 0x76 && script[1] == 0xa9 && script[2] == 0x14 &&
		script[23] == 0x88 && script[24] == 0xac {
		addr := Base58Check(0x00, script[3:23])
		return &addr
	}

	// P2SH
	if n == 23 && script[0] == 0xa9 && script[1] == 0x14 && script[22] == 0x87 {
		addr := Base58Check(0x05, script[2:22])
		return &addr
	}

	// P2WPKH
	if n == 22 && script[0] == 0x00 && script[1] == 0x14 {
		addr, err := SegwitAddress("bc", 0, script[2:22])
		if err != nil {
			return nil
		}
		return &addr
	}

	// P2WSH
	if n == 34 && script[0] == 0x00 && script[1] == 0x20 {
		addr, err := SegwitAddress("bc", 0, script[2:34])
		if err != nil {
			return nil
		}
		return &addr
	}

	// P2TR
	if n == 34 && script[0] == 0x51 && script[1] == 0x20 {
		addr, err := SegwitAddress("bc", 1, script[2:34])
		if err != nil {
			return nil
		}
		return &addr
	}

	return nil
}
