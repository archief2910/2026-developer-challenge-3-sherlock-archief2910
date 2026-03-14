package block

import (
	"encoding/binary"
	"encoding/hex"
	"fmt"
	"math"
	"os"

	lcrypto "sherlock/internal/crypto"
	"sherlock/internal/parser"
	"sherlock/internal/script"
	"sherlock/internal/types"
)

// BlockHeader represents an 80-byte Bitcoin block header
type BlockHeader struct {
	Version       int32
	PrevBlockHash string
	MerkleRoot    string
	Timestamp     uint32
	Bits          string
	Nonce         uint32
	BlockHash     string
}

// UndoData holds the prevout information from rev*.dat for one transaction
type UndoData struct {
	Prevouts []types.FixturePrevout
}

// ParsedBlock holds a fully parsed block with all data needed for analysis
type ParsedBlock struct {
	Header       BlockHeader
	Height       int64
	Transactions []*ParsedTransaction
}

// ParsedTransaction holds a parsed transaction with computed fields
type ParsedTransaction struct {
	Txid         string
	Raw          *types.RawTransaction
	Prevouts     []types.FixturePrevout // from undo data
	IsCoinbase   bool
	FeeSats      int64
	FeeRateSatVb float64
	Weight       int
	Vbytes       int
	// Classified outputs and inputs
	InputScriptTypes  []string
	OutputScriptTypes []string
	InputAddresses    []*string
	OutputAddresses   []*string
	InputValues       []int64
	OutputValues      []int64
	TotalInputSats    int64
	TotalOutputSats   int64
}

// RawBlock is unparsed block data from blk*.dat
type RawBlock struct {
	HeaderBytes [80]byte
	TxData      []byte
	FullData    []byte
}

// ProcessBlockFiles reads and parses block files, returning parsed blocks
func ProcessBlockFiles(blkPath, revPath, xorPath string) ([]ParsedBlock, error) {
	// Read XOR key
	xorKey, err := os.ReadFile(xorPath)
	if err != nil {
		return nil, fmt.Errorf("reading XOR key: %w", err)
	}

	// Read and XOR-decode block data
	blkData, err := os.ReadFile(blkPath)
	if err != nil {
		return nil, fmt.Errorf("reading block data: %w", err)
	}
	xorDecode(blkData, xorKey)

	// Read and XOR-decode undo data
	revData, err := os.ReadFile(revPath)
	if err != nil {
		return nil, fmt.Errorf("reading undo data: %w", err)
	}
	xorDecode(revData, xorKey)

	// Parse raw blocks from blk data
	rawBlocks, err := parseBlocks(blkData)
	if err != nil {
		return nil, fmt.Errorf("parsing blocks: %w", err)
	}

	// Parse undo data for all blocks
	undos, err := parseUndoData(revData, rawBlocks)
	if err != nil {
		return nil, fmt.Errorf("parsing undo data: %w", err)
	}

	fmt.Fprintf(os.Stderr, "Parsing %d blocks...\n", len(rawBlocks))

	// Analyze each block
	parsedBlocks := make([]ParsedBlock, len(rawBlocks))
	for i, raw := range rawBlocks {
		var undo []UndoData
		if i < len(undos) {
			undo = undos[i]
		}

		parsed, err := analyzeBlockData(raw, undo)
		if err != nil {
			return nil, fmt.Errorf("analyzing block %d: %w", i, err)
		}
		parsedBlocks[i] = *parsed

		if (i+1)%10 == 0 || i == len(rawBlocks)-1 {
			fmt.Fprintf(os.Stderr, "  Block %d/%d done\n", i+1, len(rawBlocks))
		}
	}

	return parsedBlocks, nil
}

func xorDecode(data []byte, key []byte) {
	if len(key) == 0 {
		return
	}
	for i := range data {
		data[i] ^= key[i%len(key)]
	}
}

func parseBlocks(data []byte) ([]RawBlock, error) {
	var blocks []RawBlock
	r := parser.NewReader(data)

	for r.Remaining() >= 8 {
		// Magic number (4 bytes)
		_, err := r.ReadUint32LE()
		if err != nil {
			break
		}

		// Block size (4 bytes)
		blockSize, err := r.ReadUint32LE()
		if err != nil {
			break
		}

		if int(blockSize) > r.Remaining() {
			break
		}

		blockData, err := r.ReadBytes(int(blockSize))
		if err != nil {
			break
		}

		var block RawBlock
		copy(block.HeaderBytes[:], blockData[:80])
		block.TxData = blockData[80:]
		block.FullData = blockData

		blocks = append(blocks, block)
	}

	if len(blocks) == 0 {
		return nil, fmt.Errorf("no blocks found in file")
	}

	return blocks, nil
}

