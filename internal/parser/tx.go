package parser

import (
	"encoding/binary"
	"fmt"

	"sherlock/internal/types"
)

// ParseTransactionBytes parses a raw Bitcoin transaction from bytes.
// Returns the parsed transaction, legacy/full serializations, and the number of consumed bytes.
func ParseTransactionBytes(data []byte) (*types.RawTransaction, []byte, []byte, int, error) {
	r := NewReader(data)
	tx := &types.RawTransaction{}

	// Version (4 bytes)
	versionStart := r.Offset()
	version, err := r.ReadUint32LE()
	if err != nil {
		return nil, nil, nil, 0, fmt.Errorf("reading version: %w", err)
	}
	tx.Version = version

	// Check for SegWit marker
	markerPos := r.Offset()
	marker, err := r.ReadByte()
	if err != nil {
		return nil, nil, nil, 0, fmt.Errorf("reading marker: %w", err)
	}

	segwit := false
	if marker == 0x00 {
		flag, err := r.ReadByte()
		if err != nil {
			return nil, nil, nil, 0, fmt.Errorf("reading flag: %w", err)
		}
		if flag != 0x01 {
			return nil, nil, nil, 0, fmt.Errorf("invalid segwit flag: %d", flag)
		}
		segwit = true
	} else {
		// Not segwit, rewind
		r.offset = markerPos
	}
	tx.IsSegwit = segwit

	// Track start of inputs+outputs section
	ioStart := r.Offset()

	// Inputs
	inputCount, err := r.ReadVarInt()
	if err != nil {
		return nil, nil, nil, 0, fmt.Errorf("reading input count: %w", err)
	}

	tx.Inputs = make([]types.TxInput, inputCount)
	for i := uint64(0); i < inputCount; i++ {
		input, err := parseInput(r)
		if err != nil {
			return nil, nil, nil, 0, fmt.Errorf("parsing input %d: %w", i, err)
		}
		tx.Inputs[i] = input
	}

	// Outputs
	outputCount, err := r.ReadVarInt()
	if err != nil {
		return nil, nil, nil, 0, fmt.Errorf("reading output count: %w", err)
	}

	tx.Outputs = make([]types.TxOutput, outputCount)
	for i := uint64(0); i < outputCount; i++ {
		output, err := parseOutput(r)
		if err != nil {
			return nil, nil, nil, 0, fmt.Errorf("parsing output %d: %w", i, err)
		}
		tx.Outputs[i] = output
	}

	// Track end of inputs+outputs section
	ioEnd := r.Offset()

	// Witness (if segwit)
	if segwit {
		tx.Witnesses = make([][]types.WitnessItem, inputCount)
		for i := uint64(0); i < inputCount; i++ {
			witnessCount, err := r.ReadVarInt()
			if err != nil {
				return nil, nil, nil, 0, fmt.Errorf("reading witness count for input %d: %w", i, err)
			}
			items := make([]types.WitnessItem, witnessCount)
			for j := uint64(0); j < witnessCount; j++ {
				itemLen, err := r.ReadVarInt()
				if err != nil {
					return nil, nil, nil, 0, fmt.Errorf("reading witness item length: %w", err)
				}
				item, err := r.ReadBytesRef(int(itemLen))
				if err != nil {
					return nil, nil, nil, 0, fmt.Errorf("reading witness item: %w", err)
				}
				items[j] = item
			}
			tx.Witnesses[i] = items
		}
	} else {
		tx.Witnesses = make([][]types.WitnessItem, inputCount)
		for i := uint64(0); i < inputCount; i++ {
			tx.Witnesses[i] = []types.WitnessItem{}
		}
	}

	// Locktime (4 bytes)
	locktimeStart := r.Offset()
	locktime, err := r.ReadUint32LE()
	if err != nil {
		return nil, nil, nil, 0, fmt.Errorf("reading locktime: %w", err)
	}
	tx.Locktime = locktime

	consumed := r.Offset()

	// Build legacy serialization
	var legacySer []byte
	if segwit {
		versionBytes := data[versionStart : versionStart+4]
		ioBytes := data[ioStart:ioEnd]
		locktimeBytes := data[locktimeStart : locktimeStart+4]
		legacySer = make([]byte, 4+len(ioBytes)+4)
		copy(legacySer[0:4], versionBytes)
		copy(legacySer[4:4+len(ioBytes)], ioBytes)
		copy(legacySer[4+len(ioBytes):], locktimeBytes)
	} else {
		legacySer = make([]byte, consumed)
		copy(legacySer, data[:consumed])
	}

	fullSer := data[:consumed]

	// Compute byte counts
	tx.TotalBytes = consumed
	if segwit {
		tx.NonWitnessBytes = len(legacySer)
		tx.WitnessBytes = tx.TotalBytes - tx.NonWitnessBytes
	} else {
		tx.NonWitnessBytes = consumed
		tx.WitnessBytes = 0
	}

	return tx, legacySer, fullSer, consumed, nil
}

func parseInput(r *Reader) (types.TxInput, error) {
	var input types.TxInput

	txidBytes, err := r.ReadBytesRef(32)
	if err != nil {
		return input, err
	}
	copy(input.Txid[:], txidBytes)

	vout, err := r.ReadUint32LE()
	if err != nil {
		return input, err
	}
	input.Vout = vout

	scriptLen, err := r.ReadVarInt()
	if err != nil {
		return input, err
	}
	scriptSig, err := r.ReadBytesRef(int(scriptLen))
	if err != nil {
		return input, err
	}
	input.ScriptSig = make([]byte, len(scriptSig))
	copy(input.ScriptSig, scriptSig)

	seq, err := r.ReadUint32LE()
	if err != nil {
		return input, err
	}
	input.Sequence = seq

	return input, nil
}

func parseOutput(r *Reader) (types.TxOutput, error) {
	var output types.TxOutput

	val, err := r.ReadUint64LE()
	if err != nil {
		return output, err
	}
	output.Value = int64(val)

	scriptLen, err := r.ReadVarInt()
	if err != nil {
		return output, err
	}
	scriptRef, err := r.ReadBytesRef(int(scriptLen))
	if err != nil {
		return output, err
	}
	output.ScriptPubKey = make([]byte, len(scriptRef))
	copy(output.ScriptPubKey, scriptRef)

	return output, nil
}

func encodeVarInt(v uint64) []byte {
	switch {
	case v < 0xFD:
		return []byte{byte(v)}
	case v <= 0xFFFF:
		b := make([]byte, 3)
		b[0] = 0xFD
		binary.LittleEndian.PutUint16(b[1:], uint16(v))
		return b
	case v <= 0xFFFFFFFF:
		b := make([]byte, 5)
		b[0] = 0xFE
		binary.LittleEndian.PutUint32(b[1:], uint32(v))
		return b
	default:
		b := make([]byte, 9)
		b[0] = 0xFF
		binary.LittleEndian.PutUint64(b[1:], v)
		return b
	}
}
