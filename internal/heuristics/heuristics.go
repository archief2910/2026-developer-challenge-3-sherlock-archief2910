package heuristics

import (
	"math"

	"sherlock/internal/block"
	"sherlock/internal/script"
)

// HeuristicResult is the result of a single heuristic for a transaction
type HeuristicResult struct {
	Detected bool                   `json:"detected"`
	Extra    map[string]interface{} `json:"-"` // extra data for JSON serialization
}

// TxHeuristics holds heuristic results for a single transaction
type TxHeuristics struct {
	CIOH            HeuristicResult
	ChangeDetection ChangeDetectionResult
	AddressReuse    HeuristicResult
	CoinJoin        HeuristicResult
	Consolidation   HeuristicResult
	SelfTransfer    HeuristicResult
	PeelingChain    HeuristicResult
	OpReturn        OpReturnHeuristicResult
	RoundNumber     HeuristicResult
}

// ChangeDetectionResult extends HeuristicResult with change-specific fields
type ChangeDetectionResult struct {
	Detected        bool   `json:"detected"`
	LikelyChangeIdx int    `json:"likely_change_index"`
	Method          string `json:"method"`
	Confidence      string `json:"confidence"`
}

// OpReturnHeuristicResult extends HeuristicResult with OP_RETURN-specific fields
type OpReturnHeuristicResult struct {
	Detected bool   `json:"detected"`
	Protocol string `json:"protocol,omitempty"`
	Count    int    `json:"count,omitempty"`
}

// AnalyzeTransaction applies all 9 heuristics to a single parsed transaction
func AnalyzeTransaction(tx *block.ParsedTransaction) TxHeuristics {
	result := TxHeuristics{}

	if tx.IsCoinbase {
		// Coinbase transactions: most heuristics don't apply
		result.CIOH = HeuristicResult{Detected: false}
		result.ChangeDetection = ChangeDetectionResult{Detected: false, LikelyChangeIdx: -1}
		result.AddressReuse = HeuristicResult{Detected: false}
		result.CoinJoin = HeuristicResult{Detected: false}
		result.Consolidation = HeuristicResult{Detected: false}
		result.SelfTransfer = HeuristicResult{Detected: false}
		result.PeelingChain = HeuristicResult{Detected: false}
		result.RoundNumber = HeuristicResult{Detected: false}

		// OP_RETURN can still be in coinbase outputs
		result.OpReturn = analyzeOpReturn(tx)
		return result
	}

	result.CIOH = analyzeCIOH(tx)
	result.ChangeDetection = analyzeChangeDetection(tx)
	result.AddressReuse = analyzeAddressReuse(tx)
	result.CoinJoin = analyzeCoinJoin(tx)
	result.Consolidation = analyzeConsolidation(tx)
	result.SelfTransfer = analyzeSelfTransfer(tx)
	result.PeelingChain = analyzePeelingChain(tx)
	result.OpReturn = analyzeOpReturn(tx)
	result.RoundNumber = analyzeRoundNumber(tx)

	return result
}

// ClassifyTransaction determines the transaction classification
func ClassifyTransaction(tx *block.ParsedTransaction, h TxHeuristics) string {
	if tx.IsCoinbase {
		return "unknown" // Coinbase is not classified as payment patterns
	}

	// CoinJoin takes priority
	if h.CoinJoin.Detected {
		return "coinjoin"
	}

	// Consolidation: many inputs, 1-2 outputs
	if h.Consolidation.Detected {
		return "consolidation"
	}

	// Self-transfer
	if h.SelfTransfer.Detected {
		return "self_transfer"
	}

	// Batch payment: 1 or few inputs, many outputs (3+)
	numInputs := len(tx.Raw.Inputs)
	numOutputs := len(tx.Raw.Outputs)
	if numOutputs >= 3 && numInputs <= 2 {
		return "batch_payment"
	}

	// Simple payment: typical 1-in, 2-out or 1-in, 1-out
	if numInputs >= 1 && numOutputs <= 2 {
		return "simple_payment"
	}

	return "unknown"
}

// 1. CIOH: Common Input Ownership Heuristic
// All inputs to a transaction likely belong to the same entity when there are multiple inputs
func analyzeCIOH(tx *block.ParsedTransaction) HeuristicResult {
	detected := len(tx.Raw.Inputs) > 1
	return HeuristicResult{Detected: detected}
}