func parseBlockHeader(headerBytes [80]byte) BlockHeader {
	version := int32(binary.LittleEndian.Uint32(headerBytes[0:4]))
	prevHash := lcrypto.ReverseBytes(headerBytes[4:36])
	merkleRoot := lcrypto.ReverseBytes(headerBytes[36:68])
	timestamp := binary.LittleEndian.Uint32(headerBytes[68:72])
	bits := binary.LittleEndian.Uint32(headerBytes[72:76])
	nonce := binary.LittleEndian.Uint32(headerBytes[76:80])

	blockHash := lcrypto.DoubleSHA256(headerBytes[:])
	blockHashHex := hex.EncodeToString(lcrypto.ReverseBytes(blockHash[:]))

	return BlockHeader{
		Version:       version,
		PrevBlockHash: hex.EncodeToString(prevHash),
		MerkleRoot:    hex.EncodeToString(merkleRoot),
		Timestamp:     timestamp,
		Bits:          fmt.Sprintf("%08x", bits),
		Nonce:         nonce,
		BlockHash:     blockHashHex,
	}
}

func parseBlockTransactions(txData []byte) ([]*types.RawTransaction, [][]byte, error) {
	r := parser.NewReader(txData)

	txCount, err := r.ReadVarInt()
	if err != nil {
		return nil, nil, fmt.Errorf("reading tx count: %w", err)
	}

	txs := make([]*types.RawTransaction, txCount)
	legacySers := make([][]byte, txCount)

	for i := uint64(0); i < txCount; i++ {
		startOffset := r.Offset()
		remainingData := r.Data()[startOffset:]

		tx, legSer, _, consumed, err := parser.ParseTransactionBytes(remainingData)
		if err != nil {
			return nil, nil, fmt.Errorf("parsing transaction %d: %w", i, err)
		}

		txs[i] = tx
		legacySers[i] = legSer
		r.SetOffset(startOffset + consumed)
	}

	return txs, legacySers, nil
}

func decodeBIP34Height(scriptSig []byte) int64 {
	if len(scriptSig) == 0 {
		return -1
	}

	numBytes := int(scriptSig[0])
	if numBytes == 0 || numBytes > 4 || numBytes+1 > len(scriptSig) {
		return -1
	}

	var height int64
	for i := 0; i < numBytes; i++ {
		height |= int64(scriptSig[1+i]) << uint(8*i)
	}

	return height
}

func parseUndoData(data []byte, blocks []RawBlock) ([][]UndoData, error) {
	r := parser.NewReader(data)
	allUndos := make([][]UndoData, len(blocks))

	for blockIdx := range blocks {
		if r.Remaining() < 8 {
			return nil, fmt.Errorf("not enough data for block undo %d header", blockIdx)
		}
		_, err := r.ReadUint32LE() // magic
		if err != nil {
			return nil, fmt.Errorf("reading undo magic for block %d: %w", blockIdx, err)
		}
		_, err = r.ReadUint32LE() // size
		if err != nil {
			return nil, fmt.Errorf("reading undo size for block %d: %w", blockIdx, err)
		}

		numTxUndos, err := r.ReadVarInt()
		if err != nil {
			return nil, fmt.Errorf("reading numTxUndos for block %d: %w", blockIdx, err)
		}

		// blockUndos[0] = coinbase (no undo data), [1..n] = non-coinbase
		blockUndos := make([]UndoData, numTxUndos+1)

		for txIdx := uint64(0); txIdx < numTxUndos; txIdx++ {
			numCoins, err := r.ReadVarInt()
			if err != nil {
				return nil, fmt.Errorf("block %d, tx %d: reading numCoins: %w", blockIdx, txIdx, err)
			}

			prevouts := make([]types.FixturePrevout, numCoins)
			for inIdx := uint64(0); inIdx < numCoins; inIdx++ {
				prevout, err := readCompressedPrevout(r)
				if err != nil {
					return nil, fmt.Errorf("block %d, tx %d, input %d: %w", blockIdx, txIdx, inIdx, err)
				}
				prevouts[inIdx] = prevout
			}

			blockUndos[txIdx+1] = UndoData{Prevouts: prevouts}
		}

		// Skip the 32-byte checksum hash
		if r.Remaining() >= 32 {
			r.ReadBytes(32)
		}

		allUndos[blockIdx] = blockUndos
	}

	return allUndos, nil
}

