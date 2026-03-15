# Sherlock - Bitcoin Chain Analysis Engine

<p align="center">
  <img src="https://img.shields.io/badge/Go-1.21+-00ADD8?style=for-the-badge&logo=go" alt="Go Version">
  <img src="https://img.shields.io/badge/React-18+-61DAFB?style=for-the-badge&logo=react" alt="React Version">
  <img src="https://img.shields.io/badge/Bitcoin-Protocol-F7931A?style=for-the-badge&logo=bitcoin" alt="Bitcoin">
</p>

## Overview

Sherlock is a comprehensive Bitcoin chain analysis engine that parses raw blockchain data and applies multiple heuristics to detect transaction patterns, classify behaviors, and identify entities. This document provides complete technical documentation including architecture, data flow, implementation details, and academic references.

---

## Technologies Used

### Technology Stack Overview

```
┌─────────────────────────────────────────────────────────────────────────────────────────┐
│                           TECHNOLOGY STACK                                            │
├─────────────────────────────────────────────────────────────────────────────────────────┤
│                                                                                          │
│  ┌─────────────────────────────────────────────────────────────────────────────────┐        │
│  │                         BACKEND (Go)                                        │        │
│  │  ┌─────────────────┐  ┌─────────────────┐  ┌─────────────────┐        │        │
│  │  │   Go 1.21+     │  │   net/http     │  │   encoding/json │        │        │
│  │  │   Language     │  │   HTTP Server  │  │   JSON Marshal  │        │        │
│  │  └─────────────────┘  └─────────────────┘  └─────────────────┘        │        │
│  └─────────────────────────────────────────────────────────────────────────────────┘        │
│                                                                                          │
│  ┌─────────────────────────────────────────────────────────────────────────────────┐        │
│  │                         FRONTEND (React)                                      │        │
│  │  ┌─────────────────┐  ┌─────────────────┐  ┌─────────────────┐        │        │
│  │  │    React       │  │    Webpack     │  │      CSS       │        │        │
│  │  │    18.x        │  │    Bundler     │  │   Styling      │        │        │
│  │  └─────────────────┘  └─────────────────┘  └─────────────────┘        │        │
│  └─────────────────────────────────────────────────────────────────────────────────┘        │
│                                                                                          │
│  ┌─────────────────────────────────────────────────────────────────────────────────┐        │
│  │                         BUILD TOOLS                                            │        │
│  │  ┌─────────────────┐  ┌─────────────────┐  ┌─────────────────┐        │        │
│  │  │     bash        │  │      go        │  │    node/npm    │        │        │
│  │  │   Scripts       │  │   Build       │  │   JS Bundling │        │        │
│  │  └─────────────────┘  └─────────────────┘  └─────────────────┘        │        │
│  └─────────────────────────────────────────────────────────────────────────────────┘        │
│                                                                                          │
└─────────────────────────────────────────────────────────────────────────────────────────┘
```

### Why These Technologies?

#### Backend: Go (Golang)

```
┌─────────────────────────────────────────────────────────────────────────────────────────┐
│                              WHY GO?                                               │
├─────────────────────────────────────────────────────────────────────────────────────────┤
│                                                                                          │
│  PERFORMANCE:                                                                            │
│  ═══════════                                                                            │
│                                                                                          │
│  • Compiled language - significantly faster than interpreted languages          │
│  • Excellent for processing large binary files (blk.dat can be 100MB+)         │
│  • Low memory footprint - important for parsing many blocks                    │
│  • Built-in concurrency (goroutines) - can parallelize block processing       │
│                                                                                          │
│  MEMORY SAFETY:                                                                         │
│  ══════════════                                                                         │
│                                                                                          │
│  • No buffer overflow vulnerabilities                                                │
│  • Automatic memory management (garbage collection)                                 │
│  • Prevents common security issues in C/C++ parsers                               │
│                                                                                          │
│  CRYPTOGRAPHIC SUPPORT:                                                                │
│  ════════════════════════                                                               │
│                                                                                          │
│  • SHA256, SHA256d needed for block hashing                                        │
│  • Hash160 for address derivation                                                 │
│  • Clean byte manipulation for binary parsing                                     │
│                                                                                          │
│  TOOLING & ECOSYSTEM:                                                                 │
│  ════════════════════                                                                   │
│                                                                                          │
│  • Single binary deployment - no runtime dependencies                                │
│  • Excellent standard library (encoding, crypto, io, net/http)                     │
│  • Fast compilation times                                                          │
│  • Great for CLI tools                                                            │
│                                                                                          │
│  ALTERNATIVES CONSIDERED:                                                            │
│  ════════════════════════                                                             │
│                                                                                          │
│  ❌ Python - Too slow for large file processing                                    │
│  ❌ C/C++ - Memory safety concerns, complex build                                 │
│  ❌ Rust - Steeper learning curve, longer compile times                            │
│  ❌ Node.js - Less suitable for binary parsing                                     │
│                                                                                          │
└─────────────────────────────────────────────────────────────────────────────────────────┘
```

#### Frontend: React

```
┌─────────────────────────────────────────────────────────────────────────────────────────┐
│                           WHY REACT?                                               │
├─────────────────────────────────────────────────────────────────────────────────────────┤
│                                                                                          │
│  COMPONENT-BASED ARCHITECTURE:                                                       │
│  ══════════════════════════════                                                         │
│                                                                                          │
│  • Reusable components: Dashboard, BlockView, TransactionTable, BarChart           │
│  • Each component has single responsibility                                       │
│  • Easy to maintain and extend                                                    │
│                                                                                          │
│  INTERACTIVE UI:                                                                    │
│  ════════════════                                                                     │
│                                                                                          │
│  • State management for filtering/sorting transactions                            │
│  • Dynamic tooltips for educational content                                       │
│  • Real-time updates when loading new blocks                                     │
│                                                                                          │
│  PERFORMANCE:                                                                        │
│  ═══════════                                                                            │
│                                                                                          │
│  • Virtual DOM - efficient rendering of large transaction lists (3000+ txs)      │
│  • Efficient updates - only re-renders changed components                        │
│                                                                                          │
│  TOOLING & COMMUNITY:                                                               │
│  ════════════════════                                                                 │
│                                                                                          │
│  • Huge ecosystem of libraries                                                   │
│  • Excellent developer experience                                                │
│  • Easy to find solutions to common problems                                    │
│                                                                                          │
│  ALTERNATIVES CONSIDERED:                                                            │
│  ════════════════════════                                                             │
│                                                                                          │
│  ❌ Vue - Smaller ecosystem                                                       │
│  ❌ Svelte - Less mature ecosystem                                               │
│  ❌ Vanilla JS - Harder to maintain complex state                                │
│                                                                                          │
└─────────────────────────────────────────────────────────────────────────────────────────┘
```

#### Build System: Webpack

```
┌─────────────────────────────────────────────────────────────────────────────────────────┐
│                          WHY WEBPACK?                                              │
├─────────────────────────────────────────────────────────────────────────────────────────┤
│                                                                                          │
│  CODE BUNDLING:                                                                       │
│  ══════════════                                                                        │
│                                                                                          │
│  • Bundles all JS/CSS into single file (bundle.js)                              │
│  • Tree shaking - removes unused code                                          │
│  • Minification for production                                                  │
│                                                                                          │
│  MODULE SUPPORT:                                                                      │
│  ════════════════                                                                     │
│                                                                                          │
│  • ES6 modules - clean import/export                                            │
│  • CSS module support                                                            │
│  • Asset management (images, fonts)                                              │
│                                                                                          │
│  DEV SERVER:                                                                          │
│  ══════════════                                                                        │
│                                                                                          │
│  • Hot reload during development                                                 │
│  • Fast iteration cycle                                                         │
│                                                                                          │
│  ALTERNATIVES CONSIDERED:                                                            │
│  ════════════════════════                                                             │
│                                                                                          │
│  ❌ Vite - Not available in this project setup                                   │
│  ❌ esbuild - Less mature at project start time                                  │
│  ❌ Rollup - More for libraries, not apps                                        │
│                                                                                          │
└─────────────────────────────────────────────────────────────────────────────────────────┘
```

#### Scripts: Bash

```
┌─────────────────────────────────────────────────────────────────────────────────────────┐
│                           WHY BASH?                                                │
├─────────────────────────────────────────────────────────────────────────────────────────┤
│                                                                                          │
│  STANDARD CLI INTERFACE:                                                             │
│  ════════════════════════                                                            │
│                                                                                          │
│  • Standard Unix/Linux convention for CLI tools                                   │
│  • Works on macOS, Linux, WSL, Git Bash                                         │
│                                                                                          │
│  SCRIPTING CAPABILITIES:                                                            │
│  ════════════════════                                                                │
│                                                                                          │
│  • Argument parsing with getopts                                                  │
│  • File existence checks                                                          │
│  • Directory creation (mkdir -p)                                                  │
│  • Conditional execution                                                           │
│                                                                                          │
│  BUILD AUTOMATION:                                                                  │
│  ═══════════════════                                                                │
│                                                                                          │
│  • Compiles Go binaries                                                           │
│  • Runs npm install                                                              │
│  • Handles errors gracefully                                                     │
│                                                                                          │
│  ALTERNATIVES CONSIDERED:                                                            │
│  ════════════════════════                                                             │
│                                                                                          │
│  ❌ PowerShell - Windows-specific                                                 │
│  ❌ Python - Additional dependency                                                │
│  ❌ Batch - Limited functionality                                                 │
│                                                                                          │
└─────────────────────────────────────────────────────────────────────────────────────────┘
```

