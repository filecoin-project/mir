package lookuptxs

import (
	"github.com/filecoin-project/mir/pkg/dsl"
	mpdsl "github.com/filecoin-project/mir/pkg/mempool/dsl"
	"github.com/filecoin-project/mir/pkg/mempool/simplemempool/internal/common"
	mppb "github.com/filecoin-project/mir/pkg/pb/mempoolpb"
	t "github.com/filecoin-project/mir/pkg/types"
)

// IncludeTransactionLookupByID registers event handlers for transaction looking up transactions in the mempool by
// their IDs.
func IncludeTransactionLookupByID(
	m dsl.Module,
	mc *common.ModuleConfig,
	params *common.ModuleParams,
	commonState *common.State,
) {
	mpdsl.UponRequestTransactions(m, func(txIDs []t.TxID, origin *mppb.RequestTransactionsOrigin) error {
		present := make([]bool, len(txIDs))
		txs := make([][]byte, len(txIDs))
		for i, txID := range txIDs {
			txs[i], present[i] = commonState.TxByID[txID]
		}

		mpdsl.TransactionsResponse(m, t.ModuleID(origin.Module), present, txs, origin)
		return nil
	})
}
