module protonet.live

go 1.16

require (
	gioui.org v0.0.0-20210319204632-1dde94d8ddc0
	gioui.org/cmd v0.0.0-20210323222646-238dd1aa863e // indirect
	github.com/google/gopacket v1.1.19 // indirect
	github.com/minio/sha256-simd v1.0.0 // indirect
	github.com/sirupsen/logrus v1.8.1
	golang.org/x/crypto v0.0.0-20210220033148-5ea612d1eb83 // indirect
	protonet.live/view v0.0.0
)

replace (
	protonet.live/database => ./database
	protonet.live/jni => ./jni
	protonet.live/service => ./service
	protonet.live/view => ./view
)
