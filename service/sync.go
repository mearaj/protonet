//go:build !js

package service

import (
	"bufio"
	"context"
	"encoding/binary"
	"encoding/gob"
	"encoding/hex"
	"github.com/google/uuid"
	"github.com/libp2p/go-libp2p-core/network"
	"github.com/mearaj/protonet/alog"
	"io"
)

type SyncType int

const (
	SyncRequest SyncType = iota
	SyncResponse
)

// Sync can act like header or body for the request/response on sync protocol
type Sync struct {
	Type      SyncType
	Delimiter string
	Metadata  interface{}
	Data      interface{}
}
type SyncMessages struct {
	// string represents slices of messages ids
	MessagesReceived []string
	MessagesRead     []string
	MessagesNotFound []string
}

func init() {
	gob.Register(SyncMessages{})
}
func (s *service) handleSyncStream(stream network.Stream) {
	pubKey := stream.Conn().RemotePublicKey()
	pubKeyBytes, err := pubKey.Raw()
	if err != nil {
		alog.Logger().Errorln(err)
		return
	}
	pubKeyStr := hex.EncodeToString(pubKeyBytes)
	s.syncStreams.Add(pubKeyStr, stream)
	go s.writeSyncStream(stream, pubKeyStr)
	go s.readSyncStream(stream, pubKeyStr)
}

func (s *service) readSyncStream(stream network.Stream, pubKeyHex string) <-chan struct{} {
	structChan := make(chan struct{}, 1)
	go func() {
		var err error
		account := s.Account()
		defer func() {
			recoverPanicCloseCh(structChan, struct{}{}, alog.Logger())
			if err != nil && err.Error() == "stream reset" {
				s.syncStreams.Delete(pubKeyHex)
			}
		}()
		for err == nil || (err.Error() != "stream reset") {
			if account.PublicKey != s.Account().PublicKey {
				return
			}
			// the first 8 bytes indicates the size of Sync struct
			b := make([]byte, 8)
			_, err = io.ReadFull(stream, b)
			if err != nil {
				alog.Logger().Errorln(err)
				continue
			}
			sizeOfSync := binary.LittleEndian.Uint32(b)
			if sizeOfSync > 0 {
				syncBytes := make([]byte, sizeOfSync)
				_, err = io.ReadFull(stream, syncBytes)
				if err != nil {
					alog.Logger().Errorln(err)
					continue
				}
				var syncData Sync
				err = DecodeToStruct(&syncData, syncBytes)
				if err != nil {
					alog.Logger().Errorln(err)
					continue
				}
				switch syncData.Type {
				case SyncRequest:
					var respMsgsRecv, respMsgsNotFound, respMsgsRead []string
					if data, ok := syncData.Data.(SyncMessages); ok {
						queryStr := "account_public_key = ? and `from` = ? and id = ?"
						for _, msgRec := range data.MessagesReceived {
							msg := Message{}
							txn := s.GormDB().Find(&msg, queryStr, account.PublicKey, pubKeyHex, msgRec)
							if txn.RowsAffected == 0 {
								respMsgsNotFound = append(respMsgsNotFound, msgRec)
							} else {
								if msg.Read {
									respMsgsRead = append(respMsgsRead, msgRec)
								} else {
									respMsgsRecv = append(respMsgsRecv, msgRec)
								}
							}
						}
						for _, msgRd := range data.MessagesRead {
							msg := Message{}
							txn := s.GormDB().Find(&msg, queryStr, account.PublicKey, pubKeyHex, msgRd)
							if txn.RowsAffected == 0 {
								respMsgsNotFound = append(respMsgsNotFound, msgRd)
							} else {
								if msg.Read {
									respMsgsRead = append(respMsgsRead, msgRd)
								}
							}
						}
						syncResp := Sync{
							Type: SyncResponse,
							Data: SyncMessages{
								MessagesReceived: respMsgsRecv,
								MessagesRead:     respMsgsRead,
								MessagesNotFound: respMsgsNotFound,
							},
						}
						if ch, ok := s.syncStreamsOutCh.Value(pubKeyHex); ok {
							select {
							case ch <- syncResp:
							default:
							}
						}
					}
				case SyncResponse:
					if data, ok := syncData.Data.(SyncMessages); ok {
						stateChanged := false
						for _, msgRec := range data.MessagesReceived {
							queryStr := "account_public_key = ? and `to` = ? and id = ? and state < ?"
							txn := s.GormDB().Model(Message{}).Where(
								queryStr, account.PublicKey, pubKeyHex, msgRec, MessageRead,
							).Updates(Message{State: MessageReceivedSent})
							if txn.Error != nil {
								alog.Logger().Errorln(txn.Error)
							}
							if txn.RowsAffected > 0 {
								stateChanged = true
							}
						}
						for _, msgRd := range data.MessagesRead {
							queryStr := "account_public_key = ? and `to` = ? and id = ? and state != ?"
							txn := s.GormDB().Model(Message{}).Where(
								queryStr, account.PublicKey, pubKeyHex, msgRd, MessageRead,
							).Updates(Message{State: MessageRead})
							if txn.Error != nil {
								alog.Logger().Errorln(txn.Error)
							}
							if txn.RowsAffected > 0 {
								stateChanged = true
							}
						}
						for _, msgNotFound := range data.MessagesNotFound {
							var id uuid.UUID
							id, err = uuid.Parse(msgNotFound)
							if err != nil {
								alog.Logger().Errorln(err)
								continue
							}
							msg := Message{ID: id}
							queryStr := "account_public_key = ? and `to` = ? and id = ?"
							txn := s.GormDB().Find(&msg, queryStr, account.PublicKey, pubKeyHex, msgNotFound)
							if txn.RowsAffected > 0 {
								peerID := stream.Conn().RemotePeer()
								hst := s.Host()
								if _, ok := s.chatStreams.Value(pubKeyHex); !ok {
									stream, err := hst.NewStream(context.Background(), peerID, ProtocolChat)
									if err != nil {
										alog.Logger().Errorln(err)
										continue
									}
									s.handleHostChatStream(stream)
								}

								if msgCh, ok := s.chatStreamsOutCh.Value(pubKeyHex); ok {
									select {
									case msgCh <- msg:
									default:
									}
								}
							}
						}
						if stateChanged {
							s.eventBroker.Fire(Event{
								Data: MessagesStateChangedEventData{
									AccountPublicKey: account.PublicKey,
									ContactPublicKey: pubKeyHex,
								},
								Topic: MessagesStateChangedEventTopic,
							})
						}
					}
				}
			}
		}
	}()
	return structChan
}

