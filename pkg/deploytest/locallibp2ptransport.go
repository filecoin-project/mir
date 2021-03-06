package deploytest

import (
	"fmt"

	"github.com/libp2p/go-libp2p-core/host"

	"github.com/filecoin-project/mir/pkg/logging"
	"github.com/filecoin-project/mir/pkg/net"
	"github.com/filecoin-project/mir/pkg/net/libp2p"
	t "github.com/filecoin-project/mir/pkg/types"
	libp2ptools "github.com/filecoin-project/mir/pkg/util/libp2p"
)

type LocalLibp2pTransport struct {
	// Complete static membership of the system.
	// Maps the node ID of each node in the system to its libp2p address.
	membership map[t.NodeID]t.NodeAddress

	// Maps node ids to their libp2p host.
	hosts map[t.NodeID]host.Host

	// Logger is used for all logging events of this LocalGrpcTransport
	logger logging.Logger
}

func NewLocalLibp2pTransport(nodeIDs []t.NodeID, logger logging.Logger) *LocalLibp2pTransport {
	lt := &LocalLibp2pTransport{
		membership: make(map[t.NodeID]t.NodeAddress, len(nodeIDs)),
		hosts:      make(map[t.NodeID]host.Host),
		logger:     logger,
	}

	for i, id := range nodeIDs {
		lt.hosts[id] = libp2ptools.NewDummyHost(i, BaseListenPort)
		lt.membership[id] = t.NodeAddress(libp2ptools.NewDummyMultiaddr(i, BaseListenPort))
	}

	return lt
}

func (t *LocalLibp2pTransport) Link(sourceID t.NodeID) (net.Transport, error) {
	if _, ok := t.hosts[sourceID]; !ok {
		panic(fmt.Errorf("unexpected node id: %v", sourceID))
	}

	return libp2p.NewTransport(
		t.hosts[sourceID],
		sourceID,
		t.logger,
	)
}

func (t *LocalLibp2pTransport) Nodes() map[t.NodeID]t.NodeAddress {
	return t.membership
}
