package service

import (
	"errors"
	"github.com/mearaj/protonet/alog"
	"syscall/js"
)

// configs for indexeddb
var keyConfigAccount = js.ValueOf(map[string]interface{}{"keyPath": "Id"}) // Id value is 0
var keyConfigAccounts = js.ValueOf(map[string]interface{}{"keyPath": "PublicKey"})
var keyConfigContacts = js.ValueOf(map[string]interface{}{
	"keyPath": []interface{}{"AccountAdr", "PublicKey"},
})
var keyConfigMessages = js.ValueOf(map[string]interface{}{
	"keyPath": []interface{}{"AccountAdr", "ContactPublicKey"},
})

func (s *service) init() {
	defer func() {
		recoverPanic(alog.Logger())
	}()
	indexedDBReq := js.Global().Get("indexedDB")
	req := indexedDBReq.Call("open", js.ValueOf(DBPathCfgDir), js.ValueOf(AppIndexedDBVersion))
	req.Set("onsuccess", js.FuncOf(s.onInitSuccess))
	req.Set("onupgradeneeded", js.FuncOf(s.onUpgradeNeeded))
	req.Set("onerror", js.FuncOf(s.onInitError))
}

func (s *service) IndexedDB() *js.Value {
	if s.database == (*js.Value)(nil) {
		return nil
	}
	if indexedDB, ok := s.database.(*js.Value); ok {
		return indexedDB
	}
	return nil
}

func (s *service) setIndexedDB(indexedDB *js.Value) {
	s.database = indexedDB
}

func (s *service) onInitSuccess(this js.Value, args []js.Value) interface{} {
	indexedDB := args[0].Get("target").Get("result")
	s.setIndexedDB(&indexedDB)
	s.initialized = true
	return nil
}

func (s *service) onUpgradeNeeded(this js.Value, args []js.Value) interface{} {
	defer func() {
		recoverPanic(alog.Logger())
	}()
	indexedDB := args[0].Get("target").Get("result")
	currentVersion := indexedDB.Get("version").Int()
	oldVersion := args[0].Get("oldVersion").Int()
	newVersion := args[0].Get("newVersion").Int() // equivalent to currentVersion
	if oldVersion < 1 {
		// The app is installed first time
		_ = currentVersion
		_ = newVersion
	}
	// Currently, regardless of the version, check if store items exists and if not then create one
	indexedDB.Call("createObjectStore", js.ValueOf(DBPathAccountsDir), keyConfigAccounts)
	indexedDB.Call("createObjectStore", js.ValueOf(DBPathCurrentAccountDir), keyConfigAccount)
	indexedDB.Call("createObjectStore", js.ValueOf(DBPathContactsDir), keyConfigContacts)
	indexedDB.Call("createObjectStore", js.ValueOf(DBPathMessagesDir), keyConfigMessages)
	return nil
}
func (s *service) onInitError(this js.Value, args []js.Value) interface{} {
	errorJs := args[0].Get("target").Get("errorCode")
	alog.Logger().Error(errorJs)
	return nil
}

// saveAccountToDB
func (s *service) saveAccount(account Account) <-chan error {
	errCh := make(chan error, 1)
	var err error
	go func() {
		defer func() { recoverPanicCloseCh(errCh, err, alog.Logger()) }()
		txn := s.IndexedDB().Call("transaction", js.ValueOf(DBPathCurrentAccountDir), js.ValueOf("readwrite"))
		objectStore := txn.Call("objectStore", js.ValueOf(DBPathCurrentAccountDir))
		req := objectStore.Call("clear")
		val := map[string]interface{}{
			"Id":        0,
			"PublicKey": account.Addr,
			"Contents":  string(account.Contents),
			"StateStr":  account.StateStr,
		}
		req = objectStore.Call("put", val)
		errCh2 := make(chan error, 1)
		req.Set("onsuccess", OnSuccess(errCh2))
		req.Set("onerror", OnError(errCh2))
		err = <-errCh2
		s.setAccount(account)
		s.reLoadServiceChan <- struct{}{}
	}()
	return errCh
}

