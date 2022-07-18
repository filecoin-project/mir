package common

import t "github.com/filecoin-project/mir/pkg/types"

// ModuleConfig sets the module ids. All replicas are expected to use identical module configurations.
type ModuleConfig struct {
	Self   t.ModuleID // id of this module
	Hasher t.ModuleID
	Crypto t.ModuleID
}

// ModuleParams sets the values for the parameters of an instance of the protocol.
// All replicas are expected to use identical module parameters.
type ModuleParams struct {
	MaxBatchSizeInBytes    int
	MaxTransactionsInBatch int
}

// State represents the common state accessible to all parts of the multisig collector implementation.
type State struct {
	TxByID map[t.TxID][]byte
}
