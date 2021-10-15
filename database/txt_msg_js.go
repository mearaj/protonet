package database

import (
	"fmt"
	"strconv"
	"syscall/js"
)

func (db *Database) LoadTxtMsgsFromDisk(userID string, contactID string) <-chan TxtMsgs {
	txtMsgsChan := make(chan TxtMsgs)
	go func() {
		txtMsgs := TxtMsgs{}
		if db.db == nil {
			select {
			case txtMsgsChan <- txtMsgs:
			default:
			}
			return
		}
		dbi := db.db.(js.Value)
		if !dbi.IsUndefined() && !dbi.IsNull() {
			req := dbi.Call("transaction", js.ValueOf(TextMessagesDir), js.ValueOf("readwrite")).Call("objectStore", js.ValueOf(TextMessagesDir)).Call("get", js.ValueOf(userID))
			req.Set("onsuccess", js.FuncOf(func(this js.Value, args []js.Value) interface{} {
				txtMsgsStore := args[0].Get("target").Get("result")
				if !txtMsgsStore.IsNull() && !txtMsgsStore.IsUndefined() {
					txtMsgsObj := txtMsgsStore.Get("txtMsgs")
					keysJs := js.Global().Get("Object").Call("keys", txtMsgsObj)
					for j := 0; j < keysJs.Get("length").Int(); j++ {
						txtMsgObj := txtMsgsObj.Get(keysJs.Index(j).String())
						timestampNano, _ := strconv.ParseInt(txtMsgObj.Get("TimestampNano").String(), 10, 64)
						timeStamp, _ := strconv.ParseInt(txtMsgObj.Get("Timestamp").String(), 10, 64)
						txtMsg := TxtMsg{
							CreatorID:             txtMsgObj.Get("CreatorID").String(),
							Timestamp:             timeStamp,
							ID:                    txtMsgObj.Get("ID").String(),
							CreatorPublicKey:      txtMsgObj.Get("CreatorPublicKey").String(),
							Sign:                  []byte(txtMsgObj.Get("Sign").String()),
							Message:               txtMsgObj.Get("Message").String(),
							TimestampNano:         timestampNano,
							AckReceivedOrSent:     txtMsgObj.Get("State").Get("AckReceivedOrSent").Bool(),
							ReadAckReceivedOrSent: txtMsgObj.Get("State").Get("ReadAckReceivedOrSent").Bool(),
							MsgRead:               txtMsgObj.Get("State").Get("MsgRead").Bool(),
						}
						txtMsgs[txtMsg.ID] = &txtMsg
					}
				} else {
					req = dbi.Call("transaction", js.ValueOf(TextMessagesDir), js.ValueOf("readwrite")).Call("objectStore", js.ValueOf(TextMessagesDir))
					req.Call("put", js.ValueOf(map[string]interface{}{
						"ID":      userID,
						"txtMsgs": js.ValueOf(map[string]interface{}{}),
					}))
				}
				select {
				case txtMsgsChan <- txtMsgs:
				default:
				}
				return nil
			}))
		} else {
			select {
			case txtMsgsChan <- txtMsgs:
			default:
			}
		}
	}()
	return txtMsgsChan
}

func (db *Database) SaveAllTxtMsgsToDisk(userID string, contactID string, messages TxtMsgs) {
	for _, msg := range messages {
		db.SaveTxtMsgToDisk(userID, contactID, msg)
	}
}

func (db *Database) SaveTxtMsgToDisk(userID string, contactID string, message *TxtMsg) {
	if db.db == nil {
		return
	}
	dbi := db.db.(js.Value)
	if !dbi.IsUndefined() && !dbi.IsNull() {
		req := dbi.Call("transaction", js.ValueOf(TextMessagesDir), js.ValueOf("readwrite")).Call("objectStore", js.ValueOf(TextMessagesDir)).Call("get", js.ValueOf(userID))
		req.Set("onsuccess", js.FuncOf(func(this js.Value, args []js.Value) interface{} {
			txtMsgsStore := args[0].Get("target").Get("result")
			if txtMsgsStore.IsNull() || txtMsgsStore.IsUndefined() {
				req.Call("put", js.ValueOf(map[string]interface{}{
					"ID":      userID,
					"txtMsgs": js.ValueOf(map[string]interface{}{}),
				}))
				txtMsgsStore = args[0].Get("target").Get("result")
			}
			txtMsgsJs := txtMsgsStore.Get("txtMsgs")
			txtMsg := js.ValueOf(map[string]interface{}{
				"CreatorID":        message.CreatorID,
				"Timestamp":        fmt.Sprintf("%d", message.Timestamp),
				"ID":               message.ID,
				"CreatorPublicKey": message.CreatorPublicKey,
				"Sign":             string(message.Sign),
				"Message":          message.Message,
				"TimestampNano":    fmt.Sprintf("%d", message.TimestampNano),
				"State": map[string]interface{}{
					"AckReceivedOrSent":     message.AckReceivedOrSent,
					"ReadAckReceivedOrSent": message.ReadAckReceivedOrSent,
					"MsgRead":               message.MsgRead,
				},
			})
			txtMsgsJs.Set(message.ID, txtMsg)
			req = dbi.Call("transaction", js.ValueOf(TextMessagesDir), js.ValueOf("readwrite")).Call("objectStore", js.ValueOf(TextMessagesDir))
			req.Call("put", js.ValueOf(map[string]interface{}{
				"ID":      userID,
				"txtMsgs": txtMsgsJs,
			}))
			return nil
		}))
	}
}
