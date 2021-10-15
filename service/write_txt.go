package service

import (
	"bufio"
	"encoding/binary"
	"github.com/libp2p/go-libp2p-core/network"
	"github.com/libp2p/go-libp2p-core/protocol"
	log "github.com/sirupsen/logrus"
	"protonet.live/database"
	"runtime"
	"time"
)

func (tcs *TxtChatService) writeTxtMsgsStream(done chan<- bool, clientMsgProtocolID protocol.ID) {
	var stream network.Stream
	var txtMsgOutChannel chan *database.TxtMsg
	var err error
	defer func() {
		//log.Println("returning from writeTxtMsgsStream")
		done <- true
	}()

	switch clientMsgProtocolID {
	case TxtMsgServerProtocol:
		fallthrough
	case TxtMsgClientProtocol:
		txtMsgOutChannel = tcs.TxtMsgOutChan
	case TxtMsgLiveServerProtocol:
		fallthrough
	case TxtMsgLiveClientProtocol:
		txtMsgOutChannel = tcs.TxtMsgLiveOutChan
	}

	for {
		if runtime.GOOS == "js" {
			time.Sleep(time.Millisecond)
		}
		switch stream == nil {
		case true:
			stream = tcs.NewClientStream(clientMsgProtocolID)
		default:
			for eachMessage := range txtMsgOutChannel {
				//log.Println("sending msg from txmsgOutchannel....")
				//log.Println("Inside writeTxtMsgsStream")
				rw := bufio.NewWriter(stream)
				err = database.SignMessage(tcs.User.PvtKeyHex, eachMessage)
				if err != nil {
					log.Println("error occurred in writeTxtMsgsStream,"+
						" in tcs.verifyMessage, err:", err)
					break
				}

				bytes, newErr := database.GetEncryptedStruct(tcs.client.PublicKey, eachMessage)
				if newErr != nil {
					log.Println("Error occurred in writeTxtMsgsStream in"+
						" tcs.GetEncryptedProtoMessage, newErr:", newErr)
					break
				}
				messageSize := uint32(len(bytes))
				b := make([]byte, 8)
				binary.LittleEndian.PutUint32(b, messageSize)
				_, err = rw.Write(b)
				if err != nil {
					log.Println("Error occurred in writeTxtMsgsStream,"+
						" in rw.Write(b) error is", err)
					_ = stream.Reset()
					stream = nil
					break
				}
				_, err = rw.Write(bytes)
				if err != nil {
					log.Println("Error occurred in writeTxtMsgsStream,"+
						" in rw.Write(bytes) error is", err)
					_ = stream.Reset()
					stream = nil
					break
				}
				err = rw.Flush()
				if err != nil {
					log.Println("Error occurred in writeTxtMsgsStream, rw.Flush error is:",
						err)
					_ = stream.Reset()
					stream = nil
					break
				}
				//log.Println("successfully written text message")
				break
			}
		}
	}
}
