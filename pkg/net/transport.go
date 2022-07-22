package net

import (
	"context"

	"github.com/multiformats/go-multiaddr"

	"github.com/filecoin-project/mir/pkg/modules"
	"github.com/filecoin-project/mir/pkg/pb/messagepb"
	t "github.com/filecoin-project/mir/pkg/types"
)

type Transport interface {
	modules.ActiveModule

	Start() error
	Stop()

	Send(dest t.NodeID, msg *messagepb.Message) error
	Connect(ctx context.Context)
	UpdateConnections(ctx context.Context, membership map[t.NodeID]multiaddr.Multiaddr)
}