func readCompressedPrevout(r *parser.Reader) (types.FixturePrevout, error) {
	var prevout types.FixturePrevout

	code, err := r.ReadCoreVarInt()
	if err != nil {
		return prevout, fmt.Errorf("reading height/coinbase: %w", err)
	}
	height := code >> 1

	if height > 0 {
		_, err = r.ReadCoreVarInt() // nVersionDummy
		if err != nil {
			return prevout, fmt.Errorf("reading version dummy: %w", err)
		}
	}

	compressedValue, err := r.ReadCoreVarInt()
	if err != nil {
		return prevout, fmt.Errorf("reading compressed value: %w", err)
	}
	prevout.ValueSats = decompressAmount(compressedValue)

	scriptBytes, err := readCompressedScript(r)
	if err != nil {
		return prevout, fmt.Errorf("reading script: %w", err)
	}
	prevout.ScriptPubKeyHex = hex.EncodeToString(scriptBytes)

	return prevout, nil
}

func decompressAmount(x uint64) int64 {
	if x == 0 {
		return 0
	}
	x--
	e := x % 10
	x /= 10
	var n uint64
	if e < 9 {
		d := (x % 9) + 1
		x /= 9
		n = x*10 + d
	} else {
		n = x + 1
	}
	for e > 0 {
		n *= 10
		e--
	}
	return int64(n)
}

func readCompressedScript(r *parser.Reader) ([]byte, error) {
	nSize, err := r.ReadCoreVarInt()
	if err != nil {
		return nil, fmt.Errorf("reading script nSize: %w", err)
	}

	switch nSize {
	case 0: // P2PKH
		hash, err := r.ReadBytes(20)
		if err != nil {
			return nil, err
		}
		s := make([]byte, 25)
		s[0] = 0x76
		s[1] = 0xa9
		s[2] = 0x14
		copy(s[3:23], hash)
		s[23] = 0x88
		s[24] = 0xac
		return s, nil

	case 1: // P2SH
		hash, err := r.ReadBytes(20)
		if err != nil {
			return nil, err
		}
		s := make([]byte, 23)
		s[0] = 0xa9
		s[1] = 0x14
		copy(s[2:22], hash)
		s[22] = 0x87
		return s, nil

	case 2, 3: // Compressed P2PK
		key, err := r.ReadBytes(32)
		if err != nil {
			return nil, err
		}
		s := make([]byte, 35)
		s[0] = 0x21
		s[1] = byte(nSize)
		copy(s[2:34], key)
		s[34] = 0xac
		return s, nil

	case 4, 5: // Uncompressed P2PK
		key, err := r.ReadBytes(32)
		if err != nil {
			return nil, err
		}
		compressedPrefix := byte(nSize - 2)
		s := make([]byte, 35)
		s[0] = 0x21
		s[1] = compressedPrefix
		copy(s[2:34], key)
		s[34] = 0xac
		return s, nil

	default: // nSize >= 6: raw script
		scriptLen := int(nSize) - 6
		if scriptLen < 0 || scriptLen > 10000 {
			return nil, fmt.Errorf("invalid script length %d (nSize=%d)", scriptLen, nSize)
		}
		scriptData, err := r.ReadBytes(scriptLen)
		if err != nil {
			return nil, err
		}
		return scriptData, nil
	}
}

