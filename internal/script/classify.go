package script

// ClassifyOutput classifies an output scriptPubKey
func ClassifyOutput(scriptPubKey []byte) string {
	n := len(scriptPubKey)

	// P2PKH: OP_DUP OP_HASH160 OP_PUSHBYTES_20 <20 bytes> OP_EQUALVERIFY OP_CHECKSIG
	if n == 25 && scriptPubKey[0] == 0x76 && scriptPubKey[1] == 0xa9 &&
		scriptPubKey[2] == 0x14 && scriptPubKey[23] == 0x88 && scriptPubKey[24] == 0xac {
		return "p2pkh"
	}

	// P2SH: OP_HASH160 OP_PUSHBYTES_20 <20 bytes> OP_EQUAL
	if n == 23 && scriptPubKey[0] == 0xa9 && scriptPubKey[1] == 0x14 && scriptPubKey[22] == 0x87 {
		return "p2sh"
	}

	// P2WPKH: OP_0 OP_PUSHBYTES_20 <20 bytes>
	if n == 22 && scriptPubKey[0] == 0x00 && scriptPubKey[1] == 0x14 {
		return "p2wpkh"
	}

	// P2WSH: OP_0 OP_PUSHBYTES_32 <32 bytes>
	if n == 34 && scriptPubKey[0] == 0x00 && scriptPubKey[1] == 0x20 {
		return "p2wsh"
	}

	// P2TR: OP_1 OP_PUSHBYTES_32 <32 bytes>
	if n == 34 && scriptPubKey[0] == 0x51 && scriptPubKey[1] == 0x20 {
		return "p2tr"
	}

	// OP_RETURN: starts with 0x6a
	if n >= 1 && scriptPubKey[0] == 0x6a {
		return "op_return"
	}

	return "unknown"
}

// ClassifyInputFromPrevout classifies based on prevout script type
func ClassifyInputFromPrevout(prevoutScript []byte) string {
	return ClassifyOutput(prevoutScript)
}

// HasNonEmptyWitness checks if witness has any non-empty items
func HasNonEmptyWitness(witness [][]byte) bool {
	for _, item := range witness {
		if len(item) > 0 {
			return true
		}
	}
	return false
}
