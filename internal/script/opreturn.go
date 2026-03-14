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
	// Only verified, factual protocol prefixes are included.
	// Reference: arXiv 2411.10325v1 — practical colored coin detection.
	//
	// Prefixes intentionally NOT included:
	//   - Counterparty ("CNTRPRTY" / 434e545250525459): Data is ARC4-encrypted
	//     using the first input's TXID as key. The "CNTRPRTY" magic bytes only
	//     appear AFTER decryption, not in raw OP_RETURN data. Matching raw bytes
	//     against this prefix would never work. (Source: counterparty.io docs)
	//   - OpenTimestamps: OTS embeds a raw 32-byte hash digest directly in
	//     OP_RETURN with no protocol prefix. There is no fixed magic byte
	//     sequence to match. (Source: petertodd.org/opentimestamps)
	//   - Veriblock: VBK PoP transactions use 80-byte OP_RETURN data identified
	//     by internal structure (version, height, timestamp), not a fixed ASCII
	//     prefix. (Source: blockchainresearchlab.org analysis of VBK)
	protocol := "unknown"
	if strings.HasPrefix(dataHex, "6f6d6e69") {
		// "omni" in ASCII — Omni Layer (formerly Mastercoin)
		// Well-documented, unencrypted prefix in OP_RETURN data.
		protocol = "omni"
	} else if strings.HasPrefix(dataHex, "4f41") {
		// "OA" in ASCII (0x4f 0x41) — Open Assets Protocol tag
		// Followed by version bytes (0x01 0x00) and asset quantity list.
		// Reference: Open Assets Protocol specification (bitcoinwiki.org)
		protocol = "openassets"
	}

	return OpReturnResult{
		DataHex:  dataHex,
		DataUTF8: dataUTF8,
		Protocol: protocol,
	}
}
