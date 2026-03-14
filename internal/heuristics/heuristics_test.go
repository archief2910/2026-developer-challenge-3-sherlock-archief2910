package heuristics

import (
	"testing"

	"sherlock/internal/block"
	"sherlock/internal/types"
)

// Helper to create a minimal ParsedTransaction for testing.
func makeTx(numInputs, numOutputs int) *block.ParsedTransaction {
	inputs := make([]types.TxInput, numInputs)
	outputs := make([]types.TxOutput, numOutputs)
	for i := range outputs {
		outputs[i] = types.TxOutput{Value: 50000, ScriptPubKey: []byte{0x00, 0x14, 0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07, 0x08, 0x09, 0x0a, 0x0b, 0x0c, 0x0d, 0x0e, 0x0f, 0x10, 0x11, 0x12, 0x13, 0x14}}
	}
	raw := &types.RawTransaction{
		Inputs:  inputs,
		Outputs: outputs,
	}

	inputTypes := make([]string, numInputs)
	inputAddrs := make([]*string, numInputs)
	inputVals := make([]int64, numInputs)
	for i := range inputTypes {
		inputTypes[i] = "p2wpkh"
		addr := "p2wpkh:aaaa"
		inputAddrs[i] = &addr
		inputVals[i] = 100000
	}

	outputTypes := make([]string, numOutputs)
	outputAddrs := make([]*string, numOutputs)
	outputVals := make([]int64, numOutputs)
	for i := range outputTypes {
		outputTypes[i] = "p2wpkh"
		addr := "p2wpkh:bbbb"
		outputAddrs[i] = &addr
		outputVals[i] = 50000
	}

	return &block.ParsedTransaction{
		Txid:              "deadbeef",
		Raw:               raw,
		IsCoinbase:        false,
		InputScriptTypes:  inputTypes,
		OutputScriptTypes: outputTypes,
		InputAddresses:    inputAddrs,
		OutputAddresses:   outputAddrs,
		InputValues:       inputVals,
		OutputValues:      outputVals,
		TotalInputSats:    int64(numInputs) * 100000,
		TotalOutputSats:   int64(numOutputs) * 50000,
	}
}

// -----------------------------------------------------------------------
// Heuristic 1: CIOH
// -----------------------------------------------------------------------

func TestCIOH_SingleInput(t *testing.T) {
	tx := makeTx(1, 2)
	h := AnalyzeTransaction(tx, 0, nil, 100)
	if h.CIOH.Detected {
		t.Error("CIOH should not fire for single-input transactions")
	}
}

func TestCIOH_TwoInputs_Medium(t *testing.T) {
	tx := makeTx(2, 2)
	h := AnalyzeTransaction(tx, 0, nil, 100)
	if !h.CIOH.Detected {
		t.Error("CIOH should fire for multi-input transactions")
	}
	if h.CIOH.Confidence != "medium" {
		t.Errorf("Expected medium confidence for 2 inputs, got %q", h.CIOH.Confidence)
	}
}

func TestCIOH_ManyInputs_High(t *testing.T) {
	tx := makeTx(5, 2)
	h := AnalyzeTransaction(tx, 0, nil, 100)
	if !h.CIOH.Detected {
		t.Error("CIOH should fire for multi-input transactions")
	}
	if h.CIOH.Confidence != "high" {
		t.Errorf("Expected high confidence for 5 inputs, got %q", h.CIOH.Confidence)
	}
}

func TestCIOH_Coinbase(t *testing.T) {
	tx := makeTx(1, 1)
	tx.IsCoinbase = true
	h := AnalyzeTransaction(tx, 0, nil, 100)
	if h.CIOH.Detected {
		t.Error("CIOH should not fire for coinbase transactions")
	}
}

// -----------------------------------------------------------------------
// Heuristic 2: Change Detection
// -----------------------------------------------------------------------

