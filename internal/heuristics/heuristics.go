package heuristics

import (
	"math"

	"sherlock/internal/block"
	"sherlock/internal/script"
)

// -----------------------------------------------------------------------
// Result types
// -----------------------------------------------------------------------

// HeuristicResult is the result of a single heuristic for a transaction.
// Every result carries a Confidence field ("high", "medium", "low", or "")
// so that all 9 heuristics output a probabilistic confidence level.
type HeuristicResult struct {
	Detected   bool   `json:"detected"`
	Confidence string `json:"confidence,omitempty"`
}

// TxHeuristics holds heuristic results for a single transaction.
type TxHeuristics struct {
	CIOH            HeuristicResult
	ChangeDetection ChangeDetectionResult
	AddressReuse    AddressReuseResult
	CoinJoin        CoinJoinResult
	Consolidation   HeuristicResult
	SelfTransfer    HeuristicResult
	PeelingChain    HeuristicResult
	OpReturn        OpReturnResult
	RoundNumber     RoundNumberResult
}

// ChangeDetectionResult extends HeuristicResult with change-specific fields.
// Reference: BlockSci change_by_address_type, change_by_optimal_change,
// change_by_power_of_ten_value (Kalodner et al., USENIX Security 2020).
type ChangeDetectionResult struct {
	Detected        bool   `json:"detected"`
	LikelyChangeIdx int    `json:"likely_change_index"`
	Method          string `json:"method"`
	Confidence      string `json:"confidence"`
}

// CoinJoinResult extends HeuristicResult with CoinJoin-specific fields.
// Reference: Schnoering & Vazirgiannis, arXiv 2023 — structural fingerprints
// for JoinMarket, Wasabi, and Whirlpool.
type CoinJoinResult struct {
	Detected           bool   `json:"detected"`
	EqualOutputCount   int    `json:"equal_output_count,omitempty"`
	InputTypeDiversity int    `json:"input_type_diversity,omitempty"`
	Confidence         string `json:"confidence,omitempty"`
}

// AddressReuseResult extends HeuristicResult with reuse-specific fields.
// Reference: README requirement — "Detect when the same address appears in
// both inputs and outputs of a transaction, or across multiple transactions
// within the same block."
type AddressReuseResult struct {
	Detected        bool     `json:"detected"`
	ReusedAddresses []string `json:"reused_addresses,omitempty"`
	CrossTx         bool     `json:"cross_tx,omitempty"`
	Confidence      string   `json:"confidence,omitempty"`
}

// OpReturnResult extends HeuristicResult with OP_RETURN-specific fields.
// Reference: arXiv 2411.10325v1 — practical colored coin detection
// (Omni, Open Asset, EPOBC) in OP_RETURN outputs.
type OpReturnResult struct {
	Detected   bool   `json:"detected"`
	Protocol   string `json:"protocol,omitempty"`
	Count      int    `json:"count,omitempty"`
	Confidence string `json:"confidence,omitempty"`
}

// RoundNumberResult extends HeuristicResult with round-number-specific fields.
// Reference: BlockSci change_by_power_of_ten_value (Kalodner et al., 2020).
type RoundNumberResult struct {
	Detected            bool   `json:"detected"`
	RoundOutputCount    int    `json:"round_output_count,omitempty"`
	HighestDenomination string `json:"highest_denomination,omitempty"`
	Confidence          string `json:"confidence,omitempty"`
}

// BlockAddressMap is a pre-built map of all addresses seen across a block's
// transactions, used for cross-transaction address reuse detection.
// Key = address string, Value = list of (txIndex, "input"/"output") pairs.
type BlockAddressMap map[string][]AddrOccurrence

// AddrOccurrence records where an address was seen in a block.
type AddrOccurrence struct {
	TxIndex int
	Role    string // "input" or "output"
}

// -----------------------------------------------------------------------
// Block-level analysis (for cross-tx heuristics)
// -----------------------------------------------------------------------

// BuildBlockAddressMap scans all transactions in a block and builds a map
// of every address to its occurrences. This enables cross-transaction
// address reuse detection as required by the README.
func BuildBlockAddressMap(txs []*block.ParsedTransaction) BlockAddressMap {
	addrMap := make(BlockAddressMap)
	for txIdx, tx := range txs {
		if tx.IsCoinbase {
			// Only record outputs for coinbase
			for _, addr := range tx.OutputAddresses {
				if addr != nil && *addr != "" {
					addrMap[*addr] = append(addrMap[*addr], AddrOccurrence{TxIndex: txIdx, Role: "output"})
				}
			}
			continue
		}
		for _, addr := range tx.InputAddresses {
			if addr != nil && *addr != "" {
				addrMap[*addr] = append(addrMap[*addr], AddrOccurrence{TxIndex: txIdx, Role: "input"})
			}
		}
		for _, addr := range tx.OutputAddresses {
			if addr != nil && *addr != "" {
				addrMap[*addr] = append(addrMap[*addr], AddrOccurrence{TxIndex: txIdx, Role: "output"})
			}
		}
	}
	return addrMap
}

// -----------------------------------------------------------------------
// Per-transaction analysis
// -----------------------------------------------------------------------

