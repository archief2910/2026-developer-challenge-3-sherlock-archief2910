export const HEURISTIC_DEFINITIONS = {
  address_reuse: {
    name: "Address Reuse",
    definition: "When the same Bitcoin address is used multiple times to receive payments. This reduces privacy as anyone can trace all transactions linked to that address.",
    whyImportant: "Address reuse is considered a privacy risk because it links multiple transactions together, making it easier to track a user's spending patterns."
  },
  change_detection: {
    name: "Change Detection",
    definition: "Identifying the 'change' output in a transaction - the portion of Bitcoin returned to the sender as change from a payment.",
    whyImportant: "Analyzing change outputs helps understand which outputs are payments to others versus returning funds to the sender's own wallet."
  },
  cioh: {
    name: "Common Input Ownership (CIOH)",
    definition: "The assumption that all inputs in a transaction belong to the same wallet/entity. When multiple addresses are used to fund a single transaction, they likely belong to the same person.",
    whyImportant: "This is the foundational heuristic for clustering addresses together to identify wallet balances and spending behavior."
  },
  coinjoin: {
    name: "CoinJoin",
    definition: "A privacy technique where multiple users combine their transactions into a single transaction, making it difficult to trace which output belongs to whom.",
    whyImportant: "CoinJoin is used to enhance privacy, but our analyzer can detect when this technique was used in a transaction."
  },
  consolidation: {
    name: "Consolidation",
    definition: "When many small inputs are combined into fewer outputs, typically done to reduce future transaction fees or prepare for a large transfer.",
    whyImportant: "Consolidation transactions often indicate the user is moving large amounts or preparing to sell Bitcoin."
  },
  self_transfer: {
    name: "Self-Transfer",
    definition: "A transaction where Bitcoin is sent from one of your own addresses to another address you control.",
    whyImportant: "Self-transfers are commonly used to organize wallets, move funds to new addresses, or activate dormant funds."
  },
  peeling_chain: {
    name: "Peeling Chain",
    definition: "A pattern where a large amount is split - a small portion goes to an external address while the remainder stays with the original owner in a new address.",
    whyImportant: "This pattern is often used to slowly spend from large holdings while keeping most of the funds accessible."
  },
  op_return: {
    name: "OP_RETURN",
    definition: "A special Bitcoin output used to store small amounts of data on the blockchain, often used for proofs, timestamps, or messages.",
    whyImportant: "OP_RETURN outputs can contain messages, document hashes, or metadata - our analyzer counts how many transactions include data storage."
  },
  round_number_payment: {
    name: "Round Number Payment",
    definition: "When a payment amount is a round number (like 0.1 BTC or 1.0 BTC) rather than a random amount.",
    whyImportant: "Round numbers are often manually chosen by humans and can indicate specific payment amounts or psychological pricing."
  }
};

export const CLASSIFICATION_DEFINITIONS = {
  simple_payment: {
    name: "Simple Payment",
    definition: "A standard Bitcoin payment from one person to another - typically 1 input going to 1-2 outputs.",
    example: "Alice sends 0.5 BTC to Bob"
  },
  coinjoin: {
    name: "CoinJoin",
    definition: "A collaborative transaction where multiple users merge their payments to enhance privacy.",
    example: "Multiple people pool their transactions together"
  },
  consolidation: {
    name: "Consolidation",
    definition: "Combining many small amounts of Bitcoin into fewer, larger amounts.",
    example: "Sweeping 50 small UTXOs into 1 large one"
  },
  self_transfer: {
    name: "Self-Transfer",
    definition: "Moving Bitcoin from one of your own addresses to another address you control.",
    example: "Moving funds from your savings to spending wallet"
  },
  batch_payment: {
    name: "Batch Payment",
    definition: "A single transaction that pays multiple recipients at once, common with exchanges or services.",
    example: "An exchange processing withdrawal requests to multiple users"
  },
  unknown: {
    name: "Unknown",
    definition: "A transaction that doesn't match any of the other classification patterns.",
    example: "Complex transactions that don't fit standard patterns"
  }
};