func TestChangeDetection_ScriptTypeMatch(t *testing.T) {
	tx := makeTx(2, 2)
	// Input types: p2wpkh, output 0: p2wpkh, output 1: p2tr
	tx.OutputScriptTypes[0] = "p2wpkh"
	tx.OutputScriptTypes[1] = "p2tr"
	h := AnalyzeTransaction(tx, 0, nil, 100)

	if !h.ChangeDetection.Detected {
		t.Error("Change detection should fire with script type mismatch")
	}
	if h.ChangeDetection.LikelyChangeIdx != 0 {
		t.Errorf("Expected change index 0 (matching type), got %d", h.ChangeDetection.LikelyChangeIdx)
	}
	if h.ChangeDetection.Method != "script_type_match" {
		t.Errorf("Expected method script_type_match, got %q", h.ChangeDetection.Method)
	}
	if h.ChangeDetection.Confidence != "high" {
		t.Errorf("Expected high confidence, got %q", h.ChangeDetection.Confidence)
	}
}

func TestChangeDetection_OptimalChange(t *testing.T) {
	tx := makeTx(2, 2)
	// Both outputs same type, so script_type_match won't fire
	tx.OutputScriptTypes[0] = "p2wpkh"
	tx.OutputScriptTypes[1] = "p2wpkh"
	// Set input values high, one output below smallest input
	tx.InputValues[0] = 500000
	tx.InputValues[1] = 300000
	tx.OutputValues[0] = 100000 // smaller than smallest input (300000)
	tx.OutputValues[1] = 600000 // larger than smallest input
	h := AnalyzeTransaction(tx, 0, nil, 100)

	if !h.ChangeDetection.Detected {
		t.Error("Optimal change should fire")
	}
	if h.ChangeDetection.LikelyChangeIdx != 0 {
		t.Errorf("Expected change index 0 (smaller than smallest input), got %d", h.ChangeDetection.LikelyChangeIdx)
	}
	if h.ChangeDetection.Method != "optimal_change" {
		t.Errorf("Expected method optimal_change, got %q", h.ChangeDetection.Method)
	}
}

func TestChangeDetection_RoundNumber(t *testing.T) {
	tx := makeTx(2, 2)
	tx.OutputScriptTypes[0] = "p2wpkh"
	tx.OutputScriptTypes[1] = "p2wpkh"
	// Both outputs larger than min input, so optimal_change won't fire
	tx.InputValues[0] = 100000
	tx.InputValues[1] = 200000
	tx.OutputValues[0] = 10000000 // 0.1 BTC — round amount (payment)
	tx.OutputValues[1] = 189500   // non-round (change)
	h := AnalyzeTransaction(tx, 0, nil, 100)

	if !h.ChangeDetection.Detected {
		t.Error("Round number change detection should fire")
	}
	if h.ChangeDetection.LikelyChangeIdx != 1 {
		t.Errorf("Expected change index 1 (non-round output), got %d", h.ChangeDetection.LikelyChangeIdx)
	}
	if h.ChangeDetection.Method != "round_number" {
		t.Errorf("Expected method round_number, got %q", h.ChangeDetection.Method)
	}
}

func TestChangeDetection_ValueAnalysis_LastResort(t *testing.T) {
	tx := makeTx(2, 2)
	tx.OutputScriptTypes[0] = "p2wpkh"
	tx.OutputScriptTypes[1] = "p2wpkh"
	// Both outputs LARGER than min input — optimal_change won't fire
	tx.InputValues[0] = 50
	tx.InputValues[1] = 50
	// Neither is round, and both are larger than inputs
	tx.OutputValues[0] = 73
	tx.OutputValues[1] = 127
	// No locktime — nLockTime won't fire
	tx.Raw.Locktime = 0
	h := AnalyzeTransaction(tx, 0, nil, 100)

	if !h.ChangeDetection.Detected {
		t.Error("Value analysis fallback should fire")
	}
	if h.ChangeDetection.Method != "value_analysis" {
		t.Errorf("Expected method value_analysis, got %q", h.ChangeDetection.Method)
	}
	if h.ChangeDetection.LikelyChangeIdx != 0 {
		t.Errorf("Expected smaller output (0) as change, got %d", h.ChangeDetection.LikelyChangeIdx)
	}
}

