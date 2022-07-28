package formbatches

import (
	"fmt"

	"google.golang.org/protobuf/proto"

	"github.com/filecoin-project/mir/pkg/dsl"
	mpdsl "github.com/filecoin-project/mir/pkg/mempool/dsl"
	"github.com/filecoin-project/mir/pkg/mempool/simplemempool/internal/common"
	mppb "github.com/filecoin-project/mir/pkg/pb/mempoolpb"
	"github.com/filecoin-project/mir/pkg/pb/requestpb"
	t "github.com/filecoin-project/mir/pkg/types"
)

type State struct {
	*common.State
	NewTxIDs []t.TxID
}

// IncludeBatchCreation registers event handlers for processing new transactions and forming batches.
func IncludeBatchCreation(
	m dsl.Module,
	mc *common.ModuleConfig,
	params *common.ModuleParams,
	commonState *common.State,
) {
	state := &State{
		State:    commonState,
		NewTxIDs: nil,
	}

	dsl.UponNewRequests(m, func(requests []*requestpb.Request) error {
		txs := make([][]byte, len(requests))
		for i, req := range requests {
			var err error
			txs[i], err = proto.Marshal(req)

			if err != nil {
				// The module that sent the NewRequests event is responsible to
				// make sure that the requests are valid.
				return err
			}

			if len(txs[i]) > params.MaxBatchSizeInBytes {
				return fmt.Errorf("transaction is too large (%v bytes)", params.MaxBatchSizeInBytes)
			}
		}

		mpdsl.RequestTransactionIDs(m, mc.Self, txs, &requestTxIDsContext{txs})
		return nil
	})

	mpdsl.UponTransactionIDsResponse(m, func(txIDs []t.TxID, context *requestTxIDsContext) error {
		for i := range txIDs {
			state.TxByID[txIDs[i]] = context.txs[i]
		}
		state.NewTxIDs = append(state.NewTxIDs, txIDs...)
		return nil
	})

	mpdsl.UponRequestBatch(m, func(origin *mppb.RequestBatchOrigin) error {
		var txIDs []t.TxID
		var txs [][]byte
		batchSize := 0

		var i int
		var txID t.TxID

		for i, txID = range state.NewTxIDs {
			tx := state.TxByID[txID]
			if i == params.MaxTransactionsInBatch || batchSize+len(tx) > params.MaxBatchSizeInBytes {
				break
			}

			txIDs = append(txIDs, txID)
			txs = append(txs, tx)
			batchSize += len(tx)
		}

		state.NewTxIDs = state.NewTxIDs[i:]

		// Note that a batch may be empty.
		mpdsl.NewBatch(m, t.ModuleID(origin.Module), txIDs, txs, origin)
		return nil
	})
}

////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////
// Context data structures                                                                                            //
////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////

type requestTxIDsContext struct {
	txs [][]byte
}