func (s *service) reqAccount() <-chan Account {
	accountCh := make(chan Account, 1)
	go func() {
		var account Account
		defer func() {
			recoverPanicCloseCh(accountCh, account, alog.Logger())
		}()
		txn := s.IndexedDB().Call("transaction", js.ValueOf(DBPathCurrentAccountDir), "readonly")
		objStore := txn.Call("objectStore", js.ValueOf(DBPathCurrentAccountDir))
		req := objStore.Call("getAll")
		waitCh := make(chan error, 1)
		req.Set("onsuccess", OnSuccess(waitCh))
		req.Set("onerror", OnError(waitCh))
		err := <-waitCh
		if err != nil {
			alog.Logger().Errorln(err)
			return
		}
		accountsJS := req.Get("target").Get("result")
		sizeOFArray := accountsJS.Get("length").Int()
		if sizeOFArray > 0 {
			eachAccountJS := accountsJS.Index(0)
			addrContents := eachAccountJS.Get("Contents").String()
			stateStr := eachAccountJS.Get("State").String()
			addr := eachAccountJS.Get("PublicKey").String()
			avatar := eachAccountJS.Get("Avatar").String()
			account = Account{
				Addr:     addr,
				Avatar:   []byte(avatar),
				Contents: []byte(addrContents),
				StateStr: stateStr,
			}
			s.setAccount(account)
		}
	}()
	return accountCh
}

func (s *service) reqAccounts() <-chan []Account {
	accountsCh := make(chan []Account, 1)
	go func() {
		accounts := make([]Account, 0)
		defer func() { recoverPanicCloseCh(accountsCh, accounts, alog.Logger()) }()
		txn := s.IndexedDB().Call("transaction", js.ValueOf(DBPathAccountsDir), "readonly")
		objStore := txn.Call("objectStore", js.ValueOf(DBPathAccountsDir))
		req := objStore.Call("getAll")
		accountsJS := req.Get("target").Get("result")
		waitCh := make(chan error, 1)
		req.Set("onsuccess", OnSuccess(waitCh))
		req.Set("onerror", OnError(waitCh))
		err := <-waitCh
		if err != nil {
			alog.Logger().Errorln(err)
			return
		}
		if accountsJS.Truthy() {
			sizeOFArray := accountsJS.Get("length").Int()
			for i := 0; i < sizeOFArray; i++ {
				eachAccountJS := accountsJS.Index(i)
				addrContents := eachAccountJS.Get("Contents").String()
				stateStr := eachAccountJS.Get("State").String()
				addr := eachAccountJS.Get("PublicKey").String()
				avatar := eachAccountJS.Get("Avatar").String()
				account := Account{
					Addr:     addr,
					Avatar:   []byte(avatar),
					Contents: []byte(addrContents),
					StateStr: stateStr,
				}
				accounts = append(accounts, account)
			}
		}
	}()
	return accountsCh
}

func (s *service) reqContacts() <-chan []Contact {
	contactsCh := make(chan []Contact, 1)
	contacts := make([]Contact, 0)
	if s.Account().Addr == "" {
		contactsCh <- contacts
		close(contactsCh)
		return contactsCh
	}
	go func() {
		defer func() { recoverPanicCloseCh(contactsCh, contacts, alog.Logger()) }()
		txn := s.IndexedDB().Call("transaction", js.ValueOf(DBPathContactsDir), "readwrite")
		objectStore := txn.Call("objectStore", js.ValueOf(DBPathContactsDir))
		accountAddr := s.Account().Addr
		req := objectStore.Call("get", accountAddr)
		contactsJs := req.Get("target").Get("result")
		waitCh := make(chan error, 1)
		req.Set("onsuccess", OnSuccess(waitCh))
		req.Set("onerror", OnError(waitCh))
		err := <-waitCh
		if err != nil {
			alog.Logger().Errorln(err)
			return
		}
		if contactsJs.Truthy() {
			contactsJsArr := contactsJs.Get("Contacts")
			sizeOFArray := contactsJsArr.Get("length").Int()
			for i := 0; i < sizeOFArray; i++ {
				contactJs := contactsJsArr.Index(i)
				addrs := contactJs.Get("PublicKey").String()
				avatar := contactJs.Get("Avatar").String()
				contact := Contact{
					PublicKey:        addrs,
					AccountPublicKey: s.Account().Addr,
					Avatar:           []byte(avatar),
				}
				contacts = append(contacts, contact)
			}
		}
		<-waitCh
	}()
	return contactsCh
}