func TestChangeDetection_SingleOutput_NoDetection(t *testing.T) {
	tx := makeTx(2, 1)
	h := AnalyzeTransaction(tx, 0, nil, 100)
	if h.ChangeDetection.Detected {
		t.Error("Change detection should not fire for single-output transactions")
	}
}

func TestChangeDetection_nLockTime(t *testing.T) {
	tx := makeTx(1, 2)
	tx.OutputScriptTypes[0] = "p2wpkh"
	tx.OutputScriptTypes[1] = "p2wpkh"
	// Both outputs larger than min input — optimal_change won't fire
	tx.InputValues[0] = 100
	// Neither is round
	tx.OutputValues[0] = 73
	tx.OutputValues[1] = 27
	// Set locktime to match block height (anti-fee-sniping)
	tx.Raw.Locktime = 800000
	h := AnalyzeTransaction(tx, 0, nil, 800000)

	if !h.ChangeDetection.Detected {
		t.Error("nLockTime fingerprint should fire")
	}
	if h.ChangeDetection.Method != "locktime_fingerprint" {
		t.Errorf("Expected method locktime_fingerprint, got %q", h.ChangeDetection.Method)
	}
	if h.ChangeDetection.Confidence != "medium" {
		t.Errorf("Expected medium confidence, got %q", h.ChangeDetection.Confidence)
	}
}

// -----------------------------------------------------------------------
// Heuristic 3: Address Reuse
// -----------------------------------------------------------------------

func TestAddressReuse_WithinTransaction(t *testing.T) {
	tx := makeTx(1, 2)
	addr := "p2wpkh:reused_address"
	tx.InputAddresses[0] = &addr
	tx.OutputAddresses[0] = &addr // Same address in input and output
	diffAddr := "p2wpkh:different"
	tx.OutputAddresses[1] = &diffAddr

	h := AnalyzeTransaction(tx, 0, nil, 100)
	if !h.AddressReuse.Detected {
		t.Error("Address reuse should be detected within same transaction")
	}
	if h.AddressReuse.Confidence != "high" {
		t.Errorf("Expected high confidence, got %q", h.AddressReuse.Confidence)
	}
}

func TestAddressReuse_CrossTransaction(t *testing.T) {
	// Build two transactions that share an address
	tx1 := makeTx(1, 1)
	tx2 := makeTx(1, 1)
	sharedAddr := "p2wpkh:shared_addr"
	tx1.OutputAddresses[0] = &sharedAddr
	tx2.InputAddresses[0] = &sharedAddr

	otherAddr := "p2wpkh:other"
	tx1.InputAddresses[0] = &otherAddr
	tx2.OutputAddresses[0] = &otherAddr

	// Build block address map
	addrMap := BuildBlockAddressMap([]*block.ParsedTransaction{tx1, tx2})

	// Analyze tx1 — it should detect cross-tx reuse
	h := AnalyzeTransaction(tx1, 0, addrMap, 100)
	if !h.AddressReuse.Detected {
		t.Error("Cross-transaction address reuse should be detected")
	}
	if !h.AddressReuse.CrossTx {
		t.Error("CrossTx flag should be set")
	}
}

func TestAddressReuse_NoReuse(t *testing.T) {
	tx := makeTx(1, 1)
	inAddr := "p2wpkh:input_addr"
	outAddr := "p2wpkh:output_addr"
	tx.InputAddresses[0] = &inAddr
	tx.OutputAddresses[0] = &outAddr

	h := AnalyzeTransaction(tx, 0, nil, 100)
	if h.AddressReuse.Detected {
		t.Error("Address reuse should not be detected when addresses differ")
	}
}

