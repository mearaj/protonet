module protonet.live/service

go 1.16

require (
	github.com/golang/groupcache v0.0.0-20191027212112-611e8accdfc9 // indirect
	github.com/google/uuid v1.2.0
	github.com/libp2p/go-libp2p v0.13.0
	github.com/libp2p/go-libp2p-connmgr v0.2.4
	github.com/libp2p/go-libp2p-core v0.8.5
	github.com/libp2p/go-libp2p-kad-dht v0.11.1
	github.com/libp2p/go-libp2p-mplex v0.4.1 // indirect
	github.com/libp2p/go-libp2p-secio v0.2.3
	github.com/libp2p/go-libp2p-webrtc-direct v0.0.0-20210321133143-d937759fb030
	github.com/libp2p/go-netroute v0.1.4 // indirect
	github.com/pion/webrtc/v3 v3.0.16 // indirect
	protonet.live/database v0.0.0
	github.com/sirupsen/logrus v1.8.1
)

replace protonet.live/database => ./../database
