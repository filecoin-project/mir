package computeids

import (
	"github.com/filecoin-project/mir/pkg/dsl"
	mpdsl "github.com/filecoin-project/mir/pkg/mempool/dsl"
	"github.com/filecoin-project/mir/pkg/mempool/simplemempool/internal/common"
	mppb "github.com/filecoin-project/mir/pkg/pb/mempoolpb"
	t "github.com/filecoin-project/mir/pkg/types"
)

//
func IncludeComputationOfTransactionAndBatchIDs(
	m dsl.Module,
	mc *common.ModuleConfig,
	params *common.ModuleParams,
	commonState *common.State,
) {
	mpdsl.UponRequestTransactionIDs(m, func(txs [][]byte, origin *mppb.RequestTransactionIDsOrigin) error {
		txMsgs := make([][][]byte, len(txs))
		for i, tx := range txs {
			txMsgs[i] = [][]byte{tx}
		}

		dsl.HashRequest(m, mc.Hasher, txMsgs, &computeHashForTransactionIDsContext{origin})
		return nil
	})

	dsl.UponHashResult(m, func(hashes [][]byte, context *computeHashForTransactionIDsContext) error {
		txIDs := make([]t.TxID, len(hashes))
		for i, hash := range hashes {
			txIDs[i] = t.TxID(hash)
		}

		mpdsl.TransactionIDsResponse(m, t.ModuleID(context.origin.Module), txIDs, context.origin)
		return nil
	})

	mpdsl.UponRequestBatchID(m, func(txIDs []t.TxID, origin *mppb.RequestBatchIDOrigin) error {
		data := make([][]byte, len(txIDs))
		for i, txID := range txIDs {
			data[i] = txID.Bytes()
		}

		dsl.HashOneMessage(m, mc.Hasher, data, &computeHashForBatchIDContext{origin})
		return nil
	})

	dsl.UponOneHashResult(m, func(hash []byte, context *computeHashForBatchIDContext) error {
		mpdsl.BatchIDResponse(m, t.ModuleID(context.origin.Module), t.BatchID(hash), context.origin)
		return nil
	})
}

////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////
// Context data structures                                                                                            //
////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////

type computeHashForTransactionIDsContext struct {
	origin *mppb.RequestTransactionIDsOrigin
}

type computeHashForBatchIDContext struct {
	origin *mppb.RequestBatchIDOrigin
}