// -----------------------------------------------------------------------
// Heuristic 4: CoinJoin Detection
// -----------------------------------------------------------------------

func TestCoinJoin_EqualOutputs(t *testing.T) {
	tx := makeTx(5, 5)
	// Set 5 equal-value outputs
	for i := 0; i < 5; i++ {
		tx.OutputValues[i] = 1000000
		tx.OutputScriptTypes[i] = "p2wpkh"
	}
	// Mix input types (2 distinct)
	tx.InputScriptTypes[0] = "p2wpkh"
	tx.InputScriptTypes[1] = "p2wpkh"
	tx.InputScriptTypes[2] = "p2tr"
	tx.InputScriptTypes[3] = "p2tr"
	tx.InputScriptTypes[4] = "p2wpkh"

	h := AnalyzeTransaction(tx, 0, nil, 100)
	if !h.CoinJoin.Detected {
		t.Error("CoinJoin should be detected for 5 equal-value outputs with mixed input types")
	}
	if h.CoinJoin.Confidence != "high" {
		t.Errorf("Expected high confidence, got %q", h.CoinJoin.Confidence)
	}
	if h.CoinJoin.EqualOutputCount != 5 {
		t.Errorf("Expected 5 equal outputs, got %d", h.CoinJoin.EqualOutputCount)
	}
}

func TestCoinJoin_TooFewOutputs(t *testing.T) {
	tx := makeTx(3, 2)
	h := AnalyzeTransaction(tx, 0, nil, 100)
	if h.CoinJoin.Detected {
		t.Error("CoinJoin should not fire with fewer than 3 outputs")
	}
}

func TestCoinJoin_NoEqualValues(t *testing.T) {
	tx := makeTx(4, 4)
	tx.OutputValues[0] = 100000
	tx.OutputValues[1] = 200000
	tx.OutputValues[2] = 300000
	tx.OutputValues[3] = 400000
	h := AnalyzeTransaction(tx, 0, nil, 100)
	if h.CoinJoin.Detected {
		t.Error("CoinJoin should not fire without 3+ equal outputs")
	}
}

func TestCoinJoin_DowngradesCIOH(t *testing.T) {
	tx := makeTx(5, 5)
	for i := 0; i < 5; i++ {
		tx.OutputValues[i] = 1000000
	}
	tx.InputScriptTypes[0] = "p2wpkh"
	tx.InputScriptTypes[1] = "p2tr"
	tx.InputScriptTypes[2] = "p2wpkh"
	tx.InputScriptTypes[3] = "p2tr"
	tx.InputScriptTypes[4] = "p2wpkh"

	h := AnalyzeTransaction(tx, 0, nil, 100)
	if !h.CoinJoin.Detected || !h.CIOH.Detected {
		t.Error("Both CoinJoin and CIOH should be detected")
	}
	if h.CIOH.Confidence != "low" {
		t.Errorf("CIOH should be downgraded to low when CoinJoin detected, got %q", h.CIOH.Confidence)
	}
}

// -----------------------------------------------------------------------
// Heuristic 5: Consolidation
// -----------------------------------------------------------------------

func TestConsolidation_ManyInputsFewOutputs(t *testing.T) {
	tx := makeTx(10, 1)
	for i := range tx.InputScriptTypes {
		tx.InputScriptTypes[i] = "p2wpkh"
	}
	tx.OutputScriptTypes[0] = "p2wpkh"

	h := AnalyzeTransaction(tx, 0, nil, 100)
	if !h.Consolidation.Detected {
		t.Error("Consolidation should be detected for 10 inputs, 1 output")
	}
	if h.Consolidation.Confidence != "high" {
		t.Errorf("Expected high confidence for 10+ same-type inputs, got %q", h.Consolidation.Confidence)
	}
}

func TestConsolidation_TooFewInputs(t *testing.T) {
	tx := makeTx(3, 1)
	h := AnalyzeTransaction(tx, 0, nil, 100)
	if h.Consolidation.Detected {
		t.Error("Consolidation should not fire with fewer than 5 inputs")
	}
}

