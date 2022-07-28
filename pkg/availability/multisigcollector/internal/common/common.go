package common

import (
	t "github.com/filecoin-project/mir/pkg/types"
)

// InstanceUID is used to uniquely identify an instance of multisig collector.
// It is used to prevent cross-instance signature replay attack and should be unique across all executions.
type InstanceUID []byte

// Bytes returns the binary representation of the InstanceUID.
func (uid InstanceUID) Bytes() []byte {
	return uid
}

// ModuleConfig sets the module ids. All replicas are expected to use identical module configurations.
type ModuleConfig struct {
	Self    t.ModuleID // id of this module
	Mempool t.ModuleID
	Net     t.ModuleID
	Crypto  t.ModuleID
}

// DefaultModuleConfig returns a valid module config with default names for all modules.
func DefaultModuleConfig() *ModuleConfig {
	return &ModuleConfig{
		Self:    "availability",
		Mempool: "mempool",
		Net:     "net",
		Crypto:  "crypto",
	}
}

// ModuleParams sets the values for the parameters of an instance of the protocol.
// All replicas are expected to use identical module parameters.
type ModuleParams struct {
	InstanceUID []byte     // unique identifier for this instance of BCB, used to prevent cross-instance replay attacks
	AllNodes    []t.NodeID // the list of participating nodes
}

// N is the total number of replicas.
func (params *ModuleParams) N() int {
	return len(params.AllNodes)
}

// F is the maximum number of replicas that can be tolerated.
func (params *ModuleParams) F() int {
	return (params.N() - 1) / 3
}

// State represents the common state accessible to all parts of the multisig collector implementation.
type State struct {
	BatchStore       map[t.BatchID][]t.TxID
	TransactionStore map[t.TxID][]byte
}

// SigData is the binary data that should be signed for forming a certificate.
func SigData(instanceUID InstanceUID, batchID t.BatchID) [][]byte {
	return [][]byte{instanceUID.Bytes(), []byte("BATCH_STORED"), batchID.Bytes()}
}