// AnalyzeTransaction applies all 9 heuristics to a single parsed transaction.
// The blockAddrMap parameter enables cross-transaction address reuse detection.
// The blockHeight parameter enables nLockTime wallet fingerprinting.
// Pass nil for blockAddrMap if cross-tx detection is not needed.
func AnalyzeTransaction(tx *block.ParsedTransaction, txIndex int, blockAddrMap BlockAddressMap, blockHeight int64) TxHeuristics {
	result := TxHeuristics{}

	if tx.IsCoinbase {
		// Coinbase transactions: most heuristics don't apply.
		// Coinbase has no real inputs, so CIOH, change detection, address reuse,
		// coinjoin, consolidation, self-transfer, peeling chain, and round number
		// are all not applicable.
		result.CIOH = HeuristicResult{Detected: false}
		result.ChangeDetection = ChangeDetectionResult{Detected: false, LikelyChangeIdx: -1}
		result.AddressReuse = AddressReuseResult{Detected: false}
		result.CoinJoin = CoinJoinResult{Detected: false}
		result.Consolidation = HeuristicResult{Detected: false}
		result.SelfTransfer = HeuristicResult{Detected: false}
		result.PeelingChain = HeuristicResult{Detected: false}
		result.RoundNumber = RoundNumberResult{Detected: false}

		// OP_RETURN can still be in coinbase outputs (e.g., mining pool metadata)
		result.OpReturn = analyzeOpReturn(tx)
		return result
	}

	// Apply all 9 heuristics independently first
	result.CIOH = analyzeCIOH(tx)
	result.ChangeDetection = analyzeChangeDetection(tx, blockHeight, blockAddrMap)
	result.AddressReuse = analyzeAddressReuse(tx, txIndex, blockAddrMap)
	result.CoinJoin = analyzeCoinJoin(tx)
	result.Consolidation = analyzeConsolidation(tx)
	result.SelfTransfer = analyzeSelfTransfer(tx)
	result.PeelingChain = analyzePeelingChain(tx)
	result.OpReturn = analyzeOpReturn(tx)
	result.RoundNumber = analyzeRoundNumber(tx)

	// -----------------------------------------------------------------------
	// Cross-heuristic interactions (post-analysis)
	// These interactions model the real-world relationships between heuristics
	// documented in the literature.
	// -----------------------------------------------------------------------

	// Interaction 1: CoinJoin detection downgrades CIOH confidence.
	// Reference: Gong et al. (2022) — "The multi-input (MI) heuristic has
	// higher false positives due to CoinJoin and mixing."
	// When a transaction is identified as CoinJoin, the inputs belong to
	// DIFFERENT entities, so CIOH's assumption of common ownership is invalid.
	if result.CoinJoin.Detected && result.CIOH.Detected {
		result.CIOH.Confidence = "low"
	}

	// Interaction 2: Consolidation suppresses self-transfer.
	// Consolidation is a more specific pattern (many-to-few, same type).
	// If consolidation is detected, self-transfer is redundant because
	// consolidation already implies all outputs belong to the same entity.
	if result.Consolidation.Detected && result.SelfTransfer.Detected {
		result.SelfTransfer.Detected = false
		result.SelfTransfer.Confidence = ""
	}

	// Interaction 3: Peeling chain reinforces change detection.
	// Reference: BlockSci change_by_peeling_chain (Kalodner et al., 2020) —
	// "If tx is a peeling chain, returns the smaller output."
	// In a peeling chain, the LARGER output is always the change going back
	// to the peeler, and the smaller output is the payment.
	if result.PeelingChain.Detected && !result.ChangeDetection.Detected {
		numOutputs := len(tx.Raw.Outputs)
		if numOutputs == 2 {
			changeIdx := 0
			if tx.OutputValues[0] < tx.OutputValues[1] {
				changeIdx = 1 // larger output is the change
			}
			result.ChangeDetection = ChangeDetectionResult{
				Detected:        true,
				LikelyChangeIdx: changeIdx,
				Method:          "peeling_chain",
				Confidence:      result.PeelingChain.Confidence,
			}
		}
	}

	// Interaction 4: Subset-sum CoinJoin verification.
	// Reference: BlockSci is_definite_coinjoin — "Uses subset matching to
	// determine whether this transaction is a JoinMarket coinjoin."
	// If we detected a CoinJoin candidate, verify by checking if inputs
	// can be partitioned to match output groups. If partitionable, it's
	// likely a batch payment (downgrade); if not, confirmed CoinJoin (upgrade).
	if result.CoinJoin.Detected {
		equalVal := findLargestEqualOutputValue(tx)
		if equalVal > 0 {
			numInputs := len(tx.Raw.Inputs)
			// Only run subset-sum on reasonably sized transactions
			// to avoid combinatorial explosion
			if numInputs <= 50 {
				feePerOutput := int64(0)
				if numInputs > 0 {
					totalFee := tx.TotalInputSats - tx.TotalOutputSats
					if totalFee > 0 {
						// Estimate fee share per equal-value output
						equalCount := result.CoinJoin.EqualOutputCount
						if equalCount > 0 {
							feePerOutput = totalFee / int64(equalCount)
						}
					}
				}
				target := equalVal + feePerOutput
				canPartition := subsetSumCheck(tx.InputValues, target, 10000)
				if !canPartition {
					// No subset of inputs sums to the target — confirmed CoinJoin
					result.CoinJoin.Confidence = "high"
				} else if result.CoinJoin.Confidence == "high" {
					// A partition exists — inputs can be matched to outputs by owner
					// This makes it less likely to be a CoinJoin (could be batch payment)
					result.CoinJoin.Confidence = "medium"
				}
			}
		}
	}

	return result
}

// ClassifyTransaction determines the transaction classification.
// Classification follows a priority order based on specificity:
// coinjoin > consolidation > self_transfer > batch_payment > simple_payment.
func ClassifyTransaction(tx *block.ParsedTransaction, h TxHeuristics) string {
	if tx.IsCoinbase {
		return "unknown" // Coinbase is a reward transaction, not a payment pattern
	}

	// CoinJoin takes highest priority — it's a distinct privacy-enhancing pattern
	if h.CoinJoin.Detected {
		return "coinjoin"
	}

	// Consolidation: many inputs combined into few outputs (wallet maintenance)
	if h.Consolidation.Detected {
		return "consolidation"
	}

	// Self-transfer: all outputs match input type, no external payment signal
	if h.SelfTransfer.Detected {
		return "self_transfer"
	}

	numInputs := len(tx.Raw.Inputs)
	numOutputs := len(tx.Raw.Outputs)

	// Batch payment: few inputs, many outputs (exchange/service payout pattern)
	if numOutputs >= 3 && numInputs <= 2 {
		return "batch_payment"
	}

	// Simple payment: standard 1-in/2-out or few-in/few-out pattern
	if numInputs >= 1 && numOutputs <= 2 {
		return "simple_payment"
	}

	return "unknown"
}