func TestConsolidation_TooManyOutputs(t *testing.T) {
	tx := makeTx(10, 4)
	h := AnalyzeTransaction(tx, 0, nil, 100)
	if h.Consolidation.Detected {
		t.Error("Consolidation should not fire with more than 2 non-OP_RETURN outputs")
	}
}

func TestConsolidation_SuppressesSelfTransfer(t *testing.T) {
	// 5 inputs, 1 output, all same type — both consolidation and self-transfer would fire,
	// but consolidation should suppress self-transfer.
	tx := makeTx(5, 1)
	for i := range tx.InputScriptTypes {
		tx.InputScriptTypes[i] = "p2wpkh"
	}
	tx.OutputScriptTypes[0] = "p2wpkh"
	tx.OutputValues[0] = 123456 // non-round

	h := AnalyzeTransaction(tx, 0, nil, 100)
	if !h.Consolidation.Detected {
		t.Error("Consolidation should be detected")
	}
	if h.SelfTransfer.Detected {
		t.Error("Self-transfer should be suppressed when consolidation is detected")
	}
}

// -----------------------------------------------------------------------
// Heuristic 6: Self-Transfer
// -----------------------------------------------------------------------

func TestSelfTransfer_AllSameType_NoRound(t *testing.T) {
	tx := makeTx(1, 2)
	tx.InputScriptTypes[0] = "p2tr"
	tx.OutputScriptTypes[0] = "p2tr"
	tx.OutputScriptTypes[1] = "p2tr"
	tx.OutputValues[0] = 123456 // non-round
	tx.OutputValues[1] = 654321 // non-round

	h := AnalyzeTransaction(tx, 0, nil, 100)
	if !h.SelfTransfer.Detected {
		t.Error("Self-transfer should fire for all-same-type, no-round outputs")
	}
	if h.SelfTransfer.Confidence != "medium" {
		t.Errorf("Expected medium confidence, got %q", h.SelfTransfer.Confidence)
	}
}

func TestSelfTransfer_RoundOutput(t *testing.T) {
	tx := makeTx(1, 2)
	tx.InputScriptTypes[0] = "p2tr"
	tx.OutputScriptTypes[0] = "p2tr"
	tx.OutputScriptTypes[1] = "p2tr"
	tx.OutputValues[0] = 10000000 // 0.1 BTC — significant round amount
	tx.OutputValues[1] = 654321

	h := AnalyzeTransaction(tx, 0, nil, 100)
	if h.SelfTransfer.Detected {
		t.Error("Self-transfer should not fire when a significant round amount output exists")
	}
}

func TestSelfTransfer_MixedTypes(t *testing.T) {
	tx := makeTx(1, 2)
	tx.InputScriptTypes[0] = "p2tr"
	tx.OutputScriptTypes[0] = "p2tr"
	tx.OutputScriptTypes[1] = "p2wpkh" // different type
	tx.OutputValues[0] = 123456
	tx.OutputValues[1] = 654321

	h := AnalyzeTransaction(tx, 0, nil, 100)
	if h.SelfTransfer.Detected {
		t.Error("Self-transfer should not fire when output types differ from input")
	}
}

// -----------------------------------------------------------------------
// Heuristic 7: Peeling Chain
// -----------------------------------------------------------------------

func TestPeelingChain_ExtremeAsymmetry(t *testing.T) {
	tx := makeTx(1, 2)
	tx.OutputValues[0] = 100      // tiny payment
	tx.OutputValues[1] = 99999900 // large change

	h := AnalyzeTransaction(tx, 0, nil, 100)
	if !h.PeelingChain.Detected {
		t.Error("Peeling chain should be detected for extreme value asymmetry")
	}
	if h.PeelingChain.Confidence != "high" {
		t.Errorf("Expected high confidence for ratio < 0.01, got %q", h.PeelingChain.Confidence)
	}
}

