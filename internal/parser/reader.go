package parser

import (
	"encoding/binary"
	"fmt"
)

// Reader wraps a byte slice with read cursor for sequential binary parsing
type Reader struct {
	data   []byte
	offset int
}

// NewReader creates a new Reader
func NewReader(data []byte) *Reader {
	return &Reader{data: data, offset: 0}
}

// Offset returns the current read position
func (r *Reader) Offset() int {
	return r.offset
}

// Remaining returns bytes left
func (r *Reader) Remaining() int {
	return len(r.data) - r.offset
}

// ReadByte reads a single byte
func (r *Reader) ReadByte() (byte, error) {
	if r.offset >= len(r.data) {
		return 0, fmt.Errorf("unexpected end of data at offset %d", r.offset)
	}
	b := r.data[r.offset]
	r.offset++
	return b, nil
}

// ReadBytes reads n bytes (makes a copy)
func (r *Reader) ReadBytes(n int) ([]byte, error) {
	if r.offset+n > len(r.data) {
		return nil, fmt.Errorf("unexpected end of data: need %d bytes at offset %d, have %d", n, r.offset, len(r.data)-r.offset)
	}
	result := make([]byte, n)
	copy(result, r.data[r.offset:r.offset+n])
	r.offset += n
	return result, nil
}

// ReadBytesRef reads n bytes WITHOUT copying — returns a slice of the underlying data.
func (r *Reader) ReadBytesRef(n int) ([]byte, error) {
	if r.offset+n > len(r.data) {
		return nil, fmt.Errorf("unexpected end of data: need %d bytes at offset %d, have %d", n, r.offset, len(r.data)-r.offset)
	}
	result := r.data[r.offset : r.offset+n]
	r.offset += n
	return result, nil
}

// ReadUint16LE reads a uint16 in little-endian
func (r *Reader) ReadUint16LE() (uint16, error) {
	b, err := r.ReadBytes(2)
	if err != nil {
		return 0, err
	}
	return binary.LittleEndian.Uint16(b), nil
}

// ReadUint32LE reads a uint32 in little-endian
func (r *Reader) ReadUint32LE() (uint32, error) {
	b, err := r.ReadBytes(4)
	if err != nil {
		return 0, err
	}
	return binary.LittleEndian.Uint32(b), nil
}

// ReadUint64LE reads a uint64 in little-endian
func (r *Reader) ReadUint64LE() (uint64, error) {
	b, err := r.ReadBytes(8)
	if err != nil {
		return 0, err
	}
	return binary.LittleEndian.Uint64(b), nil
}

// ReadVarInt reads a Bitcoin CompactSize unsigned integer
func (r *Reader) ReadVarInt() (uint64, error) {
	first, err := r.ReadByte()
	if err != nil {
		return 0, err
	}
	switch {
	case first < 0xFD:
		return uint64(first), nil
	case first == 0xFD:
		val, err := r.ReadUint16LE()
		return uint64(val), err
	case first == 0xFE:
		val, err := r.ReadUint32LE()
		return uint64(val), err
	default: // 0xFF
		return r.ReadUint64LE()
	}
}

// Slice returns the underlying data from start to end offset
func (r *Reader) Slice(start, end int) []byte {
	return r.data[start:end]
}

// Data returns the full underlying data
func (r *Reader) Data() []byte {
	return r.data
}

// SetOffset sets the current read position
func (r *Reader) SetOffset(offset int) {
	r.offset = offset
}

// ReadCoreVarInt reads a Bitcoin Core VARINT (7-bit MSB continuation encoding).
// This is different from CompactSize (ReadVarInt).
func (r *Reader) ReadCoreVarInt() (uint64, error) {
	var n uint64
	for {
		b, err := r.ReadByte()
		if err != nil {
			return 0, err
		}
		n = (n << 7) | uint64(b&0x7F)
		if b&0x80 == 0 {
			return n, nil
		}
		n++
	}
}
