package service

import (
	"context"
	"github.com/libp2p/go-libp2p"
	connmgr "github.com/libp2p/go-libp2p-connmgr"
	"github.com/libp2p/go-libp2p-core/host"
	"github.com/libp2p/go-libp2p-core/routing"
	dht "github.com/libp2p/go-libp2p-kad-dht"
	mplex "github.com/libp2p/go-libp2p-mplex"
	libp2pwebrtcdirect "github.com/libp2p/go-libp2p-webrtc-direct"
	"github.com/pion/webrtc/v3"
	log "github.com/sirupsen/logrus"
	"protonet.live/database"
	"runtime"
	"time"
)

func (cs *ChatService) createHost() () {
	pvtKey, err := database.GetPrivateKeyFromHex(cs.GetCurrentUser().PvtKeyHex)
	if err != nil {
		log.Println("error in createHost, in GetPrivateKeyFromHex err:", err)
		return
	}
	webTrs := libp2pwebrtcdirect.NewTransport(
		webrtc.Configuration{},
		new(mplex.Transport),
	)


	listenAddressStrings := []string{
		"/ip4/0.0.0.0/tcp/0", // regular tcp connections
		"/ip4/0.0.0.0/udp/0", // regular tcp connections
		"/ip4/0.0.0.0/tcp/0/ws",
		"/ip4/127.0.0.1/tcp/4005/http/p2p-webrtc-direct",
	}
	// Attempt to open ports using uPNP for NATed hosts.
	var natPortMap libp2p.Option

	if runtime.GOOS != "js" {
		listenAddressStrings = append(listenAddressStrings, "/ip4/0.0.0.0/udp/0/quic")
		natPortMap = libp2p.NATPortMap()
	}
	listenAddrsOption := libp2p.ListenAddrStrings(listenAddressStrings...)

	retry := 0
	for retry < 5 {

		// create a new libp2p Host that listens on a random TCP port
		cs.Host, err = libp2p.New(
			//context.Background(),
			// libp2p.Identity(ua.PvtKey),
			libp2p.Identity(pvtKey),
			listenAddrsOption,
			libp2p.DefaultTransports,
			libp2p.DefaultEnableRelay,
			libp2p.DefaultMuxers,
			libp2p.EnableNATService(),
			libp2p.DefaultSecurity,
			libp2p.DefaultPeerstore,
			libp2p.DefaultListenAddrs,
			libp2p.ConnectionManager(connmgr.NewConnManager(
				500,            // Lowwater
				1000,           // HighWater,
				time.Minute*10, // GracePeriod previous value time.Minute * 10
			)),
			natPortMap,
			libp2p.Transport(webTrs),
			// Let this Host use the DHT to find other hosts
			libp2p.Routing(func(h host.Host) (routing.PeerRouting, error) {
				var idht *dht.IpfsDHT
				idht, err := dht.New(context.Background(), h)
				return idht, err
			}),
		)
		if err != nil {
			log.Println("err creating libp2p Host, err is", err)
			log.Println("Retrying creating Host, max tries is", 5, "this is", retry+1, "attempt")
		} else {
			break
		}
		retry++
	}
	if err != nil {
		log.Println(err)
	}
	log.Println(cs.Host.Network().ListenAddresses())
	return
}