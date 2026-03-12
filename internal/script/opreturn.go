package script

import (
	"encoding/hex"
	"strings"
	"unicode/utf8"
)

// OpReturnResult holds decoded OP_RETURN data
type OpReturnResult struct {
	DataHex  string
	DataUTF8 *string
	Protocol string
}

// DecodeOpReturn extracts OP_RETURN payload data
func DecodeOpReturn(scriptPubKey []byte) OpReturnResult {
	if len(scriptPubKey) < 1 || scriptPubKey[0] != 0x6a {
		return OpReturnResult{DataHex: "", Protocol: "unknown"}
	}

	var allData []byte
	offset := 1

	for offset < len(scriptPubKey) {
		op := scriptPubKey[offset]
		offset++

		var pushData []byte

		if op >= 0x01 && op <= 0x4b {
			size := int(op)
			end := offset + size
			if end > len(scriptPubKey) {
				break
			}
			pushData = scriptPubKey[offset:end]
			offset = end
		} else if op == 0x4c { // OP_PUSHDATA1
			if offset >= len(scriptPubKey) {
				break
			}
			size := int(scriptPubKey[offset])
			offset++
			end := offset + size
			if end > len(scriptPubKey) {
				break
			}
			pushData = scriptPubKey[offset:end]
			offset = end
		} else if op == 0x4d { // OP_PUSHDATA2
			if offset+2 > len(scriptPubKey) {
				break
			}
			size := int(scriptPubKey[offset]) | (int(scriptPubKey[offset+1]) << 8)
			offset += 2
			end := offset + size
			if end > len(scriptPubKey) {
				break
			}
			pushData = scriptPubKey[offset:end]
			offset = end
		} else if op == 0x4e { // OP_PUSHDATA4
			if offset+4 > len(scriptPubKey) {
				break
			}
			size := int(scriptPubKey[offset]) | (int(scriptPubKey[offset+1]) << 8) | (int(scriptPubKey[offset+2]) << 16) | (int(scriptPubKey[offset+3]) << 24)
			offset += 4
			end := offset + size
			if end > len(scriptPubKey) {
				break
			}
			pushData = scriptPubKey[offset:end]
			offset = end
		} else {
			continue
		}

		allData = append(allData, pushData...)
	}

	dataHex := hex.EncodeToString(allData)

	var dataUTF8 *string
	if utf8.Valid(allData) {
		s := string(allData)
		dataUTF8 = &s
	}

	// Protocol classification by payload prefix.
	// Each prefix is the hex encoding of the protocol's magic bytes.
	// Reference: arXiv 2411.10325v1 — practical colored coin detection.
	protocol := "unknown"
	if strings.HasPrefix(dataHex, "6f6d6e69") {
		// "omni" in ASCII — Omni Layer (formerly Mastercoin)
		protocol = "omni"
	} else if strings.HasPrefix(dataHex, "0109f91102") {
		// OpenTimestamps calendar commitment marker
		protocol = "opentimestamps"
	} else if strings.HasPrefix(dataHex, "434e545250525459") {
		// "CNTRPRTY" in ASCII — Counterparty protocol magic bytes
		// Reference: https://counterparty.io — ARC4-encrypted data starts with CNTRPRTY after decryption
		protocol = "counterparty"
	} else if strings.HasPrefix(dataHex, "56424b") {
		// "VBK" in ASCII — Veriblock proof-of-proof
		protocol = "veriblock"
	} else if strings.HasPrefix(dataHex, "4f41") {
		// "OA" in ASCII — Open Assets / EPOBC colored coins
		protocol = "openassets"
	}

	return OpReturnResult{
		DataHex:  dataHex,
		DataUTF8: dataUTF8,
		Protocol: protocol,
	}
}
