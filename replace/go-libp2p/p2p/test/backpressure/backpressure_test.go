package backpressure_tests

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	logging "github.com/ipfs/go-log/v2"
	"github.com/libp2p/go-libp2p-core/network"
	"github.com/libp2p/go-libp2p-core/protocol"
	swarmt "github.com/libp2p/go-libp2p-swarm/testing"
	bhost "github.com/libp2p/go-libp2p/p2p/host/basic"
)

var log = logging.Logger("backpressure")

// TestStBackpressureStreamWrite tests whether streams see proper
// backpressure when writing data over the network streams.
func TestStBackpressureStreamWrite(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	h1, err := bhost.NewHost(swarmt.GenSwarm(t), nil)
	require.NoError(t, err)
	h2, err := bhost.NewHost(swarmt.GenSwarm(t), nil)
	require.NoError(t, err)

	// setup sender handler on 2
	h2.SetStreamHandler(protocol.TestingID, func(s network.Stream) {
		defer s.Reset()
		<-ctx.Done()
	})

	h2pi := h2.Peerstore().PeerInfo(h2.ID())
	log.Debugf("dialing %s", h2pi.Addrs)
	if err := h1.Connect(ctx, h2pi); err != nil {
		t.Fatal("Failed to connect:", err)
	}

	// open a stream, from 1->2, this is our reader
	s, err := h1.NewStream(ctx, h2.ID(), protocol.TestingID)
	require.NoError(t, err)
	defer s.Reset()

	// If nobody is reading, we should eventually time out.
	require.NoError(t, s.SetWriteDeadline(time.Now().Add(100*time.Millisecond)))
	data := make([]byte, 16*1024)
	for i := 0; i < 5*1024; i++ { // write at most 100MiB
		if _, err := s.Write(data); err != nil {
			require.True(t, os.IsTimeout(err), err)
			return
		}
	}
	t.Fatal("should have timed out")
}