// 2. Change Detection
// Identify the likely change output using multiple methods
func analyzeChangeDetection(tx *block.ParsedTransaction) ChangeDetectionResult {
	numOutputs := len(tx.Raw.Outputs)
	if numOutputs < 2 {
		return ChangeDetectionResult{Detected: false, LikelyChangeIdx: -1}
	}

	// Method 1: Script type matching — change output matches predominant input type
	inputTypeCount := make(map[string]int)
	for _, st := range tx.InputScriptTypes {
		if st != "" && st != "unknown" {
			inputTypeCount[st]++
		}
	}

	// Find predominant input script type (deterministic: tie-break by name)
	predominantType := ""
	maxCount := 0
	for st, count := range inputTypeCount {
		if count > maxCount || (count == maxCount && st < predominantType) {
			maxCount = count
			predominantType = st
		}
	}

	if predominantType != "" {
		// Find outputs matching the input type — the change is typically the one matching
		matchingIdxs := []int{}
		nonMatchingIdxs := []int{}
		for j, outType := range tx.OutputScriptTypes {
			if outType == predominantType {
				matchingIdxs = append(matchingIdxs, j)
			} else {
				nonMatchingIdxs = append(nonMatchingIdxs, j)
			}
		}

		// If exactly one output matches and others don't, that's likely change
		if len(matchingIdxs) == 1 && len(nonMatchingIdxs) >= 1 {
			return ChangeDetectionResult{
				Detected:        true,
				LikelyChangeIdx: matchingIdxs[0],
				Method:          "script_type_match",
				Confidence:      "high",
			}
		}
	}

	// Method 2: Round number analysis
	// Payment amounts tend to be round numbers; non-round is likely change
	if numOutputs == 2 {
		round0 := isRoundAmount(tx.OutputValues[0])
		round1 := isRoundAmount(tx.OutputValues[1])

		if round0 && !round1 {
			return ChangeDetectionResult{
				Detected:        true,
				LikelyChangeIdx: 1,
				Method:          "round_number",
				Confidence:      "medium",
			}
		}
		if !round0 && round1 {
			return ChangeDetectionResult{
				Detected:        true,
				LikelyChangeIdx: 0,
				Method:          "round_number",
				Confidence:      "medium",
			}
		}
	}

	// Method 3: For 2-output txs where both have same type,
	// the smaller output is more likely the change
	if numOutputs == 2 {
		if tx.OutputValues[0] < tx.OutputValues[1] {
			return ChangeDetectionResult{
				Detected:        true,
				LikelyChangeIdx: 0,
				Method:          "value_analysis",
				Confidence:      "low",
			}
		}
		return ChangeDetectionResult{
			Detected:        true,
			LikelyChangeIdx: 1,
			Method:          "value_analysis",
			Confidence:      "low",
		}
	}

	return ChangeDetectionResult{Detected: false, LikelyChangeIdx: -1}
}

// 3. Address Reuse
// Detect when the same address appears in both inputs and outputs
func analyzeAddressReuse(tx *block.ParsedTransaction) HeuristicResult {
	inputAddrs := make(map[string]bool)
	for _, addr := range tx.InputAddresses {
		if addr != nil && *addr != "" {
			inputAddrs[*addr] = true
		}
	}

	for _, addr := range tx.OutputAddresses {
		if addr != nil && *addr != "" {
			if inputAddrs[*addr] {
				return HeuristicResult{Detected: true}
			}
		}
	}

	return HeuristicResult{Detected: false}
}

// 4. CoinJoin Detection
// Multiple inputs from different owners + equal-value outputs
func analyzeCoinJoin(tx *block.ParsedTransaction) HeuristicResult {
	numInputs := len(tx.Raw.Inputs)
	numOutputs := len(tx.Raw.Outputs)

	// CoinJoin typically has many inputs and outputs
	if numInputs < 3 || numOutputs < 3 {
		return HeuristicResult{Detected: false}
	}

	// Count equal-value output groups
	valueCounts := make(map[int64]int)
	for _, val := range tx.OutputValues {
		if val > 0 {
			valueCounts[val]++
		}
	}

	// Find the largest group of equal-value outputs
	maxEqualCount := 0
	for _, count := range valueCounts {
		if count > maxEqualCount {
			maxEqualCount = count
		}
	}

	// If there are 3+ outputs with the same value, it's likely a CoinJoin
	if maxEqualCount >= 3 {
		// Check if multiple input types suggest different owners
		inputTypes := make(map[string]bool)
		for _, st := range tx.InputScriptTypes {
			if st != "" && st != "unknown" {
				inputTypes[st] = true
			}
		}

		return HeuristicResult{Detected: true}
	}

	return HeuristicResult{Detected: false}
}

