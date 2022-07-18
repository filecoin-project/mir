package deploytest

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"sync"

	"github.com/filecoin-project/mir/pkg/net"

	"github.com/filecoin-project/mir"
	mirCrypto "github.com/filecoin-project/mir/pkg/crypto"
	"github.com/filecoin-project/mir/pkg/eventlog"
	"github.com/filecoin-project/mir/pkg/events"
	"github.com/filecoin-project/mir/pkg/iss"
	"github.com/filecoin-project/mir/pkg/logging"
	"github.com/filecoin-project/mir/pkg/modules"
	"github.com/filecoin-project/mir/pkg/pb/requestpb"
	"github.com/filecoin-project/mir/pkg/requestreceiver"
	"github.com/filecoin-project/mir/pkg/simplewal"
	t "github.com/filecoin-project/mir/pkg/types"
)

// TestReplica represents one replica (that uses one instance of the mir.Node) in the test system.
type TestReplica struct {

	// ID of the replica as seen by the protocol.
	ID t.NodeID

	// Dummy test application the replica is running.
	App *FakeApp

	// Name of the directory where the persisted state of this TestReplica will be stored,
	// along with the logs produced by running the replica.
	Dir string

	// Configuration of the node corresponding to this replica.
	Config *mir.NodeConfig

	// List of replica IDs constituting the (static) membership.
	Membership []t.NodeID

	// Network transport subsystem.
	Transport net.Transport

	// Number of simulated requests inserted in the test replica by a hypothetical client.
	NumFakeRequests int

	// Configuration of the ISS protocol, if used. If set to nil, the default ISS configuration is assumed.
	ISSConfig *iss.Config
}

// EventLogFile returns the name of the file where the replica's event log is stored.
func (tr *TestReplica) EventLogFile() string {
	return filepath.Join(tr.Dir, "eventlog.gz")
}

// Run initializes all the required modules and starts the test replica.
// The function blocks until the replica stops.
// The replica stops when stopC is closed.
// Run returns the error returned by the run of the underlying Mir node.
func (tr *TestReplica) Run(ctx context.Context) error {

	// Initialize the write-ahead log.
	walPath := filepath.Join(tr.Dir, "wal")
	if err := os.MkdirAll(walPath, 0700); err != nil {
		return fmt.Errorf("error creating WAL directory: %w", err)
	}
	wal, err := simplewal.Open(walPath)
	if err != nil {
		return fmt.Errorf("error opening WAL: %w", err)
	}
	defer wal.Close()

	// Initialize recording of events.
	file, err := os.Create(tr.EventLogFile())
	if err != nil {
		return fmt.Errorf("error creating event log file: %w", err)
	}
	defer file.Close()
	interceptor := eventlog.NewRecorder(tr.ID, file, logging.Decorate(tr.Config.Logger, "Interceptor: "))
	defer func() {
		if err := interceptor.Stop(); err != nil {
			panic(err)
		}
	}()

	// If no ISS Protocol configuration has been specified, use the default one.
	if tr.ISSConfig == nil {
		tr.ISSConfig = iss.DefaultConfig(tr.Membership)
	}

	issProtocol, err := iss.New(tr.ID, tr.ISSConfig, logging.Decorate(tr.Config.Logger, "ISS: "))
	if err != nil {
		return fmt.Errorf("error creating ISS protocol module: %w", err)
	}

	cryptoModule, err := mirCrypto.NodePseudo(tr.Membership, tr.ID, mirCrypto.DefaultPseudoSeed)
	if err != nil {
		return fmt.Errorf("error creating crypto module: %w", err)
	}

	// Create the mir node for this replica.
	modulesWithDefaults, err := iss.DefaultModules(map[t.ModuleID]modules.Module{
		"app":    tr.App,
		"crypto": mirCrypto.New(cryptoModule),
		"wal":    wal,
		"iss":    issProtocol,
		"net":    tr.Transport,
	})
	if err != nil {
		return fmt.Errorf("error initializing the Mir modules: %w", err)
	}

	node, err := mir.NewNode(
		tr.ID,
		tr.Config,
		modulesWithDefaults,
		wal,
		interceptor,
	)
	if err != nil {
		return fmt.Errorf("error creating Mir node: %w", err)
	}

	// Create a RequestReceiver for request coming over the network.
	requestReceiver := requestreceiver.NewRequestReceiver(node, "iss", logging.Decorate(tr.Config.Logger, "ReqRec: "))
	p, err := strconv.Atoi(tr.ID.Pb())
	if err != nil {
		return fmt.Errorf("error converting node ID %s: %w", tr.ID, err)
	}
	err = requestReceiver.Start(RequestListenPort + p)
	if err != nil {
		return fmt.Errorf("error starting request receiver: %w", err)
	}

	// Initialize WaitGroup for the replica's request submission thread.
	var wg sync.WaitGroup
	wg.Add(1)

	// Start thread submitting requests from a (single) hypothetical client.
	// The client submits a predefined number of requests and then stops.
	go tr.submitFakeRequests(ctx, node, &wg)

	err = tr.Transport.Start()
	if err != nil {
		return fmt.Errorf("error starting the network link: %w", err)
	}
	tr.Transport.Connect(ctx)
	defer tr.Transport.Stop()

	// Run the node until it stops.
	exitErr := node.Run(ctx)
	tr.Config.Logger.Log(logging.LevelDebug, "Node run returned!")

	// Stop the request receiver.
	requestReceiver.Stop()
	if err := requestReceiver.ServerError(); err != nil {
		return fmt.Errorf("request receiver returned server error: %w", err)
	}

	// Wait for the local request submission thread.
	wg.Wait()
	tr.Config.Logger.Log(logging.LevelInfo, "Fake request submission done.")

	return exitErr
}

// Submits n fake requests to node.
// Aborts when stopC is closed.
// Decrements wg when done.
func (tr *TestReplica) submitFakeRequests(ctx context.Context, node *mir.Node, wg *sync.WaitGroup) {
	defer wg.Done()

	// The ID of the fake client is always 0.
	for i := 0; i < tr.NumFakeRequests; i++ {
		select {
		case <-ctx.Done():
			// Stop submitting if shutting down.
			break
		default:
			// Otherwise, submit next request.

			if err := node.InjectEvents(ctx, events.ListOf(events.NewClientRequests(
				"iss",
				[]*requestpb.Request{events.ClientRequest(
					t.NewClientIDFromInt(0),
					t.ReqNo(i),
					[]byte(fmt.Sprintf("Request %d", i)),
				)},
			))); err != nil {

				// TODO (Jason), failing on err causes flakes in the teardown,
				// so just returning for now, we should address later
				break
			}
			// TODO: Add some configurable delay here
		}
	}
}
