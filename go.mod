module protonet.live

go 1.16

require (
	gioui.org v0.0.0-20210319204632-1dde94d8ddc0
	github.com/gogo/protobuf v1.3.2 // indirect
	github.com/google/gopacket v1.1.19 // indirect
	github.com/ipfs/go-log/v2 v2.1.3 // indirect
	github.com/minio/sha256-simd v1.0.0 // indirect
	github.com/shurcooL/go-goon v0.0.0-20210110234559-7585751d9a17 // indirect
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
