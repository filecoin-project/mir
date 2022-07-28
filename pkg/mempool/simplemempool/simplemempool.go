package simplemempool

import (
	"github.com/filecoin-project/mir/pkg/dsl"
	"github.com/filecoin-project/mir/pkg/mempool/simplemempool/internal/common"
	"github.com/filecoin-project/mir/pkg/mempool/simplemempool/internal/parts/computeids"
	"github.com/filecoin-project/mir/pkg/mempool/simplemempool/internal/parts/formbatches"
	"github.com/filecoin-project/mir/pkg/mempool/simplemempool/internal/parts/lookuptxs"
	"github.com/filecoin-project/mir/pkg/modules"
	t "github.com/filecoin-project/mir/pkg/types"
)

// ModuleConfig sets the module ids. All replicas are expected to use identical module configurations.
type ModuleConfig = common.ModuleConfig

// ModuleParams sets the values for the parameters of an instance of the protocol.
// All replicas are expected to use identical module parameters.
type ModuleParams = common.ModuleParams

// DefaultModuleConfig returns a valid module config with default names for all modules.
func DefaultModuleConfig() *ModuleConfig {
	return &ModuleConfig{
		Self:   "availability",
		Hasher: "hasher",
		Crypto: "crypto",
	}
}

// NewModule creates a new instance of a simple mempool module implementation. It passively waits for
// eventpb.NewRequests events and stores them in a local map.
//
// On a batch request, this implementation creates a batch that consists of as many requests received since the
// previous batch request as possible with respect to params.MaxBatchSizeInBytes and params.MaxTransactionsInBatch.
//
// This implementation uses the hash function provided by the mc.Hasher module to compute transaction IDs and batch IDs.
func NewModule(mc *ModuleConfig, params *ModuleParams) modules.Module {
	m := dsl.NewModule(mc.Self)

	commonState := &common.State{
		TxByID: make(map[t.TxID][]byte),
	}

	computeids.IncludeComputationOfTransactionAndBatchIDs(m, mc, params, commonState)
	formbatches.IncludeBatchCreation(m, mc, params, commonState)
	lookuptxs.IncludeTransactionLookupByID(m, mc, params, commonState)

	return m
}
