package service

import (
	"encoding/binary"
	"errors"
	"fmt"
	"github.com/libp2p/go-libp2p-core/network"
	"github.com/libp2p/go-libp2p-core/protocol"
	log "github.com/sirupsen/logrus"
	"io"
	"protonet.live/database"
	"runtime"
	"time"
)

func (tcs *TxtChatService) readTxtMsgsStream(done chan<- bool, userProtocolID protocol.ID) {
	var stream network.Stream
	var streamChan chan network.Stream
	var err error

	defer func() {
		//log.Println("returning from readTxtMsgsStream")
		if r := recover(); r != nil {
			log.Println("Recovered in readTxtMsgsStream", r)
			if stream != nil {
				_ = stream.Reset()
			}
		}
		if stream != nil {
			_ = stream.Reset()
		}
		//log.Println("returning from readTxtMsgsStream")
		done <- true
	}()
	switch userProtocolID {
	case TxtMsgServerProtocol:
		fallthrough
	case TxtMsgClientProtocol:
		streamChan = tcs.StreamInChan
	case TxtMsgLiveServerProtocol:
		fallthrough
	case TxtMsgLiveClientProtocol:
		streamChan = tcs.StreamLiveInChan
	}

	for {
		if runtime.GOOS == "js" {
			time.Sleep(time.Millisecond)
		}
		select {
		case newStream := <-streamChan:
			if stream != nil && newStream != nil {
				_ = stream.Reset()
			}
			if newStream != nil {
				stream = newStream
			}
		default:
			for stream != nil {
				//log.Println("Inside readTxtMsgsStream")
				b := make([]byte, 8)
				_, err = io.ReadFull(stream, b)
				if err != nil {
					if err == io.EOF {
						log.Println("end of file reached...", err)
					} else {
						log.Println("error occurred in readTxtMsgsStream "+
							"TxtMsgServerProtocol, in io.ReadFull, err is:", err)
					}
					_ = stream.Reset()
					stream = nil
					break
				}
				sizeOfMsg := binary.LittleEndian.Uint32(b)

				if sizeOfMsg > 0 {
					pb := make([]byte, sizeOfMsg)
					_, err = io.ReadFull(stream, pb)
					if err != nil {
						log.Println("error occurred while reading tcs.StreamInChan from client",
							tcs, "TxtMsgServerProtocol, err is:", err)
						_ = stream.Reset()
						stream = nil
						break
					}
					msg := &database.TxtMsg{}
					err = database.GetDecryptedStruct(tcs.User.PvtKeyHex, pb, msg)
					if err != nil {
						log.Println("error occurred in readTxtMsgsStream, "+
							"TxtMsgServerProtocol, in tcs.GetDecryptedProtoMessage err is:", err)
						_ = stream.Reset()
						stream = nil
						break
					}
					err = database.VerifyMessage(msg)
					if err != nil {
						err = errors.New(fmt.Sprintf("error occurred in readTxtMsgsStream,"+
							" TxtMsgServerProtocol, in tcs.verifyMessage, %v", err))
						_ = stream.Reset()
						stream = nil
						break
					}
					if newMsg := tcs.GetTxtMsg(msg.ID); newMsg == nil {
						newMsg = msg
						if err = tcs.AddNewTxtMsg(newMsg); err != nil {
							log.Println("error occurred in readTxtMsgsStream,"+
								"TxtMsgServerProtocol, in tcs.AddNewTxtMsg:", err)
						}
						if msg.CreatorID == tcs.client.IDStr {
							tcs.showNotification <- msg
						}
					} else {
						action := msg.Action
						msg = tcs.GetTxtMsg(msg.ID)
						msg.Action = action
					}
					if msg.CreatorID == tcs.User.ID {
						// if we are receiving acknowledgement response from client, we will mark the
						// message as received by the client/client
						if msg.Action.Type == database.Response {
							switch msg.Action.Message {
							case database.MessageAck:
								if !msg.AckReceivedOrSent {
									msg.AckReceivedOrSent = true
									tcs.SaveUpdateTxtMsgToDisk(tcs.User.ID, tcs.client.IDStr, msg)
								}
								// We have received the acknowledgement, hence requesting
								// for message read acknowledgement
								if !msg.ReadAckReceivedOrSent {
									msg.Action.Type = database.Request
									msg.Action.Message = database.MessageReadAck
									tcs.TxtMsgOutChan <- msg
									if userProtocolID == tcs.GetUserMsgLiveProtocol() {
										tcs.TxtMsgLiveOutChan <- msg
									}
								}
							case database.MessageReadAck:
								// since we are receiving read acknowledgement response,
								// it indicates message is already acknowledged

								if !msg.AckReceivedOrSent ||
									!msg.ReadAckReceivedOrSent {
									msg.AckReceivedOrSent = true
									msg.ReadAckReceivedOrSent = true
									tcs.SaveUpdateTxtMsgToDisk(tcs.User.ID, tcs.client.IDStr, msg)
								}
								// since we have received the read acknowledgement, it indicates
								// all the previous acknowledged msgs are read
								//tcs.MarkAllUserAckMsgsAsReadAck(msg.Timestamp)
							case database.MessageNotFound:
								log.Println("Received Message Not Found, ", msg)
								// This indicates the message exists with us but doesn't exists with the client/client
								//newState := database.TextMessageState{
								//	LastTried:             msg.GetLastTried(),
								//	AckReceivedOrSent:     msg.AckReceivedOrSent,
								//	ReadAckReceivedOrSent: msg.ReadAckReceivedOrSent,
								//	MsgRead:               state.MsgRead,
								//}
								//err := tcs.updateTxtMsgState(msg.ID, &newState)
								//if err != nil {
								//	log.Println("error occurred in readTxtMsgsActionStream, "+
								//		"in MessageNotFound, "+
								//		"in tcs.updateTxtMsgState, message belongs to User", err)
								//}
								//tcs.SendProtoTextMessage(msg)
								//tcs.GetTxtMsgOutChan() <- msg
							}
						}
					}
					if msg.CreatorID == tcs.client.IDStr {
						if msg.Action.Type == database.Request {
							switch msg.Action.Message {
							// since msg already exists, we can send the acknowledgement
							case database.MessageAck:
								msg.Action.Type = database.Response
								msg.Action.Message = database.MessageAck
								tcs.TxtMsgOutChan <- msg
								if userProtocolID == tcs.GetUserMsgLiveProtocol() {
									tcs.TxtMsgLiveOutChan <- msg
								}
							case database.MessageReadAck:
								if msg.MsgRead {
									msg.Action.Type = database.Response
									msg.Action.Message = database.MessageReadAck
									tcs.TxtMsgOutChan <- msg
									if userProtocolID == tcs.GetUserMsgLiveProtocol() {
										tcs.TxtMsgLiveOutChan <- msg
									}
								}
							}
						}
					}
				}
				break
			}
		}
	}
}
