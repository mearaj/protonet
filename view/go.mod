module protonet.live/view

go 1.17

require (
	gioui.org v0.0.0-20210911073124-dae3b0fa5a9b
	gioui.org/x v0.0.0-20210816192830-9ea938c228a0
	github.com/libp2p/go-libp2p-core v0.9.0
	github.com/sirupsen/logrus v1.8.1
	golang.org/x/exp v0.0.0-20210722180016-6781d3edade3
	golang.org/x/image v0.0.0-20210628002857-a66eb6448b8d
	protonet.live/database v0.0.0
	protonet.live/jni v0.0.0
	protonet.live/service v0.0.0
)

replace (
	protonet.live/database => ../database
	protonet.live/jni => ../jni
	protonet.live/service => ../service
)
