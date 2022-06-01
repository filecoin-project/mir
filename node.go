/*
Copyright IBM Corp. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package mir

import (
	"context"
	"fmt"
	"sync"

	"github.com/filecoin-project/mir/pkg/logging"

	"github.com/filecoin-project/mir/pkg/events"
	"github.com/filecoin-project/mir/pkg/modules"
	"github.com/filecoin-project/mir/pkg/pb/eventpb"
	"github.com/filecoin-project/mir/pkg/pb/statuspb"
	t "github.com/filecoin-project/mir/pkg/types"
)

var ErrStopped = fmt.Errorf("stopped at caller request")

// Node is the local instance of Mir and the application's interface to the mir library.
type Node struct {
	ID     t.NodeID    // Protocol-level node ID
	Config *NodeConfig // Node-level (protocol-independent) configuration, like buffer sizes, logging, ...

	// Implementations of networking, hashing, request store, WAL, etc.
	// The state machine is also a module of the node.
	modules *modules.Modules

	// A buffer for storing outstanding events that need to be processed by the node.
	// It contains a separate sub-buffer for each type of event.
	workItems *workItems

	// Channels for routing work items between modules.
	// Whenever workItems contains events, those events will be written (by the process() method)
	// to the corresponding channel in workChans. When processed by the corresponding module,
	// the result of that processing (also a list of events) will also be written
	// to the appropriate channel in workChans, from which the process() method moves them to the workItems buffer.
	workChans workChans

	// Used to synchronize the exit of the node's worker go routines.
	workErrNotifier *workErrNotifier

	// Channel for receiving status requests.
	// A status request is itself represented as a channel,
	// to which the state machine status needs to be written once the status is obtained.
	// TODO: Implement obtaining and writing the status (Currently no one reads from this channel).
	statusC chan chan *statuspb.NodeStatus

	// If set to true, the node is in debug mode.
	// Only events received through the Step method are applied.
	// Events produced by the modules are, instead of being applied,
	debugMode bool
}

// NewNode creates a new node with numeric ID id.
// The config parameter specifies Node-level (protocol-independent) configuration, like buffer sizes, logging, ...
// The modules parameter must contain initialized, ready-to-use modules that the new Node will use.
func NewNode(
	id t.NodeID,
	config *NodeConfig,
	m *modules.Modules,
) (*Node, error) {

	// Create default modules for those not specified by the user.
	modulesWithDefaults, err := modules.Defaults(*m)
	if err != nil {
		return nil, err
	}

	// Return a new Node.
	return &Node{
		ID:     id,
		Config: config,

		workChans: newWorkChans(),
		modules:   modulesWithDefaults,

		workItems:       newWorkItems(),
		workErrNotifier: newWorkErrNotifier(),

		statusC: make(chan chan *statuspb.NodeStatus),
	}, nil
}

// Status returns a static snapshot in time of the internal state of the Node.
// TODO: Currently a call to Status blocks until the node is stopped, as obtaining status is not yet implemented.
//       Also change the return type to be a protobuf object that contains a field for each module
//       with module-specific contents.
func (n *Node) Status(ctx context.Context) (*statuspb.NodeStatus, error) {

	// Submit status request for processing by the process() function.
	// A status request is represented as a channel (statusC)
	// to which the state machine status needs to be written once the status is obtained.
	// Return an error if the node shuts down before the request is read or if the context ends.
	statusC := make(chan *statuspb.NodeStatus, 1)
	select {
	case n.statusC <- statusC:
	case <-ctx.Done():
		return nil, ctx.Err()
	case <-n.workErrNotifier.ExitStatusC():
		return n.workErrNotifier.ExitStatus()
	}

	// Read the obtained status and return it.
	// Return an error if the node shuts down before the request is read or if the context ends.
	select {
	case s := <-statusC:
		return s, nil
	case <-ctx.Done():
		return nil, ctx.Err()
	case <-n.workErrNotifier.ExitStatusC():
		return n.workErrNotifier.ExitStatus()
	}
}

// Debug runs the Node in debug mode.
// If the node has been instantiated with a WAL, its contents will be loaded.
// Then, the Node will ony process events submitted through the Step method.
// All internally generated events will be ignored
// and, if the eventsOut argument is not nil, written to eventsOut instead.
// Note that if the caller supplies such a channel, the caller is expected to read from it.
// Otherwise, the Node's execution might block while writing to the channel.
func (n *Node) Debug(ctx context.Context, eventsOut chan *events.EventList) error {
	// Enable debug mode
	n.debugMode = true

	// If a WAL implementation is available,
	// load the contents of the WAL and enqueue it for processing.
	if n.modules.WAL != nil {
		if err := n.processWAL(); err != nil {
			n.workErrNotifier.Fail(err)
			n.workErrNotifier.SetExitStatus(nil, fmt.Errorf("node not started"))
			return fmt.Errorf("could not process WAL: %w", err)
		}
	}

	// Set up channel for outputting internal events
	n.workChans.debugOut = eventsOut

	// Start processing of events.
	return n.process(ctx)

}

// Step inserts an Event in the Node.
// Useful for debugging.
func (n *Node) Step(ctx context.Context, event *eventpb.Event) error {

	// Enqueue event in a work channel to be handled by the processing thread.
	select {
	case n.workChans.debugIn <- (&events.EventList{}).PushBack(event):
		return nil
	case <-n.workErrNotifier.ExitStatusC():
		return n.workErrNotifier.Err()
	case <-ctx.Done():
		return ctx.Err()
	}
}

// SubmitRequest submits a new client request to the Node.
// clientID and reqNo uniquely identify the request.
// data constitutes the (opaque) payload of the request.
// SubmitRequest is safe to be called concurrently by multiple threads.
func (n *Node) SubmitRequest(
	ctx context.Context,
	clientID t.ClientID,
	reqNo t.ReqNo,
	data []byte,
	authenticator []byte) error {

	// Enqueue the generated events in a work channel to be handled by the processing thread.
	select {
	case n.workChans.workItemInput <- (&events.EventList{}).PushBack(
		events.ClientRequest("clientTracker", clientID, reqNo, data, authenticator),
	):
		return nil
	case <-ctx.Done():
		return ctx.Err()
	case <-n.workErrNotifier.ExitC():
		return n.workErrNotifier.Err()
	}
}

// Run starts the Node.
// First, it loads the contents of the WAL and enqueues all its contents for processing.
// This makes sure that the WAL events end up first in all the modules' processing queues.
// Then it adds an Init event to the work items, giving the modules the possibility
// to perform additional initialization based on the state recovered from the WAL.
// Run then launches the processing of incoming messages, and internal events.
// The node stops when the ctx is canceled.
// The function call is blocking and only returns when the node stops.
func (n *Node) Run(ctx context.Context) error {

	// If a WAL implementation is available,
	// load the contents of the WAL and enqueue it for processing.
	if n.modules.WAL != nil {
		if err := n.processWAL(); err != nil {
			n.workErrNotifier.Fail(err)
			n.workErrNotifier.SetExitStatus(nil, fmt.Errorf("node not started"))
			return fmt.Errorf("could not process WAL: %w", err)
		}
	}

	// Submit the Init event to the modules.
	if err := n.workItems.AddEvents((&events.EventList{}).PushBack(events.Init("iss"))); err != nil {
		n.workErrNotifier.Fail(err)
		n.workErrNotifier.SetExitStatus(nil, fmt.Errorf("node not started"))
		return fmt.Errorf("failed to add init event: %w", err)
	}

	// Start processing of events.
	return n.process(ctx)
}

// Loads all events stored in the WAL and enqueues them in the node's processing queues.
func (n *Node) processWAL() error {

	// Create empty EventList to hold all the WAL events.
	walEvents := &events.EventList{}

	// Add all events from the WAL to the new EventList.
	if err := n.modules.WAL.LoadAll(func(retIdx t.WALRetIndex, event *eventpb.Event) {
		walEvents.PushBack(events.WALEntry(event.Destination, event, retIdx))
	}); err != nil {
		return fmt.Errorf("could not load WAL events: %w", err)
	}

	// Enqueue all events to the workItems buffers.
	if err := n.workItems.AddEvents(walEvents); err != nil {
		return fmt.Errorf("could not enqueue WAL events for processing: %w", err)
	}

	// If we made it all the way here, no error occurred.
	return nil
}

// Performs all internal work of the node,
// which mostly consists of routing events between the node's modules.
// Stops and returns when ctx is canceled.
func (n *Node) process(ctx context.Context) error { //nolint:gocyclo

	var wg sync.WaitGroup // Synchronizes all the worker functions
	defer wg.Wait()       // Watch out! If process() terminates unexpectedly (e.g. by panicking), this might get stuck!

	// Start all worker functions in separate threads.
	// Those functions mostly read events from their respective channels in n.workChans,
	// process them correspondingly, and write the results (also represented as events) in the appropriate channels.
	// Each workFunc reads a single work item, processes it and writes its results.
	// The looping behavior is implemented in doUntilErr.
	for _, work := range []workFunc{
		n.doWALWork,
		n.doClientWork,
		n.doHashWork, // TODO (Jason), spawn more of these
		n.doSendingWork,
		n.doAppWork,
		n.doReqStoreWork,
		n.doProtocolWork,
		n.doCryptoWork,
		n.doTimerWork,
	} {
		// Each function is executed by a separate thread.
		// The wg is waited on before n.process() returns.
		wg.Add(1)
		go func(work workFunc) {
			defer wg.Done()
			n.doUntilErr(ctx, work)
		}(work)
	}

	// This map holds the respective channels into which events consumed by the respective modules are written.
	// There is one channel per module that is set to nil by default, making any writes to it block,
	// preventing them from being selected in the big select statement below.
	// When there are outstanding events to be written in a workChan,
	// the workChan is saved in the corresponding channel variable, making it available to the select statement.
	// When writing to the workChan (saved in one of these variables) is selected and the events written,
	// the variable is set to nil again until new events are ready.
	// This complicated construction is necessary to prevent writing empty event lists to the workChans
	// when no events are pending.
	moduleInputs := make(map[string]chan<- *events.EventList)
	for moduleName := range n.modules.ActiveModules {
		if _, ok := moduleInputs[moduleName]; ok {
			n.workErrNotifier.Fail(fmt.Errorf("duplicate module name: %s", moduleName))
		}
	}
	var (
		walEvents,
		clientEvents,
		hashEvents,
		cryptoEvents,
		timerEvents,
		netEvents,
		appEvents,
		reqStoreEvents,
		protocolEvents chan<- *events.EventList
	)

	// This loop shovels events between the appropriate channels, until a stopping condition is satisfied.
	for {

		// If any events are pending in the workItems buffers,
		// update the corresponding channel variables accordingly.
		// This needs to happen before the select statement that dispatches the work,
		// since, if there are any work items in the buffers before the first iteration,
		// they must be made available to the dispatcher. Otherwise it might get stuck.

		if protocolEvents == nil && n.workItems.Protocol().Len() > 0 {
			protocolEvents = n.workChans.protocol
		}
		if walEvents == nil && n.workItems.WAL().Len() > 0 {
			walEvents = n.workChans.wal
		}
		if clientEvents == nil && n.workItems.Client().Len() > 0 {
			clientEvents = n.workChans.clients
		}
		if hashEvents == nil && n.workItems.Hash().Len() > 0 {
			hashEvents = n.workChans.hash
		}
		if cryptoEvents == nil && n.workItems.Crypto().Len() > 0 {
			cryptoEvents = n.workChans.crypto
		}
		if timerEvents == nil && n.workItems.Timer().Len() > 0 {
			timerEvents = n.workChans.timer
		}
		if netEvents == nil && n.workItems.Net().Len() > 0 {
			netEvents = n.workChans.net
		}
		if appEvents == nil && n.workItems.App().Len() > 0 {
			appEvents = n.workChans.app
		}
		if reqStoreEvents == nil && n.workItems.ReqStore().Len() > 0 {
			reqStoreEvents = n.workChans.reqStore
		}

		// Wait until any events are ready and write them to the appropriate location.
		select {
		case <-ctx.Done():
			n.workErrNotifier.Fail(ErrStopped)

		// Write pending events to module input channels.
		// This is also the only place events are (potentially) intercepted.
		// Since only a single goroutine executes this loop, the exact sequence of the intercepted events
		// can be replayed later.

		case protocolEvents <- n.workItems.Protocol():
			n.workItems.ClearProtocol()
			protocolEvents = nil
		case walEvents <- n.workItems.WAL():
			n.workItems.ClearWAL()
			walEvents = nil
		case clientEvents <- n.workItems.Client():
			n.workItems.ClearClient()
			clientEvents = nil
		case hashEvents <- n.workItems.Hash():
			n.workItems.ClearHash()
			hashEvents = nil
		case cryptoEvents <- n.workItems.Crypto():
			n.workItems.ClearCrypto()
			cryptoEvents = nil
		case timerEvents <- n.workItems.Timer():
			n.workItems.ClearTimer()
			timerEvents = nil
		case netEvents <- n.workItems.Net():
			n.workItems.ClearNet()
			netEvents = nil
		case appEvents <- n.workItems.App():
			n.workItems.ClearApp()
			appEvents = nil
		case reqStoreEvents <- n.workItems.ReqStore():
			n.workItems.ClearReqStore()
			reqStoreEvents = nil

		// Handle messages received over the network, as obtained by the Net module.

		case receivedMessage := <-n.modules.Net.ReceiveChan():
			if n.debugMode {
				n.Config.Logger.Log(logging.LevelWarn, "Ignoring incoming message in debug mode.",
					"msg", receivedMessage)
			} else if err := n.workItems.AddEvents((&events.EventList{}).
				PushBack(events.MessageReceived("iss", receivedMessage.Sender, receivedMessage.Msg))); err != nil {
				n.workErrNotifier.Fail(err)
			}

		// Add events produced by modules and debugger to the workItems buffers and handle logical time.

		case newEvents := <-n.workChans.workItemInput:
			if n.debugMode {
				if n.workChans.debugOut != nil {
					n.workChans.debugOut <- newEvents
				}
			} else if err := n.workItems.AddEvents(newEvents); err != nil {
				n.workErrNotifier.Fail(err)
			}
		case newEvents := <-n.workChans.debugIn:
			if !n.debugMode {
				n.Config.Logger.Log(logging.LevelWarn, "Received events through debug interface but not in debug mode.",
					"numEvents", newEvents.Len())
			}
			if err := n.workItems.AddEvents(newEvents); err != nil {
				n.workErrNotifier.Fail(err)
			}

		// Handle termination of the node.

		case <-n.workErrNotifier.ExitC():
			return n.workErrNotifier.Err()
		}
	}
}

// If the interceptor module is present, passes events to it. Otherwise, does nothing.
// If an error occurs passing events to the interceptor, notifies the node by means of the workErrorNotifier.
// Note: The passed Events should be free of any follow-up Events,
// as those will be intercepted separately when processed.
// Make sure to call the Strip method of the EventList before passing it to interceptEvents.
func (n *Node) interceptEvents(events *events.EventList) {
	if n.modules.Interceptor != nil {
		if err := n.modules.Interceptor.Intercept(events); err != nil {
			n.workErrNotifier.Fail(err)
		}
	}
}