// analyzeBlockData parses a raw block with undo data and returns a ParsedBlock
func analyzeBlockData(raw RawBlock, undos []UndoData) (*ParsedBlock, error) {
	header := parseBlockHeader(raw.HeaderBytes)

	txs, legacySers, err := parseBlockTransactions(raw.TxData)
	if err != nil {
		return nil, fmt.Errorf("parsing transactions: %w", err)
	}

	// Compute txids
	parsedTxs := make([]*ParsedTransaction, len(txs))
	for i, tx := range txs {
		txidHash := lcrypto.DoubleSHA256(legacySers[i])
		txidHex := hex.EncodeToString(lcrypto.ReverseBytes(txidHash[:]))

		isCoinbase := i == 0
		weight := tx.NonWitnessBytes*4 + tx.WitnessBytes
		vbytes := int(math.Ceil(float64(weight) / 4.0))

		pt := &ParsedTransaction{
			Txid:       txidHex,
			Raw:        tx,
			IsCoinbase: isCoinbase,
			Weight:     weight,
			Vbytes:     vbytes,
		}

		// Classify outputs
		pt.OutputScriptTypes = make([]string, len(tx.Outputs))
		pt.OutputAddresses = make([]*string, len(tx.Outputs))
		pt.OutputValues = make([]int64, len(tx.Outputs))
		for j, out := range tx.Outputs {
			pt.OutputScriptTypes[j] = script.ClassifyOutput(out.ScriptPubKey)
			// For address, convert bytes to hex and derive
			addrHex := hex.EncodeToString(out.ScriptPubKey)
			pt.OutputAddresses[j] = deriveAddr(addrHex)
			pt.OutputValues[j] = out.Value
			pt.TotalOutputSats += out.Value
		}

		// Handle inputs with prevout data from undo
		pt.InputScriptTypes = make([]string, len(tx.Inputs))
		pt.InputAddresses = make([]*string, len(tx.Inputs))
		pt.InputValues = make([]int64, len(tx.Inputs))

		if isCoinbase {
			// Coinbase has no real prevouts
			pt.Prevouts = nil
		} else if i < len(undos) && undos[i].Prevouts != nil {
			pt.Prevouts = undos[i].Prevouts
			for j := range tx.Inputs {
				if j < len(undos[i].Prevouts) {
					fp := undos[i].Prevouts[j]
					prevScript, _ := hex.DecodeString(fp.ScriptPubKeyHex)
					pt.InputScriptTypes[j] = script.ClassifyOutput(prevScript)
					pt.InputAddresses[j] = deriveAddr(fp.ScriptPubKeyHex)
					pt.InputValues[j] = fp.ValueSats
					pt.TotalInputSats += fp.ValueSats
				}
			}
		}

		// Compute fee
		if !isCoinbase {
			pt.FeeSats = pt.TotalInputSats - pt.TotalOutputSats
			if vbytes > 0 {
				rawFeeRate := math.Round(float64(pt.FeeSats)/float64(vbytes)*100) / 100
				pt.FeeRateSatVb = rawFeeRate
				// Don't cap here - let the aggregator handle outliers
				// Reference: Fee rates will be post-processed to handle extreme outliers
			}
		}

		parsedTxs[i] = pt
	}

	// Extract block height from coinbase
	height := int64(-1)
	if len(txs) > 0 && len(txs[0].Inputs) > 0 {
		height = decodeBIP34Height(txs[0].Inputs[0].ScriptSig)
	}

	return &ParsedBlock{
		Header:       header,
		Height:       height,
		Transactions: parsedTxs,
	}, nil
}

// deriveAddr is a helper to convert hex script to address
func deriveAddr(scriptHex string) *string {
	if scriptHex == "" {
		return nil
	}
	scriptBytes, err := hex.DecodeString(scriptHex)
	if err != nil {
		return nil
	}
	return deriveAddressFromBytes(scriptBytes)
}

// deriveAddressFromBytes derives an address from raw script bytes
func deriveAddressFromBytes(script []byte) *string {
	n := len(script)

	// P2PKH
	if n == 25 && script[0] == 0x76 && script[1] == 0xa9 && script[2] == 0x14 &&
		script[23] == 0x88 && script[24] == 0xac {
		// Use hash directly
		addr := "p2pkh:" + hex.EncodeToString(script[3:23])
		return &addr
	}

	// P2SH
	if n == 23 && script[0] == 0xa9 && script[1] == 0x14 && script[22] == 0x87 {
		addr := "p2sh:" + hex.EncodeToString(script[2:22])
		return &addr
	}

	// P2WPKH
	if n == 22 && script[0] == 0x00 && script[1] == 0x14 {
		addr := "p2wpkh:" + hex.EncodeToString(script[2:22])
		return &addr
	}

	// P2WSH
	if n == 34 && script[0] == 0x00 && script[1] == 0x20 {
		addr := "p2wsh:" + hex.EncodeToString(script[2:34])
		return &addr
	}

	// P2TR
	if n == 34 && script[0] == 0x51 && script[1] == 0x20 {
		addr := "p2tr:" + hex.EncodeToString(script[2:34])
		return &addr
	}

	return nil
}