func (s *service) Messages(contactAddr string, offset int, limit int) <-chan []Message {
	messagesChan := make(chan []Message, 1)
	messages := make([]Message, 0)
	if s.Account().Addr == "" {
		messagesChan <- messages
		close(messagesChan)
		return messagesChan
	}
	go func() {
		defer func() { recoverPanicCloseCh(messagesChan, messages, alog.Logger()) }()
		txn := s.IndexedDB().Call("transaction", js.ValueOf(DBPathMessagesDir), "readwrite")
		objectStore := txn.Call("objectStore", js.ValueOf(DBPathMessagesDir))
		userAddr := s.Account().Addr
		req := objectStore.Call("get", []interface{}{userAddr, contactAddr})
		waitCh := make(chan error, 1)
		req.Set("onsuccess", OnSuccess(waitCh))
		req.Set("onerror", OnError(waitCh))
		err := <-waitCh
		if err != nil {
			alog.Logger().Errorln(err)
			return
		}
		messagesJS := req.Get("target").Get("result")
		if !messagesJS.Truthy() {
			return
		}
		messagesArr := messagesJS.Get("Messages")
		if !messagesArr.Truthy() {
			return
		}
		sizeOFArray := messagesArr.Get("length").Int()
		if offset > sizeOFArray {
			return
		}
		for i := offset; i < sizeOFArray && i < offset+limit; i++ {
			msg := messagesArr.Index(i)
			msgGo := Message{
				AccountPublicKey: userAddr,
				ContactPublicKey: contactAddr,
				From:             msg.Get("From").String(),
				To:               msg.Get("To").String(),
				Created:          msg.Get("Created").String(),
				Text:             msg.Get("Text").String(),
				Key:              msg.Get("Key").String(),
			}
			messages = append(messages, msgGo)
		}
	}()
	return messagesChan
}

func (s *service) saveMessage(message Message) <-chan error {
	errCh := make(chan error, 1)
	var err error
	if s.Account().Addr == "" {
		errCh <- errors.New("current id is nil")
		close(errCh)
		return errCh
	}
	go func() {
		defer func() { recoverPanicCloseCh(errCh, err, alog.Logger()) }()
		txn := s.IndexedDB().Call("transaction", js.ValueOf(DBPathMessagesDir), "readwrite")
		objectStore := txn.Call("objectStore", js.ValueOf(DBPathMessagesDir))
		userAddr := s.Account().Addr
		req := objectStore.Call("get", []interface{}{userAddr, message.ContactPublicKey})
		waitCh := make(chan error, 2)
		req.Set("onsuccess", OnSuccess(waitCh))
		req.Set("onerror", OnError(waitCh))
		err = <-waitCh
		if err != nil {
			alog.Logger().Errorln(err)
			return
		}
		messagesJS := req.Get("target").Get("result")
		if !messagesJS.Truthy() {
			return
		}
		messagesArr := messagesJS.Get("Messages")
		if !messagesArr.Truthy() {
			return
		}
		req = objectStore.Call("put", js.ValueOf(map[string]interface{}{
			"AccountPublicKey": message.AccountPublicKey,
			"ContactPublicKey": message.ContactPublicKey,
			"From":             message.From,
			"To":               message.To,
			"Messages":         messagesArr,
		}))
		req.Set("onsuccess", OnSuccess(waitCh))
		req.Set("onerror", OnError(waitCh))
		err = <-waitCh
	}()
	return errCh
}