// =======================================================================
// Heuristic 1: Common Input Ownership Heuristic (CIOH)
//
// All inputs to a transaction likely belong to the same entity.
// This is the foundational chain analysis assumption first mentioned in
// the Bitcoin whitepaper (Nakamoto, 2008, Section 10) and formally defined
// by Meiklejohn et al. (IMC 2013).
//
// Detection: Transaction has more than one input.
// Confidence:
//   - "high"   if inputs > 3  (strong multi-UTXO combination signal)
//   - "medium" if inputs = 2-3 (weaker — could be small CoinJoin/PayJoin)
//
// Known limitations:
//   - False positives for CoinJoin (mitigated by cross-heuristic downgrade)
//   - False positives for PayJoin (Ghesmati et al., 2021)
//   - Single-input transactions always false, even if entity owns many UTXOs
//
// =======================================================================
func analyzeCIOH(tx *block.ParsedTransaction) HeuristicResult {
	numInputs := len(tx.Raw.Inputs)
	if numInputs <= 1 {
		return HeuristicResult{Detected: false}
	}

	confidence := "medium"
	if numInputs > 3 {
		confidence = "high"
	}

	return HeuristicResult{Detected: true, Confidence: confidence}
}

// =======================================================================
// Heuristic 2: Change Detection
//
// Identify the likely change output — the output that returns leftover
// funds to the sender's wallet rather than going to the payment recipient.
//
// Six methods applied in priority order (highest confidence first):
//
// Method 1: Script type matching (high confidence)
//
//	BlockSci change_by_address_type: "If all inputs are of one address type,
//	it is likely that the change output has the same type."
//	Reference: Kalodner et al. (2020), BlockSci heuristics documentation.
//
// Method 2: Optimal change (high confidence)
//
//	BlockSci change_by_optimal_change: "If there exists an output that is
//	smaller than any of the inputs it is likely the change. If a change
//	output was larger than the smallest input, then the coin selection
//	algorithm wouldn't need to add the input in the first place."
//	Reference: Kalodner et al. (2020), BlockSci heuristics documentation.
//
// Method 3: Round number analysis (medium confidence)
//
//	Payment amounts tend to be round numbers; non-round outputs are more
//	likely change. Reference: Androulaki et al. (2012) — first proposed
//	round-value-based change detection (termed "Shadow Addresses").
//
// Method 4: nLockTime wallet fingerprinting (medium confidence)
//
//	Bitcoin Core sets nLockTime to the current block height to prevent fee
//	sniping. If tx.Locktime matches block height (±2), combined with script
//	type matching, identifies change. Reference: Möser & Narayanan (2021);
//	BlockSci change_by_locktime.
//
// Method 5: Fresh-address heuristic (low confidence)
//
//	Wallets generate fresh addresses for change. If exactly one output
//	address is unique within the block, it's likely change. Reference:
//	Meiklejohn et al. (2013); BlockSci change_by_client_change_address_behavior.
//	Note: block-level freshness is a weaker proxy for full-chain freshness.
//
// Method 6: Value analysis (low confidence)
//
//	For 2-output transactions, the smaller output is tentatively identified
//	as change. Weakest signal — purely statistical last-resort fallback.
//
// Known limitations:
//   - Address reuse rate has dropped below 10% on modern Bitcoin (Gong et al., 2025),
//     weakening fresh-address-based change heuristics.
//   - Fails when both payment and change use the same script type.
//   - Privacy-conscious wallets may add noise to output amounts.
//   - Batch payments have no change output.
//   - nLockTime fingerprinting only works for Bitcoin Core; other wallets set
//     nLockTime = 0.
//   - Fresh-address check is block-level only — full-chain freshness would
//     be more accurate but requires full UTXO set.
//
// =======================================================================
func analyzeChangeDetection(tx *block.ParsedTransaction, blockHeight int64, blockAddrMap BlockAddressMap) ChangeDetectionResult {
	numOutputs := len(tx.Raw.Outputs)
	if numOutputs < 2 {
		return ChangeDetectionResult{Detected: false, LikelyChangeIdx: -1}
	}

	// ----- Method 1: Script type matching (high confidence) -----
	// Find the predominant input script type
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
		matchingIdxs := []int{}
		nonMatchingIdxs := []int{}
		for j, outType := range tx.OutputScriptTypes {
			if outType == "op_return" {
				continue // OP_RETURN outputs are data carriers, not payments/change
			}
			if outType == predominantType {
				matchingIdxs = append(matchingIdxs, j)
			} else {
				nonMatchingIdxs = append(nonMatchingIdxs, j)
			}
		}

		// If exactly one output matches the input type and at least one doesn't,
		// the matching output is likely change (wallet sends change to same type).
		if len(matchingIdxs) == 1 && len(nonMatchingIdxs) >= 1 {
			return ChangeDetectionResult{
				Detected:        true,
				LikelyChangeIdx: matchingIdxs[0],
				Method:          "script_type_match",
				Confidence:      "high",
			}
		}
	}

	// ----- Method 2: Optimal change (high confidence) -----
	// BlockSci change_by_optimal_change: If an output is smaller than the
	// smallest input, it's likely change. Coin selection wouldn't include
	// an input if the change exceeded its value.
	// However, in peeling chain patterns (extreme asymmetry), the LARGER
	// output is the change going back to the peeler, not the smaller.
	// Reference: Kappos et al. (2022) "How to Peel a Million".
	if numOutputs == 2 && len(tx.InputValues) > 0 {
		minInputVal := tx.InputValues[0]
		for _, iv := range tx.InputValues[1:] {
			if iv > 0 && iv < minInputVal {
				minInputVal = iv
			}
		}

		if minInputVal > 0 {
			// Check if there's extreme asymmetry (peeling chain pattern)
			// In peeling chain, larger output = change
			val0 := tx.OutputValues[0]
			val1 := tx.OutputValues[1]
			if val0 > 0 && val1 > 0 {
				ratio := float64(val0) / float64(val1)
				if ratio > 1 {
					ratio = 1.0 / ratio
				}
				// Extreme asymmetry (< 1% ratio): larger output is change
				// This overrides the standard optimal_change logic
				// Only trigger on 99:1 or stronger splits (real peeling chains)
				if ratio < 0.01 {
					changeIdx := 0
					if val0 < val1 {
						changeIdx = 1
					}
					return ChangeDetectionResult{
						Detected:        true,
						LikelyChangeIdx: changeIdx,
						Method:          "optimal_change_peeling",
						Confidence:      "high",
					}
				}
			}

			// Standard optimal_change: check which output(s) are smaller than the smallest input
			out0Smaller := val0 > 0 && val0 < minInputVal
			out1Smaller := val1 > 0 && val1 < minInputVal

			if out0Smaller && !out1Smaller {
				return ChangeDetectionResult{
					Detected:        true,
					LikelyChangeIdx: 0,
					Method:          "optimal_change",
					Confidence:      "high",
				}
			}
			if out1Smaller && !out0Smaller {
				return ChangeDetectionResult{
					Detected:        true,
					LikelyChangeIdx: 1,
					Method:          "optimal_change",
					Confidence:      "high",
				}
			}
		}
	}

	// ----- Method 3: Round number analysis (medium confidence) -----
	// Human-chosen payment amounts tend to be round numbers.
	if numOutputs == 2 {
		round0 := isRoundAmount(tx.OutputValues[0])
		round1 := isRoundAmount(tx.OutputValues[1])

		// If one output is round (likely payment) and the other isn't (likely change)
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

	// ----- Method 4: nLockTime wallet fingerprinting (medium confidence) -----
	// Reference: Möser & Narayanan (2021) — wallet fingerprinting via nLockTime.
	// Reference: BlockSci change_by_locktime — "Bitcoin Core sets the locktime
	// to the current block height to prevent fee sniping."
	// If tx.Locktime matches block height (±2 for reorg tolerance), the tx
	// was likely created by Bitcoin Core. Combined with script type matching,
	// this identifies the change output.
	if numOutputs == 2 && blockHeight > 0 && tx.Raw.Locktime > 0 {
		locktime := int64(tx.Raw.Locktime)
		// nLockTime within ±2 of block height = Bitcoin Core anti-fee-sniping
		if locktime >= blockHeight-2 && locktime <= blockHeight {
			// Find the output matching the predominant input type
			if predominantType != "" {
				for j, outType := range tx.OutputScriptTypes {
					if outType == "op_return" {
						continue
					}
					if outType == predominantType {
						return ChangeDetectionResult{
							Detected:        true,
							LikelyChangeIdx: j,
							Method:          "locktime_fingerprint",
							Confidence:      "medium",
						}
					}
				}
			}
		}
	}

	// ----- Method 5: Fresh-address heuristic (low confidence) -----
	// Reference: Meiklejohn et al. (2013) — "The output must be a fresh address
	// (never before seen on-chain) and it must be the only fresh output."
	// Reference: BlockSci change_by_client_change_address_behavior — "Most
	// clients will generate a fresh address for the change."
	// Within-block freshness check: if an output address appears only once
	// across the entire block (just this output), it's likely fresh/change.
	// Note: true freshness requires full blockchain history; block-level is
	// a weaker proxy, hence low confidence.
	if numOutputs == 2 && blockAddrMap != nil {
		freshIdx := -1
		nonFreshIdx := -1
		for j, addr := range tx.OutputAddresses {
			if addr == nil || *addr == "" || tx.OutputScriptTypes[j] == "op_return" {
				continue
			}
			occurrences := blockAddrMap[*addr]
			if len(occurrences) == 1 {
				// Address appears only once in the entire block — likely fresh
				if freshIdx == -1 {
					freshIdx = j
				} else {
					freshIdx = -1 // Both are fresh, can't distinguish
					break
				}
			} else {
				nonFreshIdx = j
			}
		}
		// If exactly one output is "fresh" (appears once) and the other isn't,
		// the fresh one is likely change (wallets generate new addresses for change).
		if freshIdx >= 0 && nonFreshIdx >= 0 {
			return ChangeDetectionResult{
				Detected:        true,
				LikelyChangeIdx: freshIdx,
				Method:          "fresh_address",
				Confidence:      "low",
			}
		}
	}

	// ----- Method 6: Value analysis (low confidence) -----
	// For 2-output transactions, the smaller output is tentatively change.
	// This is the weakest signal — purely statistical fallback.
	// However, if we detect a peeling chain pattern (extreme asymmetry),
	// the LARGER output is the change (the remainder going back to peeler).
	// Reference: Kappos et al. (2022) "How to Peel a Million" —
	// in peeling chains, large output = change, small output = payment.
	if numOutputs == 2 {
		val0 := tx.OutputValues[0]
		val1 := tx.OutputValues[1]
		if val0 > 0 && val1 > 0 {
			ratio := float64(val0) / float64(val1)
			if ratio > 1 {
				ratio = 1.0 / ratio
			}
			// Peeling chain pattern: extreme asymmetry (< 1%) means larger = change
			if ratio < 0.01 {
				changeIdx := 0
				if val0 < val1 {
					changeIdx = 1
				}
				return ChangeDetectionResult{
					Detected:        true,
					LikelyChangeIdx: changeIdx,
					Method:          "value_analysis_peeling",
					Confidence:      "medium", // Upgraded from low — peeling pattern is strong signal
				}
			}
		}
		// Fallback: smaller output is change (weakest signal)
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

// =======================================================================
// Heuristic 3: Address Reuse
//
// Detect when the same address appears in both inputs and outputs of a
// transaction, OR across multiple transactions within the same block.
//
// Address reuse is a definitive privacy leak — it links different
// transactions to the same entity and makes tracking trivial.
//
// Reference: Zhang, Wang, Luo (IEEE Access 2020) — proposed detecting
// one-time change addresses by eliminating addresses reused later.
// Reference: Meiklejohn et al. (2013) — address reuse links transactions.
//
// Detection:
//
//	(a) Within-transaction: same address in inputs AND outputs
//	(b) Cross-transaction: same address appears as input in one tx and
//	    output in another tx within the same block
//
// Confidence: Always "high" — address reuse is a definitive, non-probabilistic
// signal. The same cryptographic key is being reused.
//
// Known limitations:
//   - Only detects reuse within a single block, not across blocks
//     (would require full UTXO set tracking)
//   - Some legitimate use cases: donation addresses, mining pool payouts
//   - Non-standard scripts may not produce identifiable addresses
//
// =======================================================================
func analyzeAddressReuse(tx *block.ParsedTransaction, txIndex int, blockAddrMap BlockAddressMap) AddressReuseResult {
	reusedAddrs := []string{}
	isCrossTx := false

	// (a) Within-transaction reuse: same address in inputs AND outputs
	inputAddrs := make(map[string]bool)
	for _, addr := range tx.InputAddresses {
		if addr != nil && *addr != "" {
			inputAddrs[*addr] = true
		}
	}

	for _, addr := range tx.OutputAddresses {
		if addr != nil && *addr != "" {
			if inputAddrs[*addr] {
				reusedAddrs = append(reusedAddrs, *addr)
			}
		}
	}

	// (b) Cross-transaction reuse within block:
	// Check if any of this transaction's addresses appear in OTHER transactions.
	if blockAddrMap != nil {
		allMyAddrs := make(map[string]bool)
		for _, addr := range tx.InputAddresses {
			if addr != nil && *addr != "" {
				allMyAddrs[*addr] = true
			}
		}
		for _, addr := range tx.OutputAddresses {
			if addr != nil && *addr != "" {
				allMyAddrs[*addr] = true
			}
		}

		for addr := range allMyAddrs {
			occurrences := blockAddrMap[addr]
			hasOtherTx := false
			for _, occ := range occurrences {
				if occ.TxIndex != txIndex {
					hasOtherTx = true
					break
				}
			}
			if hasOtherTx {
				// This address appears in another transaction within the same block
				alreadyTracked := false
				for _, ra := range reusedAddrs {
					if ra == addr {
						alreadyTracked = true
						break
					}
				}
				if !alreadyTracked {
					reusedAddrs = append(reusedAddrs, addr)
				}
				isCrossTx = true
			}
		}
	}

	if len(reusedAddrs) > 0 {
		return AddressReuseResult{
			Detected:        true,
			ReusedAddresses: reusedAddrs,
			CrossTx:         isCrossTx,
			Confidence:      "high",
		}
	}

	return AddressReuseResult{Detected: false}
}

// =======================================================================
// Heuristic 4: CoinJoin Detection
//
// Identify CoinJoin transactions — privacy-enhancing constructions where
// multiple users combine their inputs and create equal-value outputs to
// obscure the transaction graph.
//
// Reference: Schnoering & Vazirgiannis (arXiv 2023) — structural
// fingerprints for JoinMarket, Wasabi 1.0/1.1, and Whirlpool. Analysis
// covers transactions up to block 760,000.
// Reference: BlockSci is_coinjoin — "Uses basic structural features to
// quickly decide whether this transaction might be a JoinMarket coinjoin."
//
// Detection criteria:
//   - Transaction has 3+ inputs AND 3+ outputs
//   - There exists a group of 3+ outputs with identical values
//   - Input type diversity is tracked (multiple script types suggest
//     inputs from different wallets/users)
//
// Confidence:
//   - "high"   if equal-value outputs >= 5 AND >= 2 distinct input types
//   - "medium" if equal-value outputs >= 3
//
// Known limitations:
//   - Does not detect PayJoin / Stowaway (no equal-value outputs)
//   - May flag non-CoinJoin transactions with coincidentally equal outputs
//     (e.g., batch payments to the same value)
//   - Does not implement BlockSci's subset-sum matching (is_definite_coinjoin)
//     which would eliminate false positives but is computationally expensive
//
// =======================================================================
func analyzeCoinJoin(tx *block.ParsedTransaction) CoinJoinResult {
	numInputs := len(tx.Raw.Inputs)
	numOutputs := len(tx.Raw.Outputs)

	// CoinJoin typically has many inputs and outputs
	if numInputs < 3 || numOutputs < 3 {
		return CoinJoinResult{Detected: false}
	}

	// Count equal-value output groups (exclude OP_RETURN and zero-value outputs)
	valueCounts := make(map[int64]int)
	for i, val := range tx.OutputValues {
		if val > 0 && tx.OutputScriptTypes[i] != "op_return" {
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

	// Require at least 3 equal-value outputs for CoinJoin detection
	if maxEqualCount < 3 {
		return CoinJoinResult{Detected: false}
	}

	// Count distinct input script types (indicator of different wallets)
	inputTypes := make(map[string]bool)
	for _, st := range tx.InputScriptTypes {
		if st != "" && st != "unknown" {
			inputTypes[st] = true
		}
	}
	inputTypeDiversity := len(inputTypes)

	// Assign confidence based on signal strength
	confidence := "medium"
	if maxEqualCount >= 5 && inputTypeDiversity >= 2 {
		// Strong CoinJoin signal: many equal outputs + mixed input types
		// suggest multiple distinct wallets contributing inputs
		confidence = "high"
	}

	return CoinJoinResult{
		Detected:           true,
		EqualOutputCount:   maxEqualCount,
		InputTypeDiversity: inputTypeDiversity,
		Confidence:         confidence,
	}
}

// =======================================================================
// Heuristic 5: Consolidation Detection
//
// Detect consolidation transactions where many UTXOs are combined into
// one or two outputs. This is common wallet maintenance to reduce UTXO
// set size and lower future transaction fees.
//
// Reference: Standard transaction taxonomy from blockchain analysis
// (Gong et al., IFIP Digital Forensics 2022).
//
// Detection criteria:
//   - Transaction has 5+ inputs AND at most 2 non-OP_RETURN outputs
//
// Confidence:
//   - "high"   if inputs >= 10 AND all outputs match predominant input
//     script type (strong consolidation signal — same wallet)
//   - "medium" if inputs 5-9 or output types don't fully match input types
//
// Known limitations:
//   - 5-input threshold is a conservative choice; smaller consolidations exist
//   - Large payments with many UTXOs may be misclassified
//   - Exchange withdrawal batches can look similar
//
// =======================================================================
func analyzeConsolidation(tx *block.ParsedTransaction) HeuristicResult {
	numInputs := len(tx.Raw.Inputs)

	// Count non-OP_RETURN outputs
	nonOpReturnOutputs := 0
	for _, outType := range tx.OutputScriptTypes {
		if outType != "op_return" {
			nonOpReturnOutputs++
		}
	}

	// Consolidation: many inputs (5+), few outputs (1-2 non-OP_RETURN)
	if numInputs < 5 || nonOpReturnOutputs > 2 {
		return HeuristicResult{Detected: false}
	}

	// Find predominant input script type
	inputTypeCount := make(map[string]int)
	for _, st := range tx.InputScriptTypes {
		if st != "" && st != "unknown" {
			inputTypeCount[st]++
		}
	}

	predominantType := ""
	maxTypeCount := 0
	for st, count := range inputTypeCount {
		if count > maxTypeCount || (count == maxTypeCount && st < predominantType) {
			maxTypeCount = count
			predominantType = st
		}
	}

	// Check if all non-OP_RETURN outputs match the predominant input type
	allOutputsMatch := true
	if predominantType != "" {
		for _, outType := range tx.OutputScriptTypes {
			if outType == "op_return" {
				continue
			}
			if outType != predominantType {
				allOutputsMatch = false
				break
			}
		}
	} else {
		allOutputsMatch = false
	}

	// Assign confidence
	confidence := "medium"
	if numInputs >= 10 && allOutputsMatch {
		confidence = "high"
	}

	return HeuristicResult{Detected: true, Confidence: confidence}
}

// =======================================================================
// Heuristic 6: Self-Transfer Detection
//
// Identify transactions where all inputs and outputs appear to belong to
// the same entity — funds moving within one's own wallet with no
// external payment component.
//
// Detection criteria:
//   - All non-OP_RETURN outputs match the predominant input script type
//   - No outputs have round BTC amounts (suggesting no external payment)
//   - At most 3 outputs (typical wallet internal transfer)
//   - NOT already classified as consolidation (consolidation is more specific)
//
// Confidence: Always "medium" — script type uniformity is suggestive but
// NOT definitive. A user could pay someone who uses the same address type.
//
// Known limitations:
//   - False positives when paying to the same address type
//   - HD wallets (BIP32) generate consistent address types, making same-type
//     outputs common even for external payments
//   - Round number check may miss self-transfers with coincidentally round amounts
//
// =======================================================================
func analyzeSelfTransfer(tx *block.ParsedTransaction) HeuristicResult {
	numInputs := len(tx.Raw.Inputs)
	numOutputs := len(tx.Raw.Outputs)

	if numOutputs == 0 || numInputs == 0 {
		return HeuristicResult{Detected: false}
	}

	// Count non-OP_RETURN outputs
	nonOpReturnOutputs := 0
	for _, outType := range tx.OutputScriptTypes {
		if outType != "op_return" {
			nonOpReturnOutputs++
		}
	}

	// Self-transfers are typically small (1-3 outputs)
	if nonOpReturnOutputs > 3 {
		return HeuristicResult{Detected: false}
	}

	// Find predominant input script type
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

	// All non-OP_RETURN outputs must match the predominant input type
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

	// Check for round-number outputs — round amounts suggest an external payment
	// (humans choose round amounts). If any output is round, this is less likely
	// to be a self-transfer.
	hasRoundPayment := false
	for _, val := range tx.OutputValues {
		if val > 0 && isSignificantRoundAmount(val) {
			hasRoundPayment = true
			break
		}
	}

	if hasRoundPayment {
		return HeuristicResult{Detected: false}
	}

	return HeuristicResult{Detected: true, Confidence: "medium"}
}

// =======================================================================
// Heuristic 7: Peeling Chain Detection
//
// Detect peeling chain patterns where a large UTXO is progressively
// "peeled" — a small payment is made and the large remainder is sent to
// a new change address, which is then peeled again.
//
// Reference: Kappos et al. (USENIX Security 2022) — "How to Peel a
// Million: Validating and Expanding Bitcoin Clusters". The definitive
// peeling chain paper with the findNext/findPrev algorithm.
// Reference: BlockSci change_by_peeling_chain — "If tx is a peeling
// chain, returns the smaller output."
//
// Detection criteria:
//   - Exactly 1 input and 2 outputs
//   - Extreme value asymmetry between the two outputs
//
// Confidence (based on the ratio of smaller/larger output):
//   - "high"   if ratio < 0.01 (99:1 split — very strong peeling signal)
//   - "medium" if ratio 0.01-0.05 (strong but could be normal payment)
//   - "low"    if ratio 0.05-0.1 (weaker — many simple payments look like this)
//
// Known limitations:
//   - Cannot detect the "chain" aspect without cross-transaction analysis.
//     We detect individual peeling-pattern transactions, not the full chain.
//   - Many simple payments also have 1-input, 2-output with asymmetric values.
//   - The threshold ratios are engineering choices; optimal values may vary.
//
// =======================================================================
func analyzePeelingChain(tx *block.ParsedTransaction) HeuristicResult {
	numInputs := len(tx.Raw.Inputs)
	numOutputs := len(tx.Raw.Outputs)

	// Classic peeling chain: exactly 1 input, 2 outputs
	if numInputs != 1 || numOutputs != 2 {
		return HeuristicResult{Detected: false}
	}

	val0 := tx.OutputValues[0]
	val1 := tx.OutputValues[1]

	if val0 <= 0 || val1 <= 0 {
		return HeuristicResult{Detected: false}
	}

	// Compute the ratio of the smaller output to the larger output
	ratio := float64(val0) / float64(val1)
	if ratio > 1 {
		ratio = 1.0 / ratio
	}

	// Assign confidence based on value asymmetry
	if ratio < 0.01 {
		// Extreme asymmetry (99:1 or greater) — very strong peeling signal
		return HeuristicResult{Detected: true, Confidence: "high"}
	}
	if ratio < 0.05 {
		// Strong asymmetry (20:1 to 99:1) — likely peeling
		return HeuristicResult{Detected: true, Confidence: "medium"}
	}
	if ratio < 0.1 {
		// Moderate asymmetry (10:1 to 20:1) — possible peeling
		return HeuristicResult{Detected: true, Confidence: "low"}
	}

	return HeuristicResult{Detected: false}
}

// =======================================================================
// Heuristic 8: OP_RETURN Analysis
//
// Detect OP_RETURN outputs and classify the embedded data by protocol.
// OP_RETURN (opcode 0x6a) outputs are unspendable data carriers used by
// various protocols built on top of Bitcoin.
//
// Reference: arXiv 2411.10325v1 — practical colored coin detection
// (Omni, Open Asset, EPOBC) and why OP_RETURN protocol transactions
// should be excluded from standard payment heuristics.
//
// Protocol detection by verified payload prefix (unencrypted, documented):
//   - "6f6d6e69" → Omni Layer (hex encoding of "omni")
//   - "4f41"     → Open Assets / EPOBC (hex encoding of "OA" protocol tag)
//
// Protocols intentionally NOT matched (see opreturn.go for full rationale):
//   - Counterparty: ARC4-encrypted data — raw prefix won't appear
//   - OpenTimestamps: no fixed prefix — embeds raw 32-byte hash
//   - Veriblock: identified by 80-byte structure, not fixed prefix
//
// Confidence: Always "high" — OP_RETURN is identified by opcode 0x6a,
// which is a definitive, non-probabilistic detection. Protocol classification
// is limited to verified prefixes to avoid false positives.
//
// Known limitations:
//   - Only recognizes 2 protocol families with verified prefixes; many others
//     use OP_RETURN but require protocol-specific decoding (ARC4, structure parsing)
//   - Does not decode protocol-specific payload contents
//   - Some OP_RETURN data is arbitrary text, not protocol data
//
// =======================================================================
func analyzeOpReturn(tx *block.ParsedTransaction) OpReturnResult {
	count := 0
	protocol := ""

	for i, outType := range tx.OutputScriptTypes {
		if outType == "op_return" {
			count++
			decoded := script.DecodeOpReturn(tx.Raw.Outputs[i].ScriptPubKey)
			if decoded.Protocol != "unknown" && protocol == "" {
				protocol = decoded.Protocol
			}
		}
	}

	if count > 0 {
		return OpReturnResult{
			Detected:   true,
			Protocol:   protocol,
			Count:      count,
			Confidence: "high",
		}
	}

	return OpReturnResult{Detected: false}
}

// =======================================================================
// Heuristic 9: Round Number Payment Detection
//
// Identify outputs with values that are round BTC amounts.
// Round-number outputs are more likely to be intentional payments by humans;
// non-round outputs are more likely to be change.
//
// Reference: BlockSci change_by_power_of_ten_value (Kalodner et al., 2020) —
// "Detects possible change outputs by checking for output values that are
// multiples of 10^digits."
// Reference: Androulaki et al. (2012) — first proposed that if a value is
// close to a round number, that output is likely a payment.
//
// Confidence (based on the denomination of the round amount):
//   - "high"   if output is divisible by 0.1 BTC (10,000,000 sats) or larger
//   - "medium" if divisible by 0.001 BTC (100,000 sats) to 0.01 BTC
//   - "low"    if divisible by smaller round amounts (0.0001 BTC = 10,000 sats)
//
// Known limitations:
//   - False positives when change happens to be a round number
//   - Some payments use non-round amounts (e.g., fiat-equivalent invoices)
//   - The definition of "round" is somewhat subjective
//
// =======================================================================
func analyzeRoundNumber(tx *block.ParsedTransaction) RoundNumberResult {
	roundCount := 0
	bestDenomination := ""
	bestDenomLevel := 0 // higher = more significant round amount

	for _, val := range tx.OutputValues {
		if val <= 0 {
			continue
		}

		level, denom := classifyRoundAmount(val)
		if level > 0 {
			roundCount++
			if level > bestDenomLevel {
				bestDenomLevel = level
				bestDenomination = denom
			}
		}
	}

	if roundCount > 0 {
		// Assign confidence based on the most significant round denomination found
		confidence := "low"
		if bestDenomLevel >= 3 {
			confidence = "high" // 0.1 BTC or larger
		} else if bestDenomLevel >= 2 {
			confidence = "medium" // 0.001 to 0.01 BTC
		}

		return RoundNumberResult{
			Detected:            true,
			RoundOutputCount:    roundCount,
			HighestDenomination: bestDenomination,
			Confidence:          confidence,
		}
	}

	return RoundNumberResult{Detected: false}
}

// classifyRoundAmount returns a level (0=not round, 1-4=increasingly round)
// and a human-readable denomination string.
func classifyRoundAmount(sats int64) (int, string) {
	if sats <= 0 {
		return 0, ""
	}

	// Level 4: Whole BTC amounts (100,000,000 sats = 1 BTC)
	if sats%100000000 == 0 {
		return 4, "1_btc"
	}
	// Level 3: 0.05 BTC to 0.5 BTC denominations
	if sats%50000000 == 0 {
		return 3, "0.5_btc"
	}
	if sats%10000000 == 0 {
		return 3, "0.1_btc"
	}
	if sats%5000000 == 0 {
		return 3, "0.05_btc"
	}
	// Level 2: 0.001 to 0.01 BTC
	if sats%1000000 == 0 {
		return 2, "0.01_btc"
	}
	if sats%500000 == 0 {
		return 2, "0.005_btc"
	}
	if sats%100000 == 0 {
		return 2, "0.001_btc"
	}
	// Level 1: Smaller round amounts
	if sats%50000 == 0 {
		return 1, "0.0005_btc"
	}
	if sats%10000 == 0 {
		return 1, "0.0001_btc"
	}

	return 0, ""
}

// isRoundAmount checks if a satoshi value is a round BTC amount.
// Used by change_detection's round number method.
func isRoundAmount(sats int64) bool {
	level, _ := classifyRoundAmount(sats)
	return level > 0
}

// isSignificantRoundAmount checks if a satoshi value is a round BTC amount
// at a significance level of at least 0.001 BTC (level >= 2).
// Used by self_transfer to detect likely external payments — a payment
// to an external party is more likely to be a significant round amount.
func isSignificantRoundAmount(sats int64) bool {
	level, _ := classifyRoundAmount(sats)
	return level >= 2
}

// =======================================================================
// CoinJoin subset-sum helpers
//
// Reference: BlockSci is_definite_coinjoin — "Uses subset matching to
// determine whether this transaction is a JoinMarket coinjoin. If
// maxDepth != 0, it limits the total number of possible subsets the
// algorithm will check."
// =======================================================================

// findLargestEqualOutputValue returns the value of the largest group of
// equal-value outputs, or 0 if no group has 3+ members.
func findLargestEqualOutputValue(tx *block.ParsedTransaction) int64 {
	valueCounts := make(map[int64]int)
	for i, val := range tx.OutputValues {
		if val > 0 && tx.OutputScriptTypes[i] != "op_return" {
			valueCounts[val]++
		}
	}

	bestVal := int64(0)
	bestCount := 0
	for val, count := range valueCounts {
		if count > bestCount || (count == bestCount && val > bestVal) {
			bestCount = count
			bestVal = val
		}
	}

	if bestCount >= 3 {
		return bestVal
	}
	return 0
}

// subsetSumCheck performs a depth-limited search to determine whether any
// subset of the given values sums to the target amount. Returns true if
// such a subset exists (i.e., the inputs CAN be partitioned by owner,
// suggesting it's NOT a confirmed CoinJoin).
//
// The maxIterations parameter limits the search space to avoid exponential
// blowup on large transactions, following BlockSci's maxDepth approach.
func subsetSumCheck(values []int64, target int64, maxIterations int) bool {
	if target <= 0 || len(values) == 0 {
		return false
	}

	iterations := 0
	var search func(idx int, remaining int64) bool
	search = func(idx int, remaining int64) bool {
		if remaining == 0 {
			return true // Found a subset that sums to target
		}
		if remaining < 0 || idx >= len(values) {
			return false
		}
		iterations++
		if iterations >= maxIterations {
			return false // Depth limit reached — treat as "no partition found"
		}

		// Include values[idx]
		if values[idx] > 0 && values[idx] <= remaining {
			if search(idx+1, remaining-values[idx]) {
				return true
			}
		}
		// Exclude values[idx]
		return search(idx+1, remaining)
	}

	return search(0, target)
}

// =======================================================================
// Utility functions
// =======================================================================

// HasAnyDetected returns true if any heuristic was detected for a transaction.
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

// HeuristicIDs returns the list of all heuristic IDs we apply.
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

// ToMap converts TxHeuristics to a map for JSON serialization.
// Every heuristic outputs a "detected" boolean and a "confidence" string
// when detected, ensuring the JSON schema is consistent across all 9.
func (h TxHeuristics) ToMap() map[string]interface{} {
	m := make(map[string]interface{})

	// 1. CIOH
	ciohMap := map[string]interface{}{
		"detected": h.CIOH.Detected,
	}
	if h.CIOH.Detected && h.CIOH.Confidence != "" {
		ciohMap["confidence"] = h.CIOH.Confidence
	}
	m["cioh"] = ciohMap

	// 2. Change Detection
	changeMap := map[string]interface{}{
		"detected": h.ChangeDetection.Detected,
	}
	if h.ChangeDetection.Detected {
		changeMap["likely_change_index"] = h.ChangeDetection.LikelyChangeIdx
		changeMap["method"] = h.ChangeDetection.Method
		changeMap["confidence"] = h.ChangeDetection.Confidence
	}
	m["change_detection"] = changeMap

	// 3. Address Reuse
	addrMap := map[string]interface{}{
		"detected": h.AddressReuse.Detected,
	}
	if h.AddressReuse.Detected {
		if h.AddressReuse.Confidence != "" {
			addrMap["confidence"] = h.AddressReuse.Confidence
		}
		if h.AddressReuse.CrossTx {
			addrMap["cross_tx"] = true
		}
	}
	m["address_reuse"] = addrMap

	// 4. CoinJoin
	cjMap := map[string]interface{}{
		"detected": h.CoinJoin.Detected,
	}
	if h.CoinJoin.Detected {
		if h.CoinJoin.Confidence != "" {
			cjMap["confidence"] = h.CoinJoin.Confidence
		}
		if h.CoinJoin.EqualOutputCount > 0 {
			cjMap["equal_output_count"] = h.CoinJoin.EqualOutputCount
		}
		if h.CoinJoin.InputTypeDiversity > 0 {
			cjMap["input_type_diversity"] = h.CoinJoin.InputTypeDiversity
		}
	}
	m["coinjoin"] = cjMap

	// 5. Consolidation
	consMap := map[string]interface{}{
		"detected": h.Consolidation.Detected,
	}
	if h.Consolidation.Detected && h.Consolidation.Confidence != "" {
		consMap["confidence"] = h.Consolidation.Confidence
	}
	m["consolidation"] = consMap

	// 6. Self-Transfer
	stMap := map[string]interface{}{
		"detected": h.SelfTransfer.Detected,
	}
	if h.SelfTransfer.Detected && h.SelfTransfer.Confidence != "" {
		stMap["confidence"] = h.SelfTransfer.Confidence
	}
	m["self_transfer"] = stMap

	// 7. Peeling Chain
	pcMap := map[string]interface{}{
		"detected": h.PeelingChain.Detected,
	}
	if h.PeelingChain.Detected && h.PeelingChain.Confidence != "" {
		pcMap["confidence"] = h.PeelingChain.Confidence
	}
	m["peeling_chain"] = pcMap

	// 8. OP_RETURN
	opMap := map[string]interface{}{
		"detected": h.OpReturn.Detected,
	}
	if h.OpReturn.Detected {
		if h.OpReturn.Confidence != "" {
			opMap["confidence"] = h.OpReturn.Confidence
		}
		if h.OpReturn.Protocol != "" {
			opMap["protocol"] = h.OpReturn.Protocol
		}
		opMap["count"] = h.OpReturn.Count
	}
	m["op_return"] = opMap

	// 9. Round Number Payment
	rnMap := map[string]interface{}{
		"detected": h.RoundNumber.Detected,
	}
	if h.RoundNumber.Detected {
		if h.RoundNumber.Confidence != "" {
			rnMap["confidence"] = h.RoundNumber.Confidence
		}
		if h.RoundNumber.RoundOutputCount > 0 {
			rnMap["round_output_count"] = h.RoundNumber.RoundOutputCount
		}
		if h.RoundNumber.HighestDenomination != "" {
			rnMap["highest_denomination"] = h.RoundNumber.HighestDenomination
		}
	}
	m["round_number_payment"] = rnMap

	return m
}

// Ensure math import is used
var _ = math.Abs