// 5. Consolidation Detection
// Many inputs combined into 1-2 outputs, typically same script type
func analyzeConsolidation(tx *block.ParsedTransaction) HeuristicResult {
	numInputs := len(tx.Raw.Inputs)
	numOutputs := len(tx.Raw.Outputs)

	// Consolidation: many inputs (5+), few outputs (1-2)
	if numInputs < 5 || numOutputs > 2 {
		return HeuristicResult{Detected: false}
	}

	// Check if outputs share the same script type as inputs
	if numOutputs >= 1 {
		// Check if all outputs match the predominant input type
		inputTypeCount := make(map[string]int)
		for _, st := range tx.InputScriptTypes {
			if st != "" && st != "unknown" {
				inputTypeCount[st]++
			}
		}

		predominantType := ""
		maxCount := 0
		for st, count := range inputTypeCount {
			if count > maxCount || (count == maxCount && st < predominantType) {
				maxCount = count
				predominantType = st
			}
		}

		if predominantType != "" {
			allMatch := true
			for _, outType := range tx.OutputScriptTypes {
				if outType != predominantType && outType != "op_return" {
					allMatch = false
					break
				}
			}
			if allMatch {
				return HeuristicResult{Detected: true}
			}
		}

		// Even without type matching, many-to-few is consolidation
		return HeuristicResult{Detected: true}
	}

	return HeuristicResult{Detected: false}
}

// 6. Self-Transfer Detection
// All outputs match input script type, no obvious "payment"
func analyzeSelfTransfer(tx *block.ParsedTransaction) HeuristicResult {
	numInputs := len(tx.Raw.Inputs)
	numOutputs := len(tx.Raw.Outputs)

	if numOutputs == 0 || numInputs == 0 {
		return HeuristicResult{Detected: false}
	}

	// Find predominant input type
	inputTypeCount := make(map[string]int)
	for _, st := range tx.InputScriptTypes {
		if st != "" && st != "unknown" {
			inputTypeCount[st]++
		}
	}

	predominantType := ""
	maxCount := 0
	for st, count := range inputTypeCount {
		if count > maxCount || (count == maxCount && st < predominantType) {
			maxCount = count
			predominantType = st
		}
	}

	if predominantType == "" {
		return HeuristicResult{Detected: false}
	}

	// All outputs must match the input type (exclude OP_RETURN)
	allMatch := true
	for _, outType := range tx.OutputScriptTypes {
		if outType == "op_return" {
			continue
		}
		if outType != predominantType {
			allMatch = false
			break
		}
	}

	if !allMatch {
		return HeuristicResult{Detected: false}
	}

	// Additionally, for self-transfer, typically 1-2 outputs of same type as all inputs
	// and no round-number payments suggesting an external party
	hasRoundPayment := false
	for _, val := range tx.OutputValues {
		if isRoundAmount(val) && val > 0 {
			hasRoundPayment = true
			break
		}
	}

	if allMatch && !hasRoundPayment && numOutputs <= 2 {
		return HeuristicResult{Detected: true}
	}

	return HeuristicResult{Detected: false}
}

// 7. Peeling Chain Detection
// Large input split into one small output (payment) and one large output (change)
func analyzePeelingChain(tx *block.ParsedTransaction) HeuristicResult {
	numInputs := len(tx.Raw.Inputs)
	numOutputs := len(tx.Raw.Outputs)

	// Classic peeling chain: 1 input, 2 outputs
	if numInputs != 1 || numOutputs != 2 {
		return HeuristicResult{Detected: false}
	}

	val0 := tx.OutputValues[0]
	val1 := tx.OutputValues[1]

	if val0 <= 0 || val1 <= 0 {
		return HeuristicResult{Detected: false}
	}

	// One output should be significantly smaller than the other
	ratio := float64(val0) / float64(val1)
	if ratio > 1 {
		ratio = 1.0 / ratio
	}

	// If one output is less than 10% of the other, it's likely a peeling chain
	if ratio < 0.1 {
		return HeuristicResult{Detected: true}
	}

	return HeuristicResult{Detected: false}
}