//func (s *service) saveContact(contactAddr *PublicKey) <-chan error {
//	errCh := make(chan error, 1)
//	var err error
//	if s.Account().PublicKey == "" {
//		err = errors.New("cannot save Contact, current account is empty")
//		recoverPanicCloseCh(errCh, err, alog.Logger())
//		return errCh
//	}
//
//	go func() {
//		defer func() { recoverPanicCloseCh(errCh, err, alog.Logger()) }()
//		txn := s.IndexedDB().Call("transaction", js.ValueOf(DBPathContactsDir), "readwrite")
//		objectStore := txn.Call("objectStore", js.ValueOf(DBPathContactsDir))
//		userAddr := s.Account().PublicKey
//		req := objectStore.Call("get", userAddr)
//		errCh2 := make(chan error, 1)
//		req.Set("onsuccess", js.FuncOf(func(this js.Value, args []js.Value) interface{} {
//			contactsJs := args[0].Get("target").Get("result")
//			if contactsJs.Truthy() {
//				contactsJsArr := contactsJs.Get("Contacts")
//				sizeOFArray := contactsJsArr.Get("length").Int()
//				for i := 0; i < sizeOFArray; i++ {
//					contactJs := contactsJsArr.Index(i)
//					if contactJs.String() == contactAddr.String() {
//						err = errors.New("contact already exist")
//						recoverPanicCloseCh(errCh2, err, alog.Logger())
//						return nil
//					}
//				}
//				// if Contact is not found, then add new Contact
//				contactsJsArr.Call("push", contactAddr.String())
//				newReq := objectStore.Call("put", map[string]interface{}{
//					"Contacts": contactsJsArr,
//					"PublicKey":     userAddr,
//				})
//				newReq.Set("onsuccess", js.FuncOf(func(this js.Value, args []js.Value) interface{} {
//					defer func() { recoverPanicCloseCh(errCh2, err, alog.Logger()) }()
//					err = nil
//					return nil
//				}))
//				newReq.Set("onerror", js.FuncOf(func(this js.Value, args []js.Value) interface{} {
//					defer func() { recoverPanicCloseCh(errCh2, err, alog.Logger()) }()
//					errorJs := args[0].Get("target").Get("errorCode")
//					errStr := fmt.Sprintf("error saving Contact, errCode is %s", errorJs.String())
//					err = errors.New(errStr)
//					alog.Logger().Errorln(errStr)
//					return nil
//				}))
//			} else {
//				newArrJs := js.ValueOf([]interface{}{})
//				newArrJs.Call("push", contactAddr.String())
//				newReq := objectStore.Call("put", map[string]interface{}{
//					"Contacts": newArrJs,
//					"Id":       userAddr,
//				})
//				newReq.Set("onsuccess", js.FuncOf(func(this js.Value, args []js.Value) interface{} {
//					defer func() { recoverPanicCloseCh(errCh2, err, alog.Logger()) }()
//					err = nil
//					return nil
//				}))
//				newReq.Set("onerror", js.FuncOf(func(this js.Value, args []js.Value) interface{} {
//					defer func() { recoverPanicCloseCh(errCh2, err, alog.Logger()) }()
//					errorJs := args[0].Get("target").Get("errorCode")
//					errStr := fmt.Sprintf("error saving Contact, errCode is %s", errorJs.String())
//					err = errors.New(errStr)
//					alog.Logger().Errorln(errStr)
//					return nil
//				}))
//			}
//			return nil
//		}))
//		req.Set("onerror", js.FuncOf(func(this js.Value, args []js.Value) interface{} {
//			defer func() { recoverPanicCloseCh(errCh2, err, alog.Logger()) }()
//			errorJs := args[0].Get("target").Get("errorCode")
//			errStr := fmt.Sprintf("error saving Contact, errCode is %s", errorJs.String())
//			err = errors.New(errStr)
//			alog.Logger().Errorln(errStr)
//			return nil
//		}))
//		<-errCh2
//	}()
//	return errCh
//}

func (s *service) SetAsCurrentAccount(addrStr string) <-chan error {
	errCh := make(chan error, 1)
	if !s.AccountKeyExists(addrStr) {
		errCh <- errors.New("could not find the account")
		close(errCh)
		return errCh
	}
	go func() {
		var err error
		defer func() {
			recoverPanicCloseCh(errCh, err, alog.Logger())
		}()
		account := Account{Addr: addrStr}
		for _, acc := range s.Accounts() {
			if acc.Addr == addrStr {
				account = acc
				break
			}
		}
		err = <-s.saveAccountToDB(account)
	}()
	return errCh
}

func (s *service) DeleteAccounts(accounts []Account) <-chan error {
	errCh := make(chan error, 1)
	if len(accounts) == 0 {
		errCh <- errors.New("accounts is empty")
		close(errCh)
		return errCh
	}
	go func() {
		var err error
		defer func() {
			recoverPanicCloseCh(errCh, err, alog.Logger())
		}()
		var currentAccountChange bool
		//for _, eachAccount := range accounts {
		//	if s.Account().PublicKey == eachAccount.PublicKey {
		//		s.GormDB().Delete(&PrimaryAccount{AlwaysOne: 1, AccountPublicKey: eachAccount.PublicKey})
		//		currentAccountChange = true
		//	}
		//	s.GormDB().Delete(&eachAccount)
		//	s.GormDB().Where("account_addr = ?", eachAccount.PublicKey).Delete(&Contact{})
		//	s.GormDB().Where("account_addr = ?", eachAccount.PublicKey).Delete(&Message{})
		//}
		if currentAccountChange {
			acc := <-s.reqAccount()
			s.setAccount(acc)
		}
		accs := <-s.reqAccountsFromDB()
		s.setAccounts(accs)
		select {
		case s.eventChan <- struct{}{}:
		default:
		}
	}()
	return errCh
}