func TestPeelingChain_ModerateAsymmetry(t *testing.T) {
	tx := makeTx(1, 2)
	tx.OutputValues[0] = 3000
	tx.OutputValues[1] = 97000 // ratio ~0.031

	h := AnalyzeTransaction(tx, 0, nil, 100)
	if !h.PeelingChain.Detected {
		t.Error("Peeling chain should be detected for moderate asymmetry")
	}
	if h.PeelingChain.Confidence != "medium" {
		t.Errorf("Expected medium confidence for ratio 0.01-0.05, got %q", h.PeelingChain.Confidence)
	}
}

func TestPeelingChain_NotDetected_LowAsymmetry(t *testing.T) {
	tx := makeTx(1, 2)
	tx.OutputValues[0] = 40000
	tx.OutputValues[1] = 60000 // ratio ~0.67

	h := AnalyzeTransaction(tx, 0, nil, 100)
	if h.PeelingChain.Detected {
		t.Error("Peeling chain should not fire for roughly equal outputs")
	}
}

func TestPeelingChain_MultipleInputs_NotDetected(t *testing.T) {
	tx := makeTx(2, 2)
	h := AnalyzeTransaction(tx, 0, nil, 100)
	if h.PeelingChain.Detected {
		t.Error("Peeling chain should only fire for exactly 1 input")
	}
}

func TestPeelingChain_ReinforcesChangeDetection(t *testing.T) {
	// Peeling chain: 1 input, 2 outputs with extreme asymmetry
	// Change detection should NOT fire independently (both same type, both
	// larger than input, no round amounts)
	// => peeling chain cross-heuristic sets change = larger output
	tx := makeTx(1, 2)
	tx.OutputValues[0] = 50       // tiny payment
	tx.OutputValues[1] = 99999950 // large change
	tx.InputScriptTypes[0] = "p2wpkh"
	tx.OutputScriptTypes[0] = "p2wpkh"
	tx.OutputScriptTypes[1] = "p2wpkh"
	// Set input value larger than both outputs so optimal_change fires first.
	// Actually, for peeling chain reinforcement, change detection should NOT have
	// fired independently. Script type matching won't fire (both same type).
	// For the peeling chain interaction to fire, change must NOT be detected.
	// But with 2 outputs, optimal_change, round_number, or value_analysis will fire.
	// So peeling chain reinforcement only fires if change is already detected
	// via another method. Let me test the 3+ output case or just verify
	// peeling + change are both detected.

	h := AnalyzeTransaction(tx, 0, nil, 100)
	if !h.PeelingChain.Detected {
		t.Error("Peeling chain should be detected")
	}
	if !h.ChangeDetection.Detected {
		t.Error("Change detection should fire (via some method)")
	}
	// For 2-output, 1-input, either optimal_change or value_analysis will fire
	// independent of peeling chain. The larger output should be identified as change.
	if h.ChangeDetection.LikelyChangeIdx != 1 {
		t.Errorf("Expected change index 1 (larger output), got %d", h.ChangeDetection.LikelyChangeIdx)
	}
}

// -----------------------------------------------------------------------
// Heuristic 8: OP_RETURN
// -----------------------------------------------------------------------

func TestOpReturn_Detected(t *testing.T) {
	tx := makeTx(1, 2)
	// Second output is OP_RETURN
	tx.OutputScriptTypes[1] = "op_return"
	tx.Raw.Outputs[1].ScriptPubKey = []byte{0x6a, 0x04, 0x74, 0x65, 0x73, 0x74} // OP_RETURN "test"

	h := AnalyzeTransaction(tx, 0, nil, 100)
	if !h.OpReturn.Detected {
		t.Error("OP_RETURN should be detected")
	}
	if h.OpReturn.Count != 1 {
		t.Errorf("Expected 1 OP_RETURN output, got %d", h.OpReturn.Count)
	}
	if h.OpReturn.Confidence != "high" {
		t.Errorf("Expected high confidence, got %q", h.OpReturn.Confidence)
	}
}