// 8. OP_RETURN Analysis
func analyzeOpReturn(tx *block.ParsedTransaction) OpReturnHeuristicResult {
	count := 0
	protocol := ""

	for i, outType := range tx.OutputScriptTypes {
		if outType == "op_return" {
			count++
			decoded := script.DecodeOpReturn(tx.Raw.Outputs[i].ScriptPubKey)
			if decoded.Protocol != "unknown" {
				protocol = decoded.Protocol
			}
		}
	}

	if count > 0 {
		return OpReturnHeuristicResult{
			Detected: true,
			Protocol: protocol,
			Count:    count,
		}
	}

	return OpReturnHeuristicResult{Detected: false}
}

// 9. Round Number Payment Detection
func analyzeRoundNumber(tx *block.ParsedTransaction) HeuristicResult {
	for _, val := range tx.OutputValues {
		if val > 0 && isRoundAmount(val) {
			return HeuristicResult{Detected: true}
		}
	}
	return HeuristicResult{Detected: false}
}

// isRoundAmount checks if a satoshi value is a round BTC amount
func isRoundAmount(sats int64) bool {
	if sats <= 0 {
		return false
	}

	// Check common round amounts in satoshis
	roundAmounts := []int64{
		100000000, // 1 BTC
		50000000,  // 0.5 BTC
		10000000,  // 0.1 BTC
		5000000,   // 0.05 BTC
		1000000,   // 0.01 BTC
		500000,    // 0.005 BTC
		100000,    // 0.001 BTC
		50000,     // 0.0005 BTC
		10000,     // 0.0001 BTC
	}

	for _, r := range roundAmounts {
		if sats%r == 0 {
			return true
		}
	}

	// Also check if it's a multiple of common clean satoshi amounts
	btcValue := float64(sats) / 100000000.0
	rounded := math.Round(btcValue*1000) / 1000 // Round to 3 decimal places
	if math.Abs(btcValue-rounded) < 0.0000001 {
		return true
	}

	return false
}

// HasAnyDetected returns true if any heuristic was detected
func HasAnyDetected(h TxHeuristics) bool {
	return h.CIOH.Detected ||
		h.ChangeDetection.Detected ||
		h.AddressReuse.Detected ||
		h.CoinJoin.Detected ||
		h.Consolidation.Detected ||
		h.SelfTransfer.Detected ||
		h.PeelingChain.Detected ||
		h.OpReturn.Detected ||
		h.RoundNumber.Detected
}

// HeuristicIDs returns the list of all heuristic IDs we apply
func HeuristicIDs() []string {
	return []string{
		"cioh",
		"change_detection",
		"address_reuse",
		"coinjoin",
		"consolidation",
		"self_transfer",
		"peeling_chain",
		"op_return",
		"round_number_payment",
	}
}

// ToMap converts TxHeuristics to a map for JSON serialization
func (h TxHeuristics) ToMap() map[string]interface{} {
	m := make(map[string]interface{})

	m["cioh"] = map[string]interface{}{
		"detected": h.CIOH.Detected,
	}

	changeMap := map[string]interface{}{
		"detected": h.ChangeDetection.Detected,
	}
	if h.ChangeDetection.Detected {
		changeMap["likely_change_index"] = h.ChangeDetection.LikelyChangeIdx
		changeMap["method"] = h.ChangeDetection.Method
		changeMap["confidence"] = h.ChangeDetection.Confidence
	}
	m["change_detection"] = changeMap

	m["address_reuse"] = map[string]interface{}{
		"detected": h.AddressReuse.Detected,
	}

	m["coinjoin"] = map[string]interface{}{
		"detected": h.CoinJoin.Detected,
	}

	m["consolidation"] = map[string]interface{}{
		"detected": h.Consolidation.Detected,
	}

	m["self_transfer"] = map[string]interface{}{
		"detected": h.SelfTransfer.Detected,
	}

	m["peeling_chain"] = map[string]interface{}{
		"detected": h.PeelingChain.Detected,
	}

	opMap := map[string]interface{}{
		"detected": h.OpReturn.Detected,
	}
	if h.OpReturn.Detected {
		if h.OpReturn.Protocol != "" {
			opMap["protocol"] = h.OpReturn.Protocol
		}
		opMap["count"] = h.OpReturn.Count
	}
	m["op_return"] = opMap

	m["round_number_payment"] = map[string]interface{}{
		"detected": h.RoundNumber.Detected,
	}

	return m
}
