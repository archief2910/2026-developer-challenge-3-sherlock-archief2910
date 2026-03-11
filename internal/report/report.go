package report

import (
	"fmt"
	"sort"
	"strings"

	"sherlock/internal/analysis"
)

// GenerateMarkdownReport generates a human-readable Markdown report for a block file
func GenerateMarkdownReport(result *analysis.FileAnalysisResult) string {
	var sb strings.Builder

	// File overview
	sb.WriteString(fmt.Sprintf("# Chain Analysis Report: %s\n\n", result.File))
	sb.WriteString("---\n\n")

	// Summary section
	sb.WriteString("## Summary\n\n")
	sb.WriteString(fmt.Sprintf("| Metric | Value |\n"))
	sb.WriteString(fmt.Sprintf("|--------|-------|\n"))
	sb.WriteString(fmt.Sprintf("| Source File | %s |\n", result.File))
	sb.WriteString(fmt.Sprintf("| Blocks Analyzed | %d |\n", result.BlockCount))
	sb.WriteString(fmt.Sprintf("| Total Transactions | %d |\n", result.AnalysisSummary.TotalTransactionsAnalyzed))
	sb.WriteString(fmt.Sprintf("| Flagged Transactions | %d |\n", result.AnalysisSummary.FlaggedTransactions))
	sb.WriteString(fmt.Sprintf("| Heuristics Applied | %d |\n", len(result.AnalysisSummary.HeuristicsApplied)))
	sb.WriteString("\n")

	// Heuristics applied
	sb.WriteString("### Heuristics Applied\n\n")
	for _, h := range result.AnalysisSummary.HeuristicsApplied {
		sb.WriteString(fmt.Sprintf("- `%s`\n", h))
	}
	sb.WriteString("\n")

	// Fee Rate statistics
	sb.WriteString("### Fee Rate Distribution\n\n")
	fs := result.AnalysisSummary.FeeRateStats
	sb.WriteString(fmt.Sprintf("| Statistic | Value (sat/vB) |\n"))
	sb.WriteString(fmt.Sprintf("|-----------|----------------|\n"))
	sb.WriteString(fmt.Sprintf("| Minimum | %.1f |\n", fs.MinSatVb))
	sb.WriteString(fmt.Sprintf("| Median | %.1f |\n", fs.MedianSatVb))
	sb.WriteString(fmt.Sprintf("| Mean | %.1f |\n", fs.MeanSatVb))
	sb.WriteString(fmt.Sprintf("| Maximum | %.1f |\n", fs.MaxSatVb))
	sb.WriteString("\n")

	// Script type distribution
	sb.WriteString("### Script Type Distribution\n\n")
	sb.WriteString("| Script Type | Count |\n")
	sb.WriteString("|-------------|-------|\n")

	// Sort script types for deterministic output
	scriptTypes := make([]string, 0, len(result.AnalysisSummary.ScriptTypeDistribution))
	for st := range result.AnalysisSummary.ScriptTypeDistribution {
		scriptTypes = append(scriptTypes, st)
	}
	sort.Strings(scriptTypes)

	for _, st := range scriptTypes {
		sb.WriteString(fmt.Sprintf("| %s | %d |\n", st, result.AnalysisSummary.ScriptTypeDistribution[st]))
	}
	sb.WriteString("\n")

	// Per-block sections
	sb.WriteString("---\n\n")
	sb.WriteString("## Per-Block Analysis\n\n")

	for i, blk := range result.Blocks {
		sb.WriteString(fmt.Sprintf("### Block %d: `%s`\n\n", i+1, blk.BlockHash))

		sb.WriteString(fmt.Sprintf("| Field | Value |\n"))
		sb.WriteString(fmt.Sprintf("|-------|-------|\n"))
		sb.WriteString(fmt.Sprintf("| Block Hash | `%s` |\n", blk.BlockHash))
		sb.WriteString(fmt.Sprintf("| Block Height | %d |\n", blk.BlockHeight))
		sb.WriteString(fmt.Sprintf("| Transaction Count | %d |\n", blk.TxCount))
		sb.WriteString(fmt.Sprintf("| Flagged Transactions | %d |\n", blk.AnalysisSummary.FlaggedTransactions))
		sb.WriteString("\n")

		// Per-block fee rate
		bfs := blk.AnalysisSummary.FeeRateStats
		sb.WriteString("#### Fee Rate Statistics\n\n")
		sb.WriteString(fmt.Sprintf("| Statistic | Value (sat/vB) |\n"))
		sb.WriteString(fmt.Sprintf("|-----------|----------------|\n"))
		sb.WriteString(fmt.Sprintf("| Minimum | %.1f |\n", bfs.MinSatVb))
		sb.WriteString(fmt.Sprintf("| Median | %.1f |\n", bfs.MedianSatVb))
		sb.WriteString(fmt.Sprintf("| Mean | %.1f |\n", bfs.MeanSatVb))
		sb.WriteString(fmt.Sprintf("| Maximum | %.1f |\n", bfs.MaxSatVb))
		sb.WriteString("\n")

		// Per-block script type distribution
		sb.WriteString("#### Script Type Breakdown\n\n")
		sb.WriteString("| Script Type | Count |\n")
		sb.WriteString("|-------------|-------|\n")

		bst := make([]string, 0, len(blk.AnalysisSummary.ScriptTypeDistribution))
		for st := range blk.AnalysisSummary.ScriptTypeDistribution {
			bst = append(bst, st)
		}
		sort.Strings(bst)
		for _, st := range bst {
			sb.WriteString(fmt.Sprintf("| %s | %d |\n", st, blk.AnalysisSummary.ScriptTypeDistribution[st]))
		}
		sb.WriteString("\n")

		// Heuristic findings summary
		sb.WriteString("#### Heuristic Findings\n\n")
		if blk.Transactions != nil {
			heuristicCounts := make(map[string]int)
			classificationCounts := make(map[string]int)

			for _, tx := range blk.Transactions {
				classificationCounts[tx.Classification]++
				for hID, hData := range tx.Heuristics {
					if hMap, ok := hData.(map[string]interface{}); ok {
						if detected, ok := hMap["detected"].(bool); ok && detected {
							heuristicCounts[hID]++
						}
					}
				}
			}

			sb.WriteString("| Heuristic | Transactions Flagged |\n")
			sb.WriteString("|-----------|---------------------|\n")
			hIDs := make([]string, 0, len(heuristicCounts))
			for h := range heuristicCounts {
				hIDs = append(hIDs, h)
			}
			sort.Strings(hIDs)
			for _, h := range hIDs {
				sb.WriteString(fmt.Sprintf("| %s | %d |\n", h, heuristicCounts[h]))
			}
			sb.WriteString("\n")

			// Classification summary
			sb.WriteString("#### Transaction Classifications\n\n")
			sb.WriteString("| Classification | Count |\n")
			sb.WriteString("|---------------|-------|\n")
			cIDs := make([]string, 0, len(classificationCounts))
			for c := range classificationCounts {
				cIDs = append(cIDs, c)
			}
			sort.Strings(cIDs)
			for _, c := range cIDs {
				sb.WriteString(fmt.Sprintf("| %s | %d |\n", c, classificationCounts[c]))
			}
			sb.WriteString("\n")

			// Notable transactions (CoinJoin, Consolidation)
			sb.WriteString("#### Notable Transactions\n\n")
			notableCount := 0
			for _, tx := range blk.Transactions {
				if tx.Classification == "coinjoin" || tx.Classification == "consolidation" || tx.Classification == "batch_payment" {
					if notableCount < 10 { // Limit to 10 notable transactions
						sb.WriteString(fmt.Sprintf("- **%s** `%s`\n", tx.Classification, tx.Txid))
						notableCount++
					}
				}
			}
			if notableCount == 0 {
				sb.WriteString("No notable CoinJoin, consolidation, or batch payment transactions found.\n")
			}
			sb.WriteString("\n")
		}

		sb.WriteString("---\n\n")
	}

	return sb.String()
}