### Technology Summary Table

```
┌─────────────────────────────────────────────────────────────────────────────────────────┐
│                        TECHNOLOGY DECISION SUMMARY                                   │
├─────────────────────────────────────────────────────────────────────────────────────────┤
│                                                                                          │
│  Component        │ Technology      │ Version    │ Reason for Choice                 │
│  ────────────────┼─────────────────┼────────────┼─────────────────────────────       │
│  Backend Lang    │ Go              │ 1.21+     │ Performance, memory safety        │
│  HTTP Server     │ net/http        │ Built-in   │ Simple, no external deps        │
│  JSON Handling   │ encoding/json   │ Built-in   │ Native marshaling              │
│  Frontend        │ React           │ 18.2       │ Component-based, interactive    │
│  Build Tool      │ Webpack         │ 5.x        │ Bundling, optimization         │
│  Styling         │ CSS             │ -          │ Custom design                  │
│  CLI Scripts     │ Bash            │ -          │ Standard Unix interface        │
│  Package Manager │ npm             │ -          │ JS dependency management       │
│                                                                                          │
│  RUNTIME REQUIREMENTS:                                                              │
│  ═════════════════════                                                               │
│                                                                                          │
│  • CLI: Go runtime (single binary)                                               │
│  • Web Server: Go runtime + static files                                        │
│  • Frontend: Modern browser (Chrome, Firefox, Safari, Edge)                       │
│                                                                                          │
│  DEPLOYMENT:                                                                        │
│  ═══════════                                                                            │
│                                                                                          │
│  • CLI: Single binary - `go build -o bin/sherlock-cli ./cmd/cli`                  │
│  • Web: Binary + static files - `go build + web/dist/*`                         │
│  • No containers needed - simple deployment                                       │
│                                                                                          │
└─────────────────────────────────────────────────────────────────────────────────────────┘
```

---

## Table of Contents (Detailed)

1. [Technologies Used](#technologies-used)
2. [System Architecture](#system-architecture)
3. [Data Flow Pipeline](#data-flow-pipeline)
4. [Component Details](#component-details)
5. [Heuristics Implementation](#heuristics-implementation)
6. [Confidence Model](#confidence-model)
7. [Glossary](#glossary)
8. [Trade-offs](#trade-offs)
9. [References](#references)

---

## System Architecture

### High-Level System Overview

```
┌─────────────────────────────────────────────────────────────────────────────────────────┐
│                                    SHERLOCK SYSTEM                                       │
│                         Bitcoin Chain Analysis Engine                                     │
├─────────────────────────────────────────────────────────────────────────────────────────┤
│                                                                                          │
│  ┌─────────────────────────────────────────────────────────────────────────────────┐   │
│  │                                 INPUTS                                           │   │
│  │  ┌─────────────┐   ┌─────────────┐   ┌─────────────┐                          │   │
│  │  │   blk*.dat  │   │   rev*.dat  │   │   xor.dat   │                          │   │
│  │  │  (blocks)   │   │   (undo)    │   │  (XOR key)  │                          │   │
│  │  └─────────────┘   └─────────────┘   └─────────────┘                          │   │
│  └─────────────────────────────────────────────────────────────────────────────────┘   │
│                                        │                                                │
│                                        ▼                                                │
│  ┌─────────────────────────────────────────────────────────────────────────────────┐   │
│  │                              PROCESSING LAYER                                     │   │
│  │                                                                                  │   │
│  │    ┌─────────────────────────────────────────────────────────────────────┐     │   │
│  │    │                      BLOCK PARSER                                     │     │   │
│  │    │  ┌──────────┐   ┌──────────┐   ┌──────────┐   ┌──────────┐    │     │   │
│  │    │  │ XOR Decode│──▶│Block Parse│──▶│ Undo Parse│──▶│ TX Enrich │    │     │   │
│  │    │  └──────────┘   └──────────┘   └──────────┘   └──────────┘    │     │   │
│  │    └─────────────────────────────────────────────────────────────────────┘     │   │
│  │                                        │                                          │   │
│  │                                        ▼                                          │   │
│  │    ┌─────────────────────────────────────────────────────────────────────┐     │   │
│  │    │                    HEURISTICS ENGINE                                 │     │   │
│  │    │                                                                      │     │   │
│  │    │   ┌────────┐ ┌────────┐ ┌────────┐ ┌────────┐ ┌────────┐         │     │   │
│  │    │   │ CIOH   │ │Change  │ │Addr    │ │Coin    │ │Consol  │  ...   │     │   │
│  │    │   │        │ │Detect  │ │Reuse   │ │Join    │ │idation │         │     │   │
│  │    │   └───┬────┘ └───┬────┘ └───┬────┘ └───┬────┘ └───┬────┘         │     │   │
│  │    │       │          │          │          │          │                │     │   │
│  │    │       └──────────┴──────────┴──────────┴──────────┘                │     │   │
│  │    │                            │                                        │     │   │
│  │    │                            ▼                                        │     │   │
│  │    │                   ┌──────────────┐                             │     │   │
│  │    │                   │CLASSIFIER     │                             │     │   │
│  │    │                   │(priority-based)│                            │     │   │
│  │    │                   └──────────────┘                             │     │   │
│  │    └─────────────────────────────────────────────────────────────────────┘     │   │
│  │                                        │                                          │   │
│  │                                        ▼                                          │   │
│  │    ┌─────────────────────────────────────────────────────────────────────┐     │   │
│  │    │                   AGGREGATION LAYER                               │     │   │
│  │    │    ┌──────────────────┐    ┌──────────────────┐                   │     │   │
│  │    │    │  Per-Block Stats │    │ File-Level Stats │                   │     │   │
│  │    │    │  • tx_count       │    │  • total_txs    │                   │     │   │
│  │    │    │  • flagged        │    │  • flagged       │                   │     │   │
│  │    │    │  • fee_rates     │    │  • fee_rates    │                   │     │   │
│  │    │    │  • scripts        │    │  • scripts      │                   │     │   │
│  │    │    └──────────────────┘    └──────────────────┘                   │     │   │
│  │    └─────────────────────────────────────────────────────────────────────┘     │   │
│  └─────────────────────────────────────────────────────────────────────────────────┘   │
│                                        │                                                │
│                                        ▼                                                │
│  ┌─────────────────────────────────────────────────────────────────────────────────┐   │
│  │                                OUTPUTS                                          │   │
│  │                                                                                  │   │
│  │  ┌─────────────────────┐                      ┌─────────────────────┐          │   │
│  │  │    JSON OUTPUT     │                      │   MARKDOWN REPORT  │          │   │
│  │  │   out/blk*.json   │                      │    out/blk*.md    │          │   │
│  │  │                    │                      │                     │          │   │
│  │  │ {                 │                      │ # Chain Analysis    │          │   │
│  │  │   "ok": true,    │                      │ ## Summary         │          │   │
│  │  │   "blocks": [...],│                      │ ### Block 1: ...  │          │   │
│  │  │   "analysis": {...}                      │ ### Block 2: ...  │          │   │
│  │  │ }                 │                      │                    │          │   │
│  │  └─────────────────────┘                      └─────────────────────┘          │   │
│  │            │                                          │                      │   │
│  │            └──────────────────┬───────────────────────┘                      │   │
│  │                               ▼                                              │   │
│  │                    ┌─────────────────────┐                                   │   │
│  │                    │    WEB UI (React)  │                                   │   │
│  │                    │  • Dashboard        │                                   │   │
│  │                    │  • Block Explorer   │                                   │   │
│  │                    │  • Transaction View│                                   │   │
│  │                    │  • Tooltips        │                                   │   │
│  │                    └─────────────────────┘                                   │   │
│  └─────────────────────────────────────────────────────────────────────────────────┘   │
└─────────────────────────────────────────────────────────────────────────────────────────┘
```

### Code Organization Structure

```
┌─────────────────────────────────────────────────────────────────────────────────────────┐
│                              PROJECT STRUCTURE                                          │
├─────────────────────────────────────────────────────────────────────────────────────────┤
│                                                                                          │
│  sherlock/                                                                              │
│  ├── cmd/                                # Command-line interfaces                      │
│  │   ├── cli/                           │   # CLI entry point                         │
│  │   │   └── main.go                    │       └── ./cli.sh --block <files>          │
│  │   │                                  │                                              │
│  │   └── web/                          │   # Web server entry point                   │
│  │       └── main.go                    │       └── ./web.sh (serves :3000)           │
│  │                                                                                      │
│  ├── internal/                          # Core packages (not imported externally)          │
│  │   ├── parser/                       │   # Binary data parsing                       │
│  │   │   ├── reader.go                 │       └── Sequential byte reader              │
│  │   │   └── tx.go                     │       └── Transaction parsing                 │
│  │   │                                  │                                              │
│  │   ├── types/                        │   # Data structures                           │
│  │   │   └── types.go                  │       └── RawTransaction, TxInput, TxOutput   │
│  │   │                                  │                                              │
│  │   ├── crypto/                       │   # Cryptographic operations                  │
│  │   │   └── hash.go                   │       └── SHA256d, Hash160, TxID             │
│  │   │                                  │                                              │
│  │   ├── script/                       │   # Script analysis                          │
│  │   │   ├── classify.go               │       └── P2PKH/P2SH/P2WPKH/P2TR detection   │
│  │   │   └── opreturn.go               │       └── OP_RETURN protocol decoding         │
│  │   │                                  │                                              │
│  │   ├── address/                      │   # Address derivation                        │
│  │   │   └── address.go                │       └── Hex → human-readable address       │
│  │   │                                  │                                              │
│  │   ├── block/                        │   # Block file processing                    │
│  │   │   └── block.go                  │       └── ProcessBlockFiles(), XOR decode    │
│  │   │                                  │                                              │
│  │   ├── heuristics/                   │   # Chain analysis heuristics                 │
│  │   │   ├── heuristics.go              │       └── All 9 heuristics implementation    │
│  │   │   └── heuristics_test.go         │       └── Unit tests                         │
│  │   │                                  │                                              │
│  │   ├── analysis/                     │   # Analysis orchestration                    │
│  │   │   └── analyzer.go                │       └── AnalyzeBlocks(), JSON output       │
│  │   │                                  │                                              │
│  │   └── report/                       │   # Report generation                        │
│  │       └── report.go                  │       └── GenerateMarkdownReport()            │
│  │                                                                                      │
│  ├── web/                              # Web UI                                        │
│  │   ├── ui/                           │   # React source code                         │
│  │   │   ├── src/                     │       └── App.jsx, components/               │
│  │   │   │   ├── components/           │           ├── Dashboard.jsx                  │
│  │   │   │   │   ├── BlockView.jsx     │           ├── TransactionTable.jsx           │
│  │   │   │   │   ├── TransactionGraph.jsx        ├── FileUpload.jsx                   │
│  │   │   │   │   ├── BarChart.jsx     │           ├── Tooltip.jsx                     │
│  │   │   │   │   └── Definitions.jsx  │           └── StatCard.jsx                   │
│  │   │   │   ├── App.jsx               │                                              │
│  │   │   │   ├── main.jsx              │                                              │
│  │   │   │   └── index.css             │                                              │
│  │   │   └── package.json              │                                              │
│  │   └── dist/                         │   # Built static files                        │
│  │       ├── index.html                 │                                              │
│  │       └── bundle.js                  │                                              │
│  │                                                                                      │
│  ├── bin/                              # Compiled binaries                             │
│  │   ├── sherlock-cli                  │   # CLI binary                                 │
│  │   └── sherlock-web                  │   # Web server binary                         │
│  │                                                                                      │
│  ├── fixtures/                         # Test data                                      │
│  │   ├── blk04330.dat                  │   # Block data                                │
│  │   ├── blk05051.dat                  │                                              │
│  │   ├── rev04330.dat                  │   # Undo data                                │
│  │   ├── rev05051.dat                  │                                              │
│  │   └── xor.dat                       │   # XOR key                                   │
│  │                                                                                      │
│  ├── out/                              # Output directory                              │
│  │   ├── blk04330.json                 │   # Analysis JSON                             │
│  │   ├── blk04330.md                   │   # Markdown report                           │
│  │   ├── blk05051.json                 │                                              │
│  │   └── blk05051.md                   │                                              │
│  │                                                                                      │
│  ├── cli.sh                            # CLI wrapper script                           │
│  ├── web.sh                           # Web server wrapper script                     │
│  ├── setup.sh                         # Setup/installation script                      │
│  ├── APPROACH.md                      # This documentation                            │
│  └── demo.md                         # Demo video link                                │
│                                                                                          │
└─────────────────────────────────────────────────────────────────────────────────────────┘
```

---

## Data Flow Pipeline

### Step-by-Step Data Processing

```
┌─────────────────────────────────────────────────────────────────────────────────────────┐
│                              DATA FLOW: RAW FILES TO OUTPUT                              │
├─────────────────────────────────────────────────────────────────────────────────────────┤
│                                                                                          │
│  STEP 1: INPUT FILES                                                                  │
│  ════════════════════                                                                  │
│                                                                                          │
│    blk04330.dat          rev04330.dat          xor.dat                                    │
│  ┌──────────────┐     ┌──────────────┐     ┌──────────┐                               │
│  │ 0xf9beb4d9   │     │ 0xf9beb4d9   │     │ 0xAB    │  (magic number + data)   │
│  │ (magic)      │     │ (magic)      │     │ (key)    │                               │
│  │ + block data │     │ + undo data  │     │          │                               │
│  └──────┬───────┘     └──────┬───────┘     └────┬─────┘                               │
│         │                    │                   │                                      │
│         └────────────────────┼───────────────────┘                                      │
│                              ▼                                                           │
│  STEP 2: XOR DECODE                                                                   │
│  ════════════════════                                                                   │
│                                                                                          │
│    For each byte in blk.dat and rev.dat:                                               │
│      decoded[i] = encoded[i] XOR key[i % keylen]                                       │
│                                                                                          │
│  ┌─────────────────────────────────────────────────────────────────┐                    │
│  │  XOR DECODER (internal/block/block.go)                         │                    │
│  │                                                                 │                    │
│  │  Input: blk.dat + xor.dat    Output: decoded blk data         │                    │
│  │  Input: rev.dat + xor.dat    Output: decoded undo data        │                    │
│  └─────────────────────────────────────────────────────────────────┘                    │
│                              │                                                          │
│                              ▼                                                          │
│  STEP 3: BLOCK PARSING                                                                │
│  ═══════════════════════                                                                │
│                                                                                          │
│  ┌─────────────────────────────────────────────────────────────────┐                    │
│  │  PARSE BLOCKS (internal/block/block.go)                        │                    │
│  │                                                                 │                    │
│  │  For each block in decoded data:                                │                    │
│  │    1. Read magic number (4 bytes) ──▶ 0xf9beb4d9               │                    │
│  │    2. Read block size (4 bytes)                                │                    │
│  │    3. Read 80-byte header:                                     │                    │
│  │       • version (4 bytes)                                      │                    │
│  │       • prevblockhash (32 bytes)                              │                    │
│  │       • merkleroot (32 bytes)                                 │                    │
│  │       • timestamp (4 bytes)                                     │                    │
│  │       • bits (4 bytes)                                         │                    │
│  │       • nonce (4 bytes)                                        │                    │
│  │    4. Compute block_hash = SHA256d(header)                   │                    │
│  │    5. Parse transactions:                                      │                    │
│  │       • tx_count (varint)                                      │                    │
│  │       • For each tx:                                           │                    │
│  │         - inputs (count + data)                               │                    │
│  │         - outputs (count + data)                               │                    │
│  │         - locktime (4 bytes)                                  │                    │
│  └─────────────────────────────────────────────────────────────────┘                    │
│                              │                                                          │
│                              ▼                                                          │
│  STEP 4: UNDO DATA PARSING                                                            │
│  ═══════════════════════════                                                            │
│                                                                                          │
│  ┌─────────────────────────────────────────────────────────────────┐                    │
│  │  PARSE UNDO DATA (internal/block/block.go)                     │                    │
│  │                                                                 │                    │
│  │  Undo data provides "what was spent" for each input:           │                    │
│  │                                                                 │                    │
│  │  For each transaction in block:                                  │                    │
│  │    For each input (except coinbase):                           │                    │
│  │      • prevout value (compressed)                              │                    │
│  │      • prevout script (compressed)                             │                    │
│  │                                                                 │                    │
│  │  CRITICAL: This is needed to calculate fees:                   │                    │
│  │                                                                 │                    │
│  │    Fee = Σ(input_values) - Σ(output_values)                  │                    │
│  └─────────────────────────────────────────────────────────────────┘                    │
│                              │                                                          │
│                              ▼                                                          │
│  STEP 5: TRANSACTION ENRICHMENT                                                        │
│  ══════════════════════════════                                                         │
│                                                                                          │
│  ┌─────────────────────────────────────────────────────────────────┐                    │
│  │  ENRICH TRANSACTIONS (internal/block/block.go)                  │                    │
│  │                                                                 │                    │
│  │  For each transaction:                                          │                    │
│  │    1. Compute TXID = SHA256d(legacy_serialization)             │                    │
│  │    2. Classify output scripts:                                 │                    │
│  │         • 0x76a914... → P2PKH                                 │                    │
│  │         • 0xa914... → P2SH                                    │                    │
│  │         • 0x0014... → P2WPKH                                 │                    │
│  │         • 0x0020... → P2WSH                                   │                    │
│  │         • 0x5120... → P2TR                                    │                    │
│  │    3. Derive addresses from scripts                            │                    │
│  │    4. Calculate fees:                                           │                    │
│  │         • total_input_sats = Σ(prevout values)                │                    │
│  │         • total_output_sats = Σ(output values)                │                    │
│  │         • fee_sats = inputs - outputs                         │                    │
│  │    5. Calculate fee rate:                                       │                    │
│  │         • weight = witness*4 + non_witness                     │                    │
│  │         • vbytes = weight / 4                                  │                    │
│  │         • fee_rate = fee_sats / vbytes                       │                    │
│  └─────────────────────────────────────────────────────────────────┘                    │
│                              │                                                          │
│                              ▼                                                          │
│  STEP 6: BUILD ADDRESS MAP                                                           │
│  ══════════════════════════                                                            │
│                                                                                          │
│  ┌─────────────────────────────────────────────────────────────────┐                    │
│  │  BLOCK ADDRESS MAP (internal/heuristics/heuristics.go)          │                    │
│  │                                                                 │                    │
│  │  Build: Map<address, [{tx_idx, is_input, is_output}]>         │                    │
│  │                                                                 │                    │
│  │  Purpose: Detect CROSS-TRANSACTION address reuse                │                    │
│  │                                                                 │                    │
│  │  Example:                                                       │                    │
│  │    Tx1: A → B                                                 │                    │
│  │    Tx2: B → C  (B appears in tx2 input, was output in tx1)  │                    │
│  │    → Address reuse detected!                                  │                    │
│  └─────────────────────────────────────────────────────────────────┘                    │
│                              │                                                          │
│                              ▼                                                          │
│  STEP 7: APPLY HEURISTICS                                                             │
│  ════════════════════════                                                                │
│                                                                                          │
│  ┌─────────────────────────────────────────────────────────────────┐                    │
│  │  HEURISTICS ENGINE (internal/heuristics/heuristics.go)         │                    │
│  │                                                                 │                    │
│  │  For each transaction, apply all 9 heuristics:                │                    │
│  │                                                                 │                    │
│  │  ┌─────────────────────────────────────────────────────────┐   │                    │
│  │  │ 1. CIOH (Common Input Ownership)                       │   │                    │
│  │  │    if len(inputs) > 1: detected = true               │   │                    │
│  │  ├─────────────────────────────────────────────────────────┤   │                    │
│  │  │ 2. Change Detection                                     │   │                    │
│  │  │    6 methods: script_type, optimal_change, round...   │   │                    │
│  │  ├─────────────────────────────────────────────────────────┤   │                    │
│  │  │ 3. Address Reuse                                        │   │                    │
│  │  │    Check: input_addr == output_addr                   │   │                    │
│  │  ├─────────────────────────────────────────────────────────┤   │                    │
│  │  │ 4. CoinJoin Detection                                   │   │                    │
│  │  │    if 3+ equal outputs and 3+ inputs: detected=true  │   │                    │
│  │  ├─────────────────────────────────────────────────────────┤   │                    │
│  │  │ 5. Consolidation                                        │   │                    │
│  │  │    if inputs >= 5 and outputs <= 2: detected=true     │   │                    │
│  │  ├─────────────────────────────────────────────────────────┤   │                    │
│  │  │ 6. Self-Transfer                                         │   │                    │
│  │  │    if outputs == input_types and no round amounts:     │   │                    │
│  │  ├─────────────────────────────────────────────────────────┤   │                    │
│  │  │ 7. Peeling Chain                                        │   │                    │
│  │  │    if 1 input and 2 outputs and ratio < 0.1:         │   │                    │
│  │  ├─────────────────────────────────────────────────────────┤   │                    │
│  │  │ 8. OP_RETURN Analysis                                   │   │                    │
│  │  │    if script[0] == 0x6a: detected=true              │   │                    │
│  │  ├─────────────────────────────────────────────────────────┤   │                    │
│  │  │ 9. Round Number Payment                                 │   │                    │
│  │  │    if output % 100000 == 0: round=true               │   │                    │
│  │  └─────────────────────────────────────────────────────────┘   │                    │
│  │                                                                 │                    │
│  │  AFTER all heuristics: Apply cross-heuristic interactions      │                    │
│  │    • If CoinJoin: downgrade CIOH to low                       │                    │
│  │    • If Peeling Chain: set change to larger output           │                    │
│  │    • If Consolidation: suppress self-transfer                │                    │
│  └─────────────────────────────────────────────────────────────────┘                    │
│                              │                                                          │
│                              ▼                                                          │
│  STEP 8: CLASSIFY TRANSACTION                                                         │
│  ══════════════════════════════                                                         │
│                                                                                          │
│  ┌─────────────────────────────────────────────────────────────────┐                    │
│  │  TRANSACTION CLASSIFIER                                        │                    │
│  │                                                                 │                    │
│  │  Priority order (first match wins):                             │                    │
│  │                                                                 │                    │
│  │    ┌─────────────────────────────────────────────────────────┐│                    │
│  │    │ 1. coinjoin        ← if CoinJoin detected              ││                    │
│  │    │ 2. consolidation   ← if Consolidation detected           ││                    │
│  │    │ 3. self_transfer  ← if SelfTransfer detected           ││                    │
│  │    │ 4. batch_payment  ← 3+ outputs, ≤2 inputs            ││                    │
│  │    │ 5. simple_payment ← 1 input, ≤2 outputs              ││                    │
│  │    │ 6. unknown        ← otherwise                          ││                    │
│  │    └─────────────────────────────────────────────────────────┘│                    │
│  └─────────────────────────────────────────────────────────────────┘                    │
│                              │                                                          │
│                              ▼                                                          │
│  STEP 9: AGGREGATE STATISTICS                                                         │
│  ═══════════════════════════                                                            │
│                                                                                          │
│  ┌─────────────────────────────────────────────────────────────────┐                    │
│  │  STATISTICS AGGREGATION (internal/analysis/analyzer.go)        │                    │
│  │                                                                 │                    │
│  │  Per-Block:                                                     │                    │
│  │    • tx_count = len(transactions)                               │                    │
│  │    • flagged_transactions = count(tx with any heuristic)       │                    │
│  │    • script_type_distribution = {type: count}                  │                    │
│  │    • fee_rate_stats = {min, max, median, mean}              │                    │
│  │    • heuristic_counts = {heuristic: count}                    │                    │
│  │    • classification_counts = {class: count}                    │                    │
│  │                                                                 │                    │
│  │  File-Level:                                                    │                    │
│  │    • Sum all per-block values                                   │                    │
│  │    • Aggregate fee rates from all transactions                  │                    │
│  └─────────────────────────────────────────────────────────────────┘                    │
│                              │                                                          │
│                              ▼                                                          │
│  STEP 10: OUTPUT                                                                     │
│  ══════════════════                                                                    │
│                                                                                          │
│    ┌─────────────────────┐           ┌─────────────────────┐                         │
│    │   JSON OUTPUT       │           │   MARKDOWN REPORT   │                         │
│    │   out/blk*.json    │           │   out/blk*.md      │                         │
│    └─────────────────────┘           └─────────────────────┘                         │
│              │                                   │                                      │
│              │                                   │                                      │
│              ▼                                   ▼                                      │
│    {                           # Chain Analysis Report: blk04330.dat        │
│      "ok": true,          ──────────────────────────────────────        │
│      "mode": "chain_     │                                             │
│        analysis",          │ ## Summary                                   │
│      "file": "blk04330    │ | Metric | Value |                         │
│        .dat",            │ |--------|-------|                         │
│      "block_count": 84,  │ | Blocks | 84 |                           │
│      "analysis_summary": │ ...                                         │
│        {...},            │                                             │
│      "blocks": [...]     │                                             │
│    }                     │                                             │
│                                                                                          │
└─────────────────────────────────────────────────────────────────────────────────────────┘
```

---

## Component Details

### CLI Component Flow

```
┌─────────────────────────────────────────────────────────────────────────────────────────┐
│                              CLI WORKFLOW                                               │
├─────────────────────────────────────────────────────────────────────────────────────────┤
│                                                                                          │
│  $ ./cli.sh --block blk04330.dat rev04330.dat xor.dat                                  │
│                                                                                          │
│  ┌────────────────────────────────────────────────────────────────────────────────────┐  │
│  │ cli.sh                                                                         │  │
│  │  1. Parse arguments                                                            │  │
│  │  2. Validate files exist                                                       │  │
│  │  3. Create out/ directory                                                      │  │
│  │  4. Build sherlock-cli (if needed)                                            │  │
│  │  5. exec sherlock-cli --block <files>                                         │  │
│  └────────────────────────────────────────────────────────────────────────────────────┘  │
│                                        │                                                │
│                                        ▼                                                │
│  ┌────────────────────────────────────────────────────────────────────────────────────┐  │
│  │ sherlock-cli (cmd/cli/main.go)                                                 │  │
│  │                                                                                  │  │
│  │  1. ParseArgs()                                                                │  │
│  │     - Check --block flag                                                        │  │
│  │     - Get 3 file paths                                                         │  │
│  │     - Validate files exist                                                       │  │
│  │                                                                                  │  │
│  │  2. block.ProcessBlockFiles(blkPath, revPath, xorPath)                         │  │
│  │     ├── Read xor.dat (XOR key)                                                 │  │
│  │     ├── Read blk.dat & XOR decode                                              │  │
│  │     ├── Read rev.dat & XOR decode                                              │  │
│  │     ├── Parse blocks                                                            │  │
│  │     └── Return []ParsedBlock                                                    │  │
│  │                                                                                  │  │
│  │  3. analysis.AnalyzeBlocks(parsedBlocks, filename, false)                        │  │
│  │     ├── For each block:                                                        │  │
│  │     │   ├── Build address map                                                   │  │
│  │     │   ├── For each tx: apply 9 heuristics                                   │  │
│  │     │   ├── Classify tx                                                        │  │
│  │     │   └── Aggregate stats                                                     │  │
│  │     └── Return *FileAnalysisResult                                              │  │
│  │                                                                                  │  │
│  │  4. WriteJSON(result) → out/blk*.json                                          │  │
│  │  5. report.GenerateMarkdownReport(result) → out/blk*.md                        │  │
│  │  6. Print success message                                                       │  │
│  └────────────────────────────────────────────────────────────────────────────────────┘  │
│                                        │                                                │
│                                        ▼                                                │
│  Output:                                                                               │
│  ┌────────────────────────────────────────────────────────────────────────────────────┐  │
│  │ stderr:                                                                         │  │
│  │   Wrote out/blk04330.json (XXX bytes)                                          │  │
│  │   Wrote out/blk04330.md (XXX bytes)                                            │  │
│  │   Chain analysis complete for blk04330.dat                                      │  │
│  └────────────────────────────────────────────────────────────────────────────────────┘  │
│                                                                                          │
└─────────────────────────────────────────────────────────────────────────────────────────┘
```

### Web Server Flow

```
┌─────────────────────────────────────────────────────────────────────────────────────────┐
│                              WEB SERVER WORKFLOW                                        │
├─────────────────────────────────────────────────────────────────────────────────────────┤
│                                                                                          │
│  $ ./web.sh                                                                             │
│                                                                                          │
│  ┌────────────────────────────────────────────────────────────────────────────────────┐  │
│  │ web.sh                                                                           │  │
│  │  1. Set PORT (default 3000)                                                    │  │
│  │  2. Build sherlock-web (if needed)                                             │  │
│  │  3. exec sherlock-web                                                         │  │
│  └────────────────────────────────────────────────────────────────────────────────────┘  │
│                                        │                                                │
│                                        ▼                                                │
│  ┌────────────────────────────────────────────────────────────────────────────────────┐  │
│  │ sherlock-web (cmd/web/main.go)                                                 │  │
│  │                                                                                  │  │
│  │  HTTP Server listening on 127.0.0.1:3000                                        │  │
│  │                                                                                  │  │
│  │  ┌──────────────────────────────────────────────────────────────────────────┐    │  │
│  │  │ GET /api/health                                                        │    │  │
│  │  │   → Returns: { "ok": true }                                          │    │  │
│  │  └──────────────────────────────────────────────────────────────────────────┘    │  │
│  │                                                                                  │  │
│  │  ┌──────────────────────────────────────────────────────────────────────────┐    │  │
│  │  │ GET /api/blocks                                                       │    │  │
│  │  │   → Lists all .json files in out/                                    │    │  │
│  │  │   → Returns: { "ok": true, "files": ["blk04330", "blk05051"] }      │    │  │
│  │  └──────────────────────────────────────────────────────────────────────────┘    │  │
│  │                                                                                  │  │
│  │  ┌──────────────────────────────────────────────────────────────────────────┐    │  │
│  │  │ GET /api/blocks/blk04330                                              │    │  │
│  │  │   → Reads out/blk04330.json                                          │    │  │
│  │  │   → Adds timestamps (estimated from block height)                     │    │  │
│  │  │   → Returns JSON content                                              │    │  │
│  │  └──────────────────────────────────────────────────────────────────────────┘    │  │
│  │                                                                                  │  │
│  │  ┌──────────────────────────────────────────────────────────────────────────┐    │  │
│  │  │ POST /api/upload                                                        │    │  │
│  │  │   1. Parse multipart form (blk, rev, xor files)                       │    │  │
│  │  │   2. Save to temp directory                                           │    │  │
│  │  │   3. block.ProcessBlockFiles(...)                                      │    │  │
│  │  │   4. analysis.AnalyzeBlocks(..., includeAllTx=true)                    │    │  │
│  │  │   5. Add timestamps (actual from parsed blocks)                         │    │  │
│  │  │   6. Write to out/                                                    │    │  │
│  │  │   7. Return result to client                                           │    │  │
│  │  └──────────────────────────────────────────────────────────────────────────┘    │  │
│  │                                                                                  │  │
│  │  ┌──────────────────────────────────────────────────────────────────────────┐    │  │
│  │  │ GET /* (SPA handler)                                                  │    │  │
│  │  │   → Serve static files from web/dist/                                 │    │  │
│  │  │   → index.html for unknown paths                                       │    │  │
│  │  └──────────────────────────────────────────────────────────────────────────┘    │  │
│  └────────────────────────────────────────────────────────────────────────────────────┘  │
│                                        │                                                │
│                                        ▼                                                │
│  Browser accesses http://127.0.0.1:3000                                               │
│                                                                                          │
│  ┌────────────────────────────────────────────────────────────────────────────────────┐  │
│  │                          REACT FRONTEND                                         │  │
│  │                                                                                  │  │
│  │  ┌────────────┐    ┌────────────┐    ┌────────────┐    ┌────────────┐     │  │
│  │  │ Dashboard  │───▶│ BlockView  │───▶│  TX Table  │───▶│ TX Graph   │     │  │
│  │  │ (overview) │    │ (per-block)│    │ (filtered) │    │ (visual)   │     │  │
│  │  └────────────┘    └────────────┘    └────────────┘    └────────────┘     │  │
│  │       │                  │                  │                  │             │  │
│  │       └──────────────────┴──────────────────┴──────────────────┘             │  │
│  │                                    │                                          │  │
│  │                                    ▼                                          │  │
│  │                         ┌────────────────────┐                               │  │
│  │                         │     Tooltips       │                               │  │
│  │                         │  (Definitions for  │                               │  │
│  │                         │   non-technical    │                               │  │
│  │                         │   users)          │                               │  │
│  │                         └────────────────────┘                               │  │
│  └────────────────────────────────────────────────────────────────────────────────────┘  │
│                                                                                          │
└─────────────────────────────────────────────────────────────────────────────────────────┘
```

---

## Heuristics Implementation

### Heuristic Categories Overview

```
┌─────────────────────────────────────────────────────────────────────────────────────────┐
│                        9 HEURISTICS IMPLEMENTED                                        │
├─────────────────────────────────────────────────────────────────────────────────────────┤
│                                                                                          │
│  ┌─────────────────────────────────────────────────────────────────────────────────┐     │
│  │                    CLUSTERING HEURISTICS                                      │     │
│  ├───────────────────────────────────────────────────────────────────────────────┤     │
│  │  1. CIOH (Common Input Ownership)                                           │     │
│  │     ─────────────────────────────────────                                     │     │
│  │     Detects: Multiple inputs → same wallet                                   │     │
│  │     Logic:   len(inputs) > 1                                                │     │
│  │     Confidence: high (>3 inputs), medium (2-3), low (CoinJoin detected)     │     │
│  │                                                                               │     │
│  │  2. Address Reuse                                                          │     │
│  │     ─────────────────────────────────────                                     │     │
│  │     Detects: Same address in inputs/outputs across txs                       │     │
│  │     Logic:   Cross-reference address map                                     │     │
│  │     Confidence: always high                                                   │     │
│  └─────────────────────────────────────────────────────────────────────────────────┘     │
│                                                                                          │
│  ┌─────────────────────────────────────────────────────────────────────────────────┐     │
│  │                    CHANGE DETECTION HEURISTICS                               │     │
│  ├───────────────────────────────────────────────────────────────────────────────┤     │
│  │  3. Change Detection (6 methods)                                           │     │
│  │     ─────────────────────────────────────                                     │     │
│  │     • Script type matching (high)                                          │     │
│  │     • Optimal change (high)                                                 │     │
│  │     • Round number (medium)                                                 │     │
│  │     • nLockTime fingerprint (medium)                                        │     │
│  │     • Fresh address (low)                                                   │     │
│  │     • Value analysis (low)                                                  │     │
│  │                                                                               │     │
│  │  4. Round Number Payment                                                    │     │
│  │     ─────────────────────────────────────                                     │     │
│  │     Detects: Round BTC amounts = human-chosen payments                       │     │
│  │     Logic:   value % 100000 == 0                                           │     │
│  │     Confidence: high (≥0.1 BTC), medium, low                                │     │
│  └─────────────────────────────────────────────────────────────────────────────────┘     │
│                                                                                          │
│  ┌─────────────────────────────────────────────────────────────────────────────────┐     │
│  │                    PRIVACY HEURISTICS                                        │     │
│  ├───────────────────────────────────────────────────────────────────────────────┤     │
│  │  5. CoinJoin Detection                                                     │     │
│  │     ─────────────────────────────────────                                     │     │
│  │     Detects: Multi-party privacy transactions                               │     │
│  │     Logic:   3+ inputs, 3+ outputs, equal values                          │     │
│  │     Confidence: high (confirmed), medium (possible)                        │     │
│  │                                                                               │     │
│  │  6. Self-Transfer                                                          │     │
│  │     ─────────────────────────────────────                                     │     │
│  │     Detects: Funds moved between own addresses                              │     │
│  │     Logic:   output_types == input_types, no round amounts                 │     │
│  │     Confidence: medium                                                      │     │
│  │                                                                               │     │
│  │  7. Consolidation                                                          │     │
│  │     ─────────────────────────────────────                                     │     │
│  │     Detects: Many UTXOs combined                                            │     │
│  │     Logic:   inputs >= 5, outputs <= 2                                     │     │
│  │     Confidence: high (10+), medium (5-9)                                   │     │
│  │                                                                               │     │
│  │  8. Peeling Chain                                                          │     │
│  │     ─────────────────────────────────────                                     │     │
│  │     Detects: Large UTXO progressively peeled                                │     │
│  │     Logic:   1 input, 2 outputs, ratio < 0.1                               │     │
│  │     Confidence: high (<1%), medium, low                                     │     │
│  └─────────────────────────────────────────────────────────────────────────────────┘     │
│                                                                                          │
│  ┌─────────────────────────────────────────────────────────────────────────────────┐     │
│  │                    DATA HEURISTICS                                            │     │
│  ├───────────────────────────────────────────────────────────────────────────────┤     │
│  │  9. OP_RETURN Analysis                                                    │     │
│  │     ─────────────────────────────────────                                     │     │
│  │     Detects: Data carrier outputs                                           │     │
│  │     Logic:   opcode == 0x6a                                               │     │
│  │     Protocols: Omni (6f6d6e69), OpenAssets (4f41)                        │     │
│  │     Confidence: high                                                       │     │
│  └─────────────────────────────────────────────────────────────────────────────────┘     │
│                                                                                          │
└─────────────────────────────────────────────────────────────────────────────────────────┘
```

### Heuristic Implementation Details

## Heuristics Implemented

All 9 heuristics from the challenge catalogue are implemented. Each heuristic includes a research-backed confidence model, documented limitations, and proper academic citations.

### 1. Common Input Ownership Heuristic (CIOH)

**What it detects:** All inputs to a transaction likely belong to the same entity (wallet). This is the foundational assumption of chain analysis, first mentioned in the Bitcoin whitepaper (Nakamoto, 2008, Section 10) and formally defined by Meiklejohn et al. (2013).

**How it is detected/computed:** A transaction is flagged when it has more than one input (`len(inputs) > 1`). Every multi-input transaction inherently combines UTXOs, and wallet software automatically selects and combines UTXOs from the same wallet.

**Confidence model:**
- **high:** >3 inputs — a strong cluster signal; wallet software is aggregating many UTXOs from the same key pool.
- **medium:** 2-3 inputs — weaker signal; could be a small CoinJoin or PayJoin with few participants.

**Cross-heuristic interaction:** If CoinJoin is also detected on the same transaction, CIOH confidence is downgraded to low, because CoinJoin inputs belong to different entities by design. This is backed by Gong et al. (2022), who showed that "the multi-input (MI) heuristic has higher false positives due to CoinJoin and mixing."

**Known limitations:**
- False positives for CoinJoin transactions where multiple parties contribute inputs (Schnoering & Vazirgiannis, 2023)
- False positives for PayJoin (Ghesmati et al., 2021), where the privacy technique deliberately breaks CIOH
- False negatives for single-input transactions, even if the entity controls many UTXOs
- MtGox historically imported user private keys, causing cluster collapse when applied naively (Ron & Shamir, 2012, documented in "The Scroll" by Bitcoin Magazine)

### 2. Change Detection

**What it detects:** Identifies the likely change output — the output that returns leftover funds to the sender's wallet rather than going to the payment recipient.

**How it is detected/computed:** Six methods are applied in priority order (highest confidence first):

1. **Script type matching (high confidence):** If exactly one output matches the predominant input script type and others don't, that output is likely change. Wallets typically send change to the same address type. Reference: BlockSci change_by_address_type — "If all inputs are of one address type, it is likely that the change output has the same type" (Kalodner et al., 2020).

2. **Optimal change (high confidence):** If an output is smaller than the smallest input, it is likely change. The coin selection algorithm would not include an input if the resulting change exceeded that input's value. Reference: BlockSci change_by_optimal_change — "If there exists an output that is smaller than any of the inputs it is likely the change" (Kalodner et al., 2020).

3. **Round number analysis (medium confidence):** For 2-output transactions, if one output is a round BTC amount and the other isn't, the non-round output is likely change. Human-chosen payment amounts tend to be round numbers. Reference: Androulaki et al. (2012) first proposed round-value-based change detection using their "Shadow Addresses" framework.

4. **nLockTime wallet fingerprinting (medium confidence):** Bitcoin Core sets nLockTime to the current block height to prevent fee sniping. If tx.Locktime matches the block height (±2 for reorg tolerance), the transaction was likely created by Bitcoin Core. Combined with script type matching, the output matching the predominant input type is identified as change. Reference: Möser & Narayanan (2021) — wallet fingerprinting via nLockTime patterns; BlockSci change_by_locktime.

5. **Fresh-address heuristic (low confidence):** Wallet software generates fresh addresses for change outputs. Within a block, if an output address appears only once across all transactions (i.e., it's "fresh" at the block level), and it's the only fresh output in a 2-output transaction, it is likely change. Reference: Meiklejohn et al. (2013). Limitation: True freshness requires full blockchain history; block-level is a weaker proxy, hence low confidence.

6. **Value analysis (low confidence):** For 2-output transactions where all other methods don't apply, the smaller output is tentatively identified as change — a purely statistical last-resort fallback and the weakest signal.

**Cross-heuristic interaction:** If peeling chain detection fires and change detection did not independently identify change, the larger output is assigned as the likely change (the peeler's remaining balance). Reference: BlockSci change_by_peeling_chain.

**Confidence model:**
- **high:** script_type_match or optimal_change — strong structural signals
- **medium:** round_number or locktime_fingerprint — reasonable but not definitive
- **low:** fresh_address or value_analysis — weakest signals, block-level proxy or statistical fallback

**Known limitations:**
- Address reuse rate has dropped below 10% on modern Bitcoin, weakening fresh-address-based change heuristics (Gong et al., 2025)
- Fails when both payment and change use the same script type (common with HD wallets)
- Privacy-conscious wallets may add noise to output amounts
- Batch payments have no change output, leading to false detections
- nLockTime fingerprinting only works for Bitcoin Core; other wallets set nLockTime = 0
- Fresh-address check is block-level only — full-chain freshness would be more accurate

### 3. Address Reuse

**What it detects:** Detects when the same address appears in both inputs and outputs of a transaction, or across multiple transactions within the same block. Address reuse is a definitive privacy weakness — it links different transactions to the same entity.

**How it is detected/computed:** Two detection modes:

- **Within-transaction:** Address identifiers are derived from input prevout scripts and output scripts. If any address appears in both sets, the transaction is flagged.

- **Cross-transaction (within the same block):** Before analyzing individual transactions, a block-level address map is constructed mapping every address to its occurrences across all transactions in the block. If an address appears in multiple transactions (as input in one and output in another, or across any combination), it is flagged as cross-transaction reuse.

**Confidence model:** Always high. Address reuse is a definitive, non-probabilistic signal — the same cryptographic key is being reused. There is no uncertainty in the detection itself.

**Known limitations:**
- Cross-block reuse detection would require full UTXO set tracking (out of scope; we only analyze within a single block file)
- Some legitimate use cases exist (e.g., donation addresses, mining pool payouts)
- Non-standard scripts may not produce identifiable addresses
- Dust attacks — attackers send tiny amounts to known addresses, hoping the recipient's wallet combines them with real UTXOs via CIOH

### 4. CoinJoin Detection

**What it detects:** Identifies CoinJoin transactions — privacy-enhancing constructions where multiple users combine their inputs and create equal-value outputs to obscure the transaction graph.

**How it is detected/computed:** A transaction is classified as CoinJoin when:
- It has 3+ inputs AND 3+ outputs
- There exists a group of 3+ non-OP_RETURN outputs with identical values
- Input script type diversity is tracked (multiple types suggest different wallets)

**Confidence model:**
- **high:** ≥5 equal-value outputs AND ≥2 distinct input script types, OR confirmed via subset-sum verification (no partition exists). This strongly suggests multiple wallets from different users contributing inputs.
- **medium:** 3-4 equal-value outputs, or a partition was found by subset-sum check (suggesting possible batch payment).

**Subset-sum CoinJoin verification:** After initial CoinJoin detection, a depth-limited subset-sum check attempts to partition inputs into groups matching the equal-value output amount + estimated fee share. If no partition exists, the transaction is a confirmed CoinJoin (confidence upgraded to high). If a partition is found, it suggests the transaction may be a batch payment rather than a CoinJoin (confidence capped at medium). Reference: BlockSci is_definite_coinjoin.

**Known limitations:**
- Does not detect PayJoin / Stowaway (these privacy techniques do not use equal-value outputs by design)
- May flag non-CoinJoin transactions with coincidentally equal outputs (mitigated by subset-sum verification)
- Variable-denomination CoinJoin implementations may be missed
- The 3-output threshold may miss smaller CoinJoin constructions (e.g., 2-party CoinJoin)
- Subset-sum check uses estimated fee share; exact fee attribution per CoinJoin participant is unknown

### 5. Consolidation Detection

**What it detects:** Identifies consolidation transactions where many UTXOs are combined into one or two outputs. This is common wallet maintenance to reduce UTXO set size and lower future transaction fees.

**How it is detected/computed:** A transaction is flagged as consolidation when:
- It has 5+ inputs AND at most 2 non-OP_RETURN outputs
- Output script type matching against the predominant input type is checked for stronger signal

**Confidence model:**
- **high:** ≥10 inputs AND all non-OP_RETURN outputs match the predominant input script type. This is a classic consolidation pattern.
- **medium:** 5-9 inputs, or output types don't fully match input types. The pattern is consistent with consolidation but could also be a large payment.

**Cross-heuristic interaction:** If consolidation is detected, self-transfer detection is suppressed (consolidation is the more specific pattern).

**Known limitations:**
- The 5-input threshold is conservative; smaller consolidations exist
- Exchange withdrawal patterns can resemble consolidation
- Large payments with many inputs may be misclassified

### 6. Self-Transfer Detection

**What it detects:** Identifies transactions where all inputs and outputs appear to belong to the same entity — funds moving within one's own wallet with no external payment component.

**How it is detected/computed:** A transaction is flagged as self-transfer when:
- All non-OP_RETURN outputs match the predominant input script type
- No outputs have significant round BTC amounts (≥0.001 BTC, suggesting an external payment)
- At most 3 non-OP_RETURN outputs

**Confidence model:** Always medium. Script type uniformity is suggestive but not definitive — a user could pay someone who uses the same address type.

**Known limitations:**
- False positives when paying to the same address type
- HD wallets generate fresh addresses of the same type, making this pattern common even for external payments
- Suppressed when consolidation is detected, to avoid redundant classification

### 7. Peeling Chain Detection

**What it detects:** Detects peeling chain patterns where a large UTXO is progressively "peeled" — a small payment is made and the large remainder is sent to a new change address, which is then peeled again in a subsequent transaction.

**How it is detected/computed:** A transaction matches the peeling chain pattern when:
- It has exactly 1 input and 2 outputs
- One output is significantly smaller than the other (extreme value asymmetry)

Reference: Kappos et al. (2022) — "How to Peel a Million: Validating and Expanding Bitcoin Clusters" provides the definitive treatment of peeling chains.

**Confidence model (based on the ratio of smaller/larger output):**
- **high:** ratio < 0.01 (99:1 split or more extreme) — very strong peeling signal
- **medium:** ratio 0.01-0.05 (20:1 to 99:1) — likely peeling but could be a normal payment
- **low:** ratio 0.05-0.1 (10:1 to 20:1) — possible peeling but many simple payments exhibit this pattern

**Cross-heuristic interaction:** When peeling chain is detected and change detection did not independently fire, the larger output is assigned as the likely change output.

**Known limitations:**
- Cannot detect the "chain" aspect without cross-transaction analysis. We detect individual peeling-pattern transactions, not the full multi-hop chain.
- Many simple payments also have a 1-input, 2-output structure with asymmetric values.
- Multi-hop analysis would greatly improve accuracy.

### 8. OP_RETURN Analysis

**What it detects:** Detects OP_RETURN outputs (unspendable data carriers) and classifies the embedded data by protocol. OP_RETURN outputs are identified by opcode 0x6a.

**How it is detected/computed:** Outputs with opcode 0x6a are identified as OP_RETURN. The payload data is extracted from push operations and classified by verified, unencrypted protocol prefixes:

- **Omni Layer:** 6f6d6e69 (hex encoding of "omni")
- **Open Assets:** 4f41 (hex encoding of "OA")

Reference: arXiv 2411.10325v1 documents practical colored coin detection.

**Confidence model:** Always high for detection — OP_RETURN is identified by a definitive opcode, not a probabilistic pattern. Protocol classification is limited to verified, unencrypted prefixes.

**Known limitations:**
- Only recognizes 2 protocol families with verified raw prefixes; other protocols (Counterparty, OpenTimestamps, Veriblock) use encryption or no prefix
- Does not decode protocol-specific payload contents
- Some OP_RETURN data is arbitrary text, not protocol data

### 9. Round Number Payment Detection

**What it detects:** Identifies outputs with values that are round BTC amounts (e.g., 0.001, 0.01, 0.1, 1.0 BTC). Round-number outputs are more likely to be intentional payments; non-round outputs are more likely to be change.

**How it is detected/computed:** An output value is classified as "round" by checking divisibility against standard satoshi thresholds: 100,000,000 (1 BTC), 50,000,000, 10,000,000, 5,000,000, 1,000,000, 500,000, 100,000, 50,000, 10,000 sats.

**Confidence model:**
- **high:** divisible by 0.1 BTC (10,000,000 sats) or larger — very strong signal
- **medium:** divisible by 0.001-0.01 BTC (100,000-1,000,000 sats) — common round payment amounts
- **low:** divisible by 0.0005 BTC or 0.0001 BTC — weaker signal

Reference: BlockSci change_by_power_of_ten_value; Androulaki et al. (2012).

**Known limitations:**
- False positives when change happens to be a round number
- Some payments use non-round amounts (e.g., fiat-equivalent invoices)
- The definition of "round" is somewhat subjective

## Confidence Model

### Philosophy

All heuristics are probabilistic — there is no certainty in on-chain analysis. Our confidence model is a rule-based signal accumulation system, not machine learning. Each heuristic assigns a confidence level (high, medium, low) based on the structural strength of the detected pattern.

### Cross-Heuristic Interactions

| Trigger | Effect | Justification |
|---------|--------|---------------|
| CoinJoin detected + CIOH detected | CIOH confidence → low | CoinJoin inputs are from different entities; CIOH assumption is invalid (Gong et al., 2022) |
| Peeling chain detected + change not detected | Change detection set to peeling chain's larger output | In peeling chains, the larger output is always change |
| Consolidation detected + self-transfer detected | Self-transfer suppressed | Consolidation is a more specific pattern |
| CoinJoin detected (structural) | Subset-sum verification upgrades/downgrades confidence | If no input partition exists → confirmed CoinJoin (high); if partition found → possible batch (medium) |

### When the Model Breaks Down

- **Modern Bitcoin (post-2020):** Address reuse has dropped below 10%, significantly weakening fresh-address-based heuristics (Gong et al., 2025).
- **Taproot adoption (P2TR):** Taproot obscures script types, making script-type-based change detection less effective.
- **Privacy-enhancing techniques:** CoinJoin, PayJoin, and mixing services deliberately break heuristic assumptions.
- **Protocol transactions:** OP_RETURN-based protocols should be excluded from payment heuristics.

## References

### Foundational Papers
- Nakamoto, S. (2008). "Bitcoin: A Peer-to-Peer Electronic Cash System." Section 10 introduces the common-input-ownership assumption. https://bitcoin.org/bitcoin.pdf
- Meiklejohn, S., Pomarole, M., Jordan, G., Levchenko, K., McCoy, D., Voelker, G.M., Savage, S. (2013). "A Fistful of Bitcoin: Characterizing Payments Among Men with No Names." IMC 2013. https://cseweb.ucsd.edu/~smeiklejohn/files/imc13.pdf
- Kalodner, H., Goldfeder, S., Chator, A., Möser, M., Narayanan, A. (2020). "BlockSci: Design and Applications of a Blockchain Analysis Platform." USENIX Security 2020. https://www.usenix.org/system/files/sec20-kalodner.pdf
- Kappos, G., Yousaf, H., Stütz, R., Rollet, S., Haslhofer, B., Meiklejohn, S. (2022). "How to Peel a Million: Validating and Expanding Bitcoin Clusters." USENIX Security 2022. https://www.usenix.org/system/files/sec22-kappos.pdf

### Per-Heuristic References
- Androulaki, E., Karame, G.O., Roeschlin, M., Scherer, T., Capkun, S. (2012). "Evaluating User Privacy in Bitcoin." https://eprint.iacr.org/2012/596.pdf
- Schnoering, H., Vazirgiannis, M. (2023). "Heuristics for Detecting CoinJoin Transactions on the Bitcoin Blockchain." arXiv 2311.12491.
- Gong, J., Chow, K.P., Ting, A., Yiu, S.M. (2022). "Analyzing the Error Rates of Bitcoin Clustering Heuristics." IFIP Digital Forensics 2022. https://inria.hal.science/hal-05315736v1/document
- Gong, J., Chow, K.P., Yiu, S.M. (2025). "Improved Bitcoin Simulation Model and Address Heuristic Method." Forensic Science International.
- Möser, M., Narayanan, A. (2021). "Resurrecting Address Clustering in Bitcoin." arXiv (Financial Cryptography 2022).
- Ghesmati, S., Fdhila, W., Weippl, E. (2021). "Unnecessary Input Heuristics and PayJoin Transactions." HCI International 2021. DOI: 10.1007/978-3-030-78642-7_56
- Ron, D., Shamir, A. (2012). "Quantitative Analysis of the Full Bitcoin Transaction Graph." https://eprint.iacr.org/2012/584.pdf

### Additional References
- BlockSci Heuristics Documentation — https://citp.github.io/old-blocksci-docs/0.4.5/heuristics/heuristics.html
- BIP34: Block v2, Height in Coinbase
- BIP141: Segregated Witness (Consensus Layer)
- BIP173: Base32 address format for native v0-16 witness outputs (Bech32)
- BIP350: Bech32m format for v1+ witness addresses



### Bitcoin Terms

```
┌─────────────────────────────────────────────────────────────────────────────────────────┐
│                              GLOSSARY                                                │
├─────────────────────────────────────────────────────────────────────────────────────────┤
│                                                                                          │
│  BLOCKCHAIN TERMS:                                                                       │
│  ══════════════════                                                                      │
│                                                                                          │
│  Block         │ A collection of verified Bitcoin transactions, ~1MB data              │
│  Transaction  │ A transfer of Bitcoin from sender(s) to recipient(s)               │
│  TXID         │ Transaction ID - double SHA256 hash of transaction                    │
│  UTXO         │ Unspent Transaction Output - spendable Bitcoin balance                │
│  Coinbase     │ First tx in block - creates new Bitcoin (miner reward)              │
│  Prevout      │ Previous output being spent as input                                 │
│  Fee          │ Input sum - Output sum (paid to miner)                               │
│  Fee Rate     │ Fee per virtual byte (sats/vB)                                      │
│  VBytes       │ Virtual bytes = weight / 4                                          │
│                                                                                          │
│  SCRIPT TYPES:                                                                           │
│  ════════════                                                                            │
│                                                                                          │
│  P2PKH   │ Pay to Public Key Hash      │ Legacy    │ Addresses: 1...              │
│  P2SH    │ Pay to Script Hash          │ Legacy    │ Addresses: 3...              │
│  P2WPKH  │ Pay to Witness PKH         │ SegWit    │ Addresses: bc1q...          │
│  P2WSH   │ Pay to Witness Script      │ SegWit    │ Addresses: bc1q...          │
│  P2TR    │ Pay to Taproot             │ Taproot   │ Addresses: bc1p...          │
│  OP_RET  │ OP_RETURN (data carrier)   │ -         │ No address                  │
│                                                                                          │
│  HEURISTIC TERMS:                                                                        │
│  ════════════════                                                                        │
│                                                                                          │
│  CIOH       │ Common Input Ownership Heuristic                                         │
│  Change    │ Output returning funds to sender                                         │
│  CoinJoin  │ Multi-party privacy transaction                                         │
│  Consolid  │ Combining many UTXOs into fewer                                         │
│  Peeling   │ Progressive splitting: small payment + large change                    │
│  Cluster   │ Group of addresses believed same owner                                  │
│                                                                                          │
│  ANALYSIS TERMS:                                                                        │
│  ══════════════                                                                         │
│                                                                                          │
│  False Positive │ Heuristic incorrectly flags something                              │
│  False Negative │ Heuristic misses something actual                                   │
│  Confidence    │ Reliability level of detection                                      │
│  Classification│ Transaction type (simple_payment, coinjoin, etc.)                   │
│                                                                                          │
└─────────────────────────────────────────────────────────────────────────────────────────┘
```

---

## Performance Optimizations

### Zero-Copy Binary Parser

```
┌─────────────────────────────────────────────────────────────────────────────────────────┐
│                          PERFORMANCE OPTIMIZATIONS                                      │
├─────────────────────────────────────────────────────────────────────────────────────────┤
│                                                                                          │
│  PROBLEM:                                                                                │
│  ════════                                                                                │
│  Standard Bitcoin libraries allocate heavily for each transaction, creating thousands    │
│  of objects per block. Processing block data initially took significant time, making     │
│  CI/CD and rapid local testing difficult.                                                │
│                                                                                          │
│  SOLUTION: CUSTOM BINARY PARSER                                                          │
│  ══════════════════════════════                                                          │
│                                                                                          │
│  1. Zero-Copy Operations (`internal/parser/reader.go`):                                  │
│     • Uses byte slices (`data[start:end]`) instead of copying data                       │
│     • Reads integers (VarInt, Uint32LE) directly from memory                             │
│     • Avoids generic heap allocations for temporary buffers                              │
│                                                                                          │
│  2. Direct Transaction Parsing (`internal/parser/tx.go`):                                │
│     • Single-pass extraction of all necessary fields (inputs, outputs, scripts)          │
│     • Extracts SegWit markers efficiently                                                │
│     • Computes legacy serialization simultaneously for TXID generation                   │
│     • Eliminates the need to reconstruct bytes for hashing                               │
│                                                                                          │
│  RESULTS:                                                                                │
│  ════════                                                                                │
│  • Time to process dataset dramatically reduced (under 10 minutes requirement met)       │
│  • Highly efficient memory usage suitable for scale                                      │
│  • CI/CD pipeline compatibility achieved successfully                                    │
│                                                                                          │
└─────────────────────────────────────────────────────────────────────────────────────────┘
```

---

## Trade-offs

### Design Decisions Summary

```
┌─────────────────────────────────────────────────────────────────────────────────────────┐
│                          TRADE-OFFS AND DESIGN DECISIONS                              │
├─────────────────────────────────────────────────────────────────────────────────────────┤
│                                                                                          │
│  ACCURACY vs PERFORMANCE:                                                                │
│  ════════════════════════                                                               │
│                                                                                          │
│  Decision                          │ Impact                                            │
│  ──────────────────────────────────┼───────────────────────────────────                  │
│  Block-level address map          │ Enables cross-tx reuse detection                 │
│                                   │ O(n) memory per block, acceptable                │
│                                                                                       │
│  Transaction array optimization   │ Only first block has full tx data               │
│                                   │ Reduces JSON size by ~90%                        │
│                                                                                       │
│  Conservative thresholds         │ CoinJoin: 3+ outputs (not 2)                    │
│                                   │ Consolidation: 5+ inputs (not 3)                  │
│                                   │ Reduces false positives, some false negatives     │
│                                                                                       │
│  Subset-sum verification limit   │ 10,000 iterations max                          │
│                                   │ Prevents performance issues on large txs          │
│                                                                                       │
│  SIMPLICITY vs COVERAGE:                                                                   │
│  ════════════════════════                                                                   │
│                                                                                          │
│  Decision                          │ Impact                                            │
│  ──────────────────────────────────┼───────────────────────────────────                  │
│  All 9 heuristics                 │ Comprehensive coverage (vs minimum 5)             │
│                                   │ More complete analysis                           │
│                                                                                       │
│  Independent heuristic design   │ Each testable in isolation                       │
│                                   │ Easier debugging                                │
│                                                                                       │
│  Deferred multi-tx analysis    │ Simpler initial implementation                   │
│                                   │ (full peeling chain tracking deferred)          │
│                                                                                       │
│  DESIGN DECISIONS:                                                                       │
│  ══════════════════                                                                       │
│                                                                                          │
│  Decision                          │ Rationale                                         │
│  ──────────────────────────────────┼───────────────────────────────────                  │
│  Confidence as string             │ More interpretable than numeric scores            │
│  ("high/medium/low")             │ Matches JSON schema example                      │
│                                   │ Avoids false precision                           │
│                                                                                       │
│  OP_RETURN protocol detection    │ Only verified prefixes (Omni, OpenAssets)       │
│                                   │ Avoids false positives                          │
│                                                                                       │
│  0.001 BTC round threshold       │ Below = dust/fees, not payments               │
│                                   │ Avoids false positives                          │
│                                                                                       │
│  Block-level freshness only      │ Full-chain requires UTXO set                    │
│                                   │ Not available in block-file analysis             │
│                                                                                       │
└─────────────────────────────────────────────────────────────────────────────────────────┘
```

---

*Document generated for Sherlock Bitcoin Chain Analysis Challenge*

*Total size: ~60KB covering Architecture, Data Flow, Heuristics, Confidence Model, Trade-offs, and References*
