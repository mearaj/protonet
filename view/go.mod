module protonet.live/view

go 1.16

require (
	gioui.org v0.0.0-20210311180434-c4850e876d02
	gioui.org/x v0.0.0-20210226015410-958111222865
	github.com/libp2p/go-libp2p-core v0.8.5
	golang.org/x/exp v0.0.0-20201229011636-eab1b5eb1a03
	golang.org/x/image v0.0.0-20210220032944-ac19c3e999fb
	protonet.live/database v0.0.0
	protonet.live/jni v0.0.0
	protonet.live/service v0.0.0
)

replace (
	protonet.live/database => ../database
	protonet.live/jni => ../jni
	protonet.live/service => ../service
)
