module protonet.live/service

go 1.17

require (
	github.com/google/uuid v1.3.0
	github.com/libp2p/go-libp2p v0.15.0
	github.com/libp2p/go-libp2p-connmgr v0.2.4
	github.com/libp2p/go-libp2p-core v0.9.0
	github.com/libp2p/go-libp2p-kad-dht v0.13.1
	//github.com/libp2p/go-libp2p-secio v0.2.3
	github.com/libp2p/go-libp2p-webrtc-direct v0.0.0-20210521184902-a8cb0f997a78
	github.com/sirupsen/logrus v1.8.1
	protonet.live/database v0.0.0
)

replace protonet.live/database => ./../database