func (s *service) DeleteContacts(contacts []Contact) <-chan error {
	errCh := make(chan error, 1)
	if len(contacts) == 0 {
		errCh <- errors.New("contacts is empty")
		close(errCh)
		return errCh
	}
	go func() {
		var err error
		defer func() {
			recoverPanicCloseCh(errCh, err, alog.Logger())
		}()
		//for _, eachContact := range contacts {
		//	s.GormDB().Where("account_addr = ? and addr = ?", eachContact.AccountPublicKey, eachContact.PublicKey).Delete(&Contact{})
		//	s.GormDB().Where("account_addr = ? and contact_addr = ?", eachContact.AccountPublicKey, eachContact.PublicKey).Delete(&Message{})
		//}
		contacts := <-s.Contacts()
		s.setContacts(contacts)
		select {
		case s.eventChan <- struct{}{}:
		default:
		}
	}()
	return errCh
}

func (s *service) LastMessage(contactAddr string) <-chan Message {
	msg := Message{}
	msgCh := make(chan Message, 1)
	go func() {
		defer func() {
			recoverPanicCloseCh(msgCh, msg, alog.Logger())
		}()
		//txn := s.GormDB().Model(&Message{}).Order("created desc").First(&msg,
		//	"account_addr = ? and contact_addr = ?", s.Account().PublicKey, contactAddr)
		//if txn.Error != nil {
		//	alog.Logger().Println(txn.Error)
		//}
	}()
	return msgCh
}

func (s *service) UnreadMessagesCount(contactAddr string) <-chan int64 {
	countCh := make(chan int64, 1)
	go func() {
		count := int64(0)
		defer func() {
			recoverPanicCloseCh(countCh, count, alog.Logger())
		}()
		if s.Account().Addr == "" {
			return
		}
		//txn := s.GormDB().Model(&Message{}).Where(map[string]interface{}{
		//	"From": contactAddr,
		//	"To":   s.Account().PublicKey,
		//	"Read": false,
		//}).Count(&count)
		//if txn.Error != nil {
		//	alog.Logger().Println(txn.Error)
		//}
	}()
	return countCh
}

func (s *service) MessagesCount(contactAddr string) <-chan int64 {
	countCh := make(chan int64, 1)
	go func() {
		count := int64(0)
		defer func() {
			recoverPanicCloseCh(countCh, count, alog.Logger())
		}()
		if s.Account().Addr == "" {
			return
		}
		//txn := s.GormDB().Model(&Message{}).Where(map[string]interface{}{
		//	"contact_addr": contactAddr,
		//	"account_addr": s.Account().PublicKey,
		//}).Count(&count)
		//if txn.Error != nil {
		//	alog.Logger().Println(txn.Error)
		//}
	}()
	return countCh
}

func (s *service) MarkPrevMessagesAsRead(contactAddr string) <-chan error {
	//created := time.Now().UTC().Format(time.RFC3339)
	errCh := make(chan error, 1)
	go func() {
		var err error
		defer func() {
			recoverPanicCloseCh(errCh, err, alog.Logger())
		}()
		if s.Account().Addr == "" {
			err = errors.New("account addr is empty")
			return
		}
		//txn := s.GormDB().Model(Message{}).Where(
		//	"`from` = ? and `to` = ? and created < ?",
		//	contactAddr, s.Account().PublicKey, created,
		//).Updates(Message{Read: true})
		//if txn.Error != nil {
		//	err = txn.Error
		//	alog.Logger().Println(txn.Error)
		//}
		select {
		case s.eventChan <- struct{}{}:
		default:
		}
	}()
	return errCh
}

func OnSuccess[t interface{}](goChan chan<- t) js.Func {
	var cb js.Func
	var val t
	cb = js.FuncOf(func(this js.Value, args []js.Value) any {
		goChan <- val
		cb.Release() // release the function
		return nil
	})
	return cb
}
func OnError(goChan chan<- error) js.Func {
	var cb js.Func
	cb = js.FuncOf(func(this js.Value, args []js.Value) any {
		errorJs := args[0].Get("target").Get("errorCode")
		goChan <- errors.New(errorJs.String())
		cb.Release() // release the function
		return nil
	})
	return cb
}
