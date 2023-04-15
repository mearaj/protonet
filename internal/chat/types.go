package chat

import (
	"github.com/libp2p/go-libp2p/core/protocol"
	"github.com/mearaj/protonet/internal/model"
)

type Account = model.Account
type Contact = model.Contact

const (
	ProtocolChat protocol.ID = "/protonet.wallet/msg-chat/0.0.1"
)

const (
	MessageStateStateless = model.MessageStateStateless
	MessageStateReceived  = model.MessageStateReceived
	MessageStateRead      = model.MessageStateRead
)

type Message = model.Message