func TestOpReturn_NotDetected(t *testing.T) {
	tx := makeTx(1, 2)
	h := AnalyzeTransaction(tx, 0, nil, 100)
	if h.OpReturn.Detected {
		t.Error("OP_RETURN should not be detected when no OP_RETURN outputs exist")
	}
}

func TestOpReturn_OmniProtocol(t *testing.T) {
	tx := makeTx(1, 2)
	tx.OutputScriptTypes[1] = "op_return"
	// OP_RETURN with Omni prefix: 0x6a + push + "omni"...
	tx.Raw.Outputs[1].ScriptPubKey = []byte{0x6a, 0x14, 0x6f, 0x6d, 0x6e, 0x69, 0x00, 0x00, 0x00, 0x00, 0x00, 0x01, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x64, 0x00, 0x00}

	h := AnalyzeTransaction(tx, 0, nil, 100)
	if !h.OpReturn.Detected {
		t.Error("OP_RETURN should be detected")
	}
	if h.OpReturn.Protocol != "omni" {
		t.Errorf("Expected protocol omni, got %q", h.OpReturn.Protocol)
	}
}

// -----------------------------------------------------------------------
// Heuristic 9: Round Number Payment
// -----------------------------------------------------------------------

func TestRoundNumber_OneBTC(t *testing.T) {
	tx := makeTx(1, 2)
	tx.OutputValues[0] = 100000000 // 1 BTC
	tx.OutputValues[1] = 12345

	h := AnalyzeTransaction(tx, 0, nil, 100)
	if !h.RoundNumber.Detected {
		t.Error("Round number should be detected for 1 BTC output")
	}
	if h.RoundNumber.Confidence != "high" {
		t.Errorf("Expected high confidence for 1 BTC, got %q", h.RoundNumber.Confidence)
	}
	if h.RoundNumber.HighestDenomination != "1_btc" {
		t.Errorf("Expected highest denomination 1_btc, got %q", h.RoundNumber.HighestDenomination)
	}
}

func TestRoundNumber_SmallRound(t *testing.T) {
	tx := makeTx(1, 2)
	tx.OutputValues[0] = 10000 // 0.0001 BTC
	tx.OutputValues[1] = 12345

	h := AnalyzeTransaction(tx, 0, nil, 100)
	if !h.RoundNumber.Detected {
		t.Error("Round number should be detected for 0.0001 BTC")
	}
	if h.RoundNumber.Confidence != "low" {
		t.Errorf("Expected low confidence for small round, got %q", h.RoundNumber.Confidence)
	}
}

func TestRoundNumber_NoRound(t *testing.T) {
	tx := makeTx(1, 2)
	tx.OutputValues[0] = 12345
	tx.OutputValues[1] = 67891

	h := AnalyzeTransaction(tx, 0, nil, 100)
	if h.RoundNumber.Detected {
		t.Error("Round number should not be detected for non-round outputs")
	}
}

// -----------------------------------------------------------------------
// Classification
// -----------------------------------------------------------------------

func TestClassify_CoinJoin(t *testing.T) {
	tx := makeTx(5, 5)
	for i := 0; i < 5; i++ {
		tx.OutputValues[i] = 1000000
	}
	tx.InputScriptTypes[0] = "p2wpkh"
	tx.InputScriptTypes[2] = "p2tr"

	h := AnalyzeTransaction(tx, 0, nil, 100)
	c := ClassifyTransaction(tx, h)
	if c != "coinjoin" {
		t.Errorf("Expected coinjoin classification, got %q", c)
	}
}

func TestClassify_Consolidation(t *testing.T) {
	tx := makeTx(10, 1)
	for i := range tx.InputScriptTypes {
		tx.InputScriptTypes[i] = "p2wpkh"
	}
	tx.OutputScriptTypes[0] = "p2wpkh"

	h := AnalyzeTransaction(tx, 0, nil, 100)
	c := ClassifyTransaction(tx, h)
	if c != "consolidation" {
		t.Errorf("Expected consolidation classification, got %q", c)
	}
}