export const SCRIPT_TYPE_DEFINITIONS = {
  p2pkh: {
    name: "Pay to Public Key Hash (P2PKH)",
    definition: "The original Bitcoin script type. Payments are sent to a hash of a public key. Legacy format, now rarely used for new wallets.",
    example: "Old-style Bitcoin addresses starting with '1'"
  },
  p2sh: {
    name: "Pay to Script Hash (P2SH)",
    definition: "A more flexible script type where payments are sent to the hash of a script. Often used for multi-signature wallets.",
    example: "Addresses starting with '3'"
  },
  p2wpkh: {
    name: "Pay to Witness Public Key Hash (P2WPKH)",
    definition: "SegWit version of P2PKH. Offers lower transaction fees and increased privacy. Modern standard for most wallets.",
    example: "Native SegWit addresses starting with 'bc1q'"
  },
  p2wsh: {
    name: "Pay to Witness Script Hash (P2WSH)",
    definition: "SegWit version of P2SH. Used for more complex scripts like multi-signature setups with lower fees.",
    example: "Native SegWit script hashes starting with 'bc1q'"
  },
  p2tr: {
    name: "Pay to Taproot (P2TR)",
    definition: "The newest Bitcoin script type. Enables smart contracts and the most privacy and efficiency improvements.",
    example: "Taproot addresses starting with 'bc1p'"
  },
  op_return: {
    name: "OP_RETURN",
    definition: "A special output type that embeds custom data into the blockchain. Used for proofs, timestamps, or small data storage.",
    example: "Storing a document hash or short message"
  },
  unknown: {
    name: "Unknown",
    definition: "A script type that our analyzer couldn't identify or classify.",
    example: "Non-standard or experimental scripts"
  }
};

export const OVERVIEW_DEFINITIONS = {
  whatIsBitcoin: {
    title: "What is Bitcoin?",
    content: "Bitcoin is a digital currency that operates without a central authority. Every transaction is recorded on a public ledger called the blockchain - like a giant, transparent notebook that everyone can see but no one can erase."
  },
  whatIsBlockchain: {
    title: "What is the Blockchain?",
    content: "The blockchain is a chain of 'blocks' containing transaction data. Each block contains many transactions, and they're linked together chronologically. Once a block is added, its transactions become part of the permanent public record."
  },
  whatIsBlock: {
    title: "What is a Block?",
    content: "A block is like a page in a ledger. It contains a batch of Bitcoin transactions that have been verified and recorded. Blocks are 'mined' by powerful computers and added to the blockchain approximately every 10 minutes."
  },
  whatIsTransaction: {
    title: "What is a Transaction?",
    content: "A Bitcoin transaction is a transfer of Bitcoin from one person to another. It includes inputs (where the Bitcoin came from) and outputs (where it's going). Each transaction has a small fee paid to miners for processing it."
  },
  whatIsFee: {
    title: "What are Transaction Fees?",
    content: "Transaction fees are small payments to miners for processing your transaction. Fees are measured in 'sat/vB' (satoshis per virtual byte) - the more data your transaction uses, the higher the fee. Higher fees mean faster confirmation."
  },
  whatIsHeuristic: {
    title: "What are Heuristics?",
    content: "Heuristics are analytical techniques used to detect patterns in blockchain data. By examining transaction patterns, we can make educated guesses about what type of activity occurred - like whether someone is consolidating funds or making a payment."
  },
  dashboardSummary: {
    title: "Dashboard Overview",
    content: "This dashboard shows analysis of blockchain data files. You can see statistics about transactions, fee rates, and detected patterns. Click on individual blocks to see detailed transaction-level forensics."
  }
};

export const TOOLTIP_STYLE = {
  background: "rgba(15, 23, 42, 0.95)",
  border: "1px solid var(--accent)",
  borderRadius: "12px",
  padding: "16px",
  maxWidth: "400px",
  boxShadow: "0 8px 32px rgba(0, 0, 0, 0.4)"
};
