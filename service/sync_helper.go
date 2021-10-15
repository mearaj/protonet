package service

import (
	log "github.com/sirupsen/logrus"
	"protonet.live/database"
	"time"
)

func (tcs *TxtChatService) msgsSyncHelper(done chan<- bool) {
	filteredMessages := make([]*database.TxtMsg, 0, 0)
	var lastReadAckMsgTimestamp int64
	defer func() {
		//log.Println("returning from msgsSyncHelper", len(filteredMessages))
		done <- true
	}()

	filteredMessages = tcs.GetPendingMessages()
	if len(filteredMessages) < 1 {
		return
	}

	time.Sleep(time.Second * 1)
	for _, msg := range filteredMessages {
		// if we have received the read acknowledgement then all the previous acknowledged msgs
		// should be marked as read acknowledged
		if msg.CreatorID == tcs.User.ID &&
			msg.ReadAckReceivedOrSent &&
			msg.Timestamp > lastReadAckMsgTimestamp {
			lastReadAckMsgTimestamp = msg.Timestamp
		}
	}
	if lastReadAckMsgTimestamp > 0 {
		tcs.MarkAllUserAckMsgsAsReadAck(lastReadAckMsgTimestamp)
	}
	for _, msg := range filteredMessages {
		if msg.CreatorID == tcs.User.ID {
			if !msg.AckReceivedOrSent {
				msg.Action.Type = database.Response
				msg.Action.Message = database.MessageAck
				tcs.TxtMsgOutChan <- msg
			}

			if msg.ReadAckReceivedOrSent &&
				!msg.AckReceivedOrSent {
				msg.AckReceivedOrSent = true
				tcs.SaveUpdateTxtMsgToDisk(tcs.User.ID, tcs.client.IDStr, msg)
			}
			if !msg.ReadAckReceivedOrSent {
				msg.Action.Type = database.Request
				msg.Action.Message = database.MessageReadAck
				log.Println("sending request for messageReadAck to outchannel")
				tcs.TxtMsgOutChan <- msg
			}
		}
		if msg.CreatorID == tcs.client.IDStr {
			// we will resend the response if we haven't sent the response acknowledgement
			if !msg.AckReceivedOrSent {
				msg.Action.Type = database.Response
				msg.Action.Message = database.MessageAck
				tcs.TxtMsgOutChan <- msg
			}
			if msg.MsgRead && !msg.ReadAckReceivedOrSent {
				msg.Action.Type = database.Response
				msg.Action.Message = database.MessageReadAck
				tcs.TxtMsgOutChan <- msg
			}
		}
	}
}
