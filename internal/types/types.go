package types

// FixturePrevout holds value + script from undo data
type FixturePrevout struct {
	Txid            string `json:"txid"`
	Vout            uint32 `json:"vout"`
	ValueSats       int64  `json:"value_sats"`
	ScriptPubKeyHex string `json:"script_pubkey_hex"`
}

// RawTransaction is a parsed raw Bitcoin transaction
type RawTransaction struct {
	Version   uint32
	Inputs    []TxInput
	Outputs   []TxOutput
	Witnesses [][]WitnessItem
	Locktime  uint32
	IsSegwit  bool

	// Byte tracking for weight
	TotalBytes      int
	WitnessBytes    int
	NonWitnessBytes int
}

// TxInput is a parsed transaction input
type TxInput struct {
	Txid      [32]byte
	Vout      uint32
	ScriptSig []byte
	Sequence  uint32
}

// TxOutput is a parsed transaction output
type TxOutput struct {
	Value        int64
	ScriptPubKey []byte
}

// WitnessItem is a single witness stack item
type WitnessItem = []byte