func (s *service) writeSyncStream(stream network.Stream, pubKeyHex string) <-chan struct{} {
	structChan := make(chan struct{}, 1)
	go func() {
		var err error
		defer func() {
			recoverPanicCloseCh(structChan, struct{}{}, alog.Logger())
			if err != nil && err.Error() == "stream reset" {
				s.syncStreams.Delete(pubKeyHex)
			}
		}()
		totalMessages := int(<-s.MessagesCount(pubKeyHex))
		limit := 10
		account := s.Account()
		offset := 0
	continueLabel:
		// if current account is changed, then return
		if account.PublicKey != s.Account().PublicKey {
			return
		}
		if ch, ok := s.syncStreamsOutCh.Value(pubKeyHex); ok {
			shouldBreak := false
			for {
				select {
				case syncStr, ok := <-ch:
					if ok {
						bs := EncodeToBytes(&syncStr)
						rw := bufio.NewWriter(stream)
						msgSize := uint32(len(bs))
						bi := make([]byte, 8) // carries msgSize info in 8 bytes
						binary.LittleEndian.PutUint32(bi, msgSize)
						bi = append(bi, bs...)
						_, err = rw.Write(bi)
						if err != nil {
							alog.Logger().Errorln(err)
							if err.Error() == "stream reset" {
								return
							}
							break
						}
						err = rw.Flush()
						if err != nil {
							alog.Logger().Errorln(err)
							if err.Error() == "stream reset" {
								return
							}
							break
						}
					}
				default:
					shouldBreak = true
				}
				// break out of for loop
				if shouldBreak {
					break
				}
			}
		}
		if offset < totalMessages {
			var msgs []Message
			queryStr := "account_public_key = ? and `to` = ? and state = ?"
			txn := s.GormDB().Order("created desc").Offset(offset).Limit(limit).Find(
				&msgs, queryStr, account.PublicKey, pubKeyHex, MessageStateless)
			if txn.Error != nil {
				alog.Logger().Errorln(txn.Error)
			}
			var reqMsgsRecv, reqMsgsRead []string
			for _, eachMessage := range msgs {
				reqMsgsRecv = append(reqMsgsRecv, eachMessage.ID.String())
				reqMsgsRead = append(reqMsgsRead, eachMessage.ID.String())
			}
			msgs = []Message{}
			txn = s.GormDB().Order("created desc").Offset(offset).Limit(limit).Find(
				&msgs, queryStr, account.PublicKey, pubKeyHex, MessageReceivedSent)
			if txn.Error != nil {
				alog.Logger().Errorln(txn.Error)
			}
			for _, eachMessage := range msgs {
				reqMsgsRead = append(reqMsgsRead, eachMessage.ID.String())
			}
			if len(reqMsgsRead) > 0 || len(reqMsgsRecv) > 0 {
				syncReq := Sync{
					Type: SyncRequest,
					Data: SyncMessages{
						MessagesReceived: reqMsgsRecv,
						MessagesRead:     reqMsgsRead,
					},
				}
				bs := EncodeToBytes(&syncReq)
				rw := bufio.NewWriter(stream)
				msgSize := uint32(len(bs))
				bi := make([]byte, 8) // carries msgSize info in 8 bytes
				binary.LittleEndian.PutUint32(bi, msgSize)
				bi = append(bi, bs...)
				_, err = rw.Write(bi)
				if err != nil {
					alog.Logger().Errorln(err)
					if err.Error() == "stream reset" {
						return
					}
				}
				err = rw.Flush()
				if err != nil {
					alog.Logger().Errorln(err)
					if err.Error() == "stream reset" {
						return
					}
				}
			}
			offset += limit
			goto continueLabel
		}
		offset = 0
		goto continueLabel
	}()
	return structChan
}