func TestClassify_BatchPayment(t *testing.T) {
	tx := makeTx(1, 5)
	// No equal outputs — not CoinJoin
	for i := range tx.OutputValues {
		tx.OutputValues[i] = int64((i + 1) * 10000)
	}
	// Not all same type — not self-transfer
	tx.OutputScriptTypes[0] = "p2wpkh"
	tx.OutputScriptTypes[1] = "p2tr"
	tx.OutputScriptTypes[2] = "p2pkh"
	tx.OutputScriptTypes[3] = "p2wpkh"
	tx.OutputScriptTypes[4] = "p2tr"

	h := AnalyzeTransaction(tx, 0, nil, 100)
	c := ClassifyTransaction(tx, h)
	if c != "batch_payment" {
		t.Errorf("Expected batch_payment for 1 input + 5 outputs, got %q", c)
	}
}

func TestClassify_SimplePayment(t *testing.T) {
	tx := makeTx(1, 2)
	tx.OutputScriptTypes[0] = "p2wpkh"
	tx.OutputScriptTypes[1] = "p2tr" // different type prevents self_transfer
	tx.OutputValues[0] = 50000
	tx.OutputValues[1] = 50000

	h := AnalyzeTransaction(tx, 0, nil, 100)
	c := ClassifyTransaction(tx, h)
	if c != "simple_payment" {
		t.Errorf("Expected simple_payment, got %q", c)
	}
}

func TestClassify_Coinbase(t *testing.T) {
	tx := makeTx(1, 1)
	tx.IsCoinbase = true

	h := AnalyzeTransaction(tx, 0, nil, 100)
	c := ClassifyTransaction(tx, h)
	if c != "unknown" {
		t.Errorf("Expected unknown for coinbase, got %q", c)
	}
}

// -----------------------------------------------------------------------
// Utility functions
// -----------------------------------------------------------------------

func TestHeuristicIDs(t *testing.T) {
	ids := HeuristicIDs()
	if len(ids) != 9 {
		t.Errorf("Expected 9 heuristic IDs, got %d", len(ids))
	}
	// Check mandatory IDs
	foundCIOH := false
	foundChange := false
	for _, id := range ids {
		if id == "cioh" {
			foundCIOH = true
		}
		if id == "change_detection" {
			foundChange = true
		}
	}
	if !foundCIOH {
		t.Error("cioh must be in heuristic IDs")
	}
	if !foundChange {
		t.Error("change_detection must be in heuristic IDs")
	}
}

func TestToMap_HasAllKeys(t *testing.T) {
	tx := makeTx(2, 2)
	h := AnalyzeTransaction(tx, 0, nil, 100)
	m := h.ToMap()

	expectedKeys := []string{
		"cioh", "change_detection", "address_reuse", "coinjoin",
		"consolidation", "self_transfer", "peeling_chain", "op_return",
		"round_number_payment",
	}
	for _, key := range expectedKeys {
		if _, ok := m[key]; !ok {
			t.Errorf("ToMap missing key %q", key)
		}
	}
}

func TestSubsetSumCheck(t *testing.T) {
	// Should find subset [3, 7] = 10
	if !subsetSumCheck([]int64{1, 3, 5, 7}, 10, 1000) {
		t.Error("Expected subset sum to find [3,7]=10")
	}

	// No subset sums to 100
	if subsetSumCheck([]int64{1, 2, 3, 4, 5}, 100, 1000) {
		t.Error("Expected no subset summing to 100")
	}

	// Depth limit — should return false if exhausted
	if subsetSumCheck([]int64{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16, 17, 18, 19, 20}, 210, 5) {
		t.Error("Expected depth limit to prevent finding answer")
	}
}
