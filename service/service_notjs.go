//go:build !js

package service

import (
	"errors"
	"gioui.org/app"
	libcrypto "github.com/libp2p/go-libp2p-core/crypto"
	"github.com/mearaj/protonet/alog"
	log "github.com/sirupsen/logrus"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"os"
	"path/filepath"
	"time"
)

func (s *service) init() {
	err := <-s.openDatabase()
	if err != nil {
		log.Fatal(err)
	}
	accounts := <-s.reqAccountsFromDB()
	if len(accounts) > 0 {
		s.setAccount(accounts[0])
	}
	go s.runService()
}

func (s *service) GormDB() *gorm.DB {
	s.databaseMutex.RLock()
	defer s.databaseMutex.RUnlock()
	if s.database == (*gorm.DB)(nil) {
		return nil
	}
	if gormDB, ok := s.database.(*gorm.DB); ok {
		return gormDB
	}
	return nil
}

func (s *service) setGormDB(gormDB *gorm.DB) {
	s.databaseMutex.Lock()
	defer s.databaseMutex.Unlock()
	s.database = gormDB
}

func (s *service) SetUserPassword(passwd string) <-chan error {
	errCh := make(chan error, 1)
	if s.getUserPassword() != "" {
		errCh <- errors.New("password already exists")
		return errCh
	}
	if passwd == "" {
		errCh <- errors.New("password cannot be empty")
		return errCh
	}
	go func() {
		var err error
		defer func() {
			recoverPanicCloseCh(errCh, err, alog.Logger())
		}()
		accs := <-s.Accounts()
		if !(len(accs) > 0) {
			s.setUserPassword(passwd)
			return
		}
		_, err = accs[0].PrivateKey(passwd)
		if err != nil {
			return
		}
		s.setUserPassword(passwd)
		// Following code is for changing the password
		//for _, eachAccount := range accs {
		//	var pvtKey string
		//	pvtKey, err = eachAccount.PrivateKey(s.getUserPassword())
		//	if err != nil {
		//		return
		//	}
		//	var pvtKeyBs []byte
		//	pvtKeyBs, err = hex.DecodeString(pvtKey)
		//	if err != nil {
		//		return
		//	}
		//	var pvtKeyEncBs []byte
		//	pvtKeyEncBs, err = Encrypt([]byte(passwd), pvtKeyBs)
		//	if err != nil {
		//		return
		//	}
		//	pvtKeyEnc := hex.EncodeToString(pvtKeyEncBs)
		//	eachAccount.PrivateKeyEnc = pvtKeyEnc
		//}
		//for _, eachAccount := range accs {
		//	s.GormDB().Scan(&eachAccount)
		//}

	}()
	return errCh
}

func (s *service) setUserPassword(passwd string) {
	s.userPasswordMutex.Lock()
	defer s.userPasswordMutex.Unlock()
	s.userPassword = passwd
}

func (s *service) getUserPassword() string {
	s.userPasswordMutex.RLock()
	defer s.userPasswordMutex.RUnlock()
	return s.userPassword
}

// UserPasswordExist if the database is not created, it indicates user account hasn't been set up yet
func (s *service) UserPasswordSet() bool {
	s.userPasswordMutex.RLock()
	defer s.userPasswordMutex.RUnlock()
	return s.getUserPassword() != ""
	//dirPath, err := app.DataDir()
	//if err != nil {
	//	alog.Logger().Errorln(err)
	//	return false
	//}
	//dirPath = filepath.Join(dirPath, DBPathCfgDir)
	//if _, err = os.Stat(dirPath); os.IsNotExist(err) {
	//	alog.Logger().Errorln(err)
	//	return false
	//}
	//dbFullName := filepath.Join(dirPath, DBPathFileName)
	//if _, err = os.Stat(dbFullName); os.IsNotExist(err) {
	//	alog.Logger().Errorln(err)
	//	return false
	//}
	//return true
}

func (s *service) openDatabase() <-chan error {
	errCh := make(chan error, 1)
	go func() {
		var err error
		defer func() {
			if err != nil {
				alog.Logger().Errorln(err)
			}
			recoverPanicCloseCh(errCh, err, alog.Logger())
		}()
		dirPath, err := app.DataDir()
		if err != nil {
			return
		}
		dirPath = filepath.Join(dirPath, DBPathCfgDir)
		if _, err = os.Stat(dirPath); os.IsNotExist(err) {
			err = os.MkdirAll(dirPath, 0700)
			if err != nil {
				return
			}
		}
		dbFullName := filepath.Join(dirPath, DBPathFileName)
		if _, err = os.Stat(dbFullName); os.IsNotExist(err) {
			var file *os.File
			file, err = os.OpenFile(
				dbFullName,
				os.O_CREATE|os.O_APPEND|os.O_RDWR,
				0700,
			)
			if err != nil {
				return
			}
			_ = file.Close()
		}
		gormDB, err := gorm.Open(sqlite.Open(dbFullName), &gorm.Config{})
		if err != nil {
			return
		}
		s.setGormDB(gormDB)
		err = s.GormDB().AutoMigrate(&Account{}, &Contact{}, &Message{})
		if err != nil {
			return
		}
	}()
	return errCh
}

// saveAccountToDB saves as primary account
func (s *service) saveAccountToDB(acc Account) <-chan error {
	errCh := make(chan error, 1)
	var err error
	go func() {
		defer func() { recoverPanicCloseCh(errCh, err, alog.Logger()) }()
		txn := s.GormDB().Save(&acc)
		if txn.Error != nil {
			err = txn.Error
			alog.Logger().Errorln(err)
		}
	}()
	return errCh
}

func (s *service) reqAccountsFromDB() <-chan []Account {
	accountsCh := make(chan []Account, 1)
	go func() {
		accounts := make([]Account, 0)
		defer func() {
			recoverPanicCloseCh(accountsCh, accounts, alog.Logger())
		}()
		if s.GormDB() == nil {
			return
		}
		result := s.GormDB().Model(&Account{}).Order("updated_at desc").Find(&accounts)
		if result.Error != nil {
			alog.Logger().Errorln(result.Error)
		}
	}()
	return accountsCh
}

func (s *service) Accounts() <-chan []Account {
	accountsCh := make(chan []Account, 1)
	accounts := make([]Account, 0)
	go func() {
		defer func() {
			recoverPanicCloseCh(accountsCh, accounts, alog.Logger())
		}()
		txn := s.GormDB().Order("updated_at desc").Find(&accounts)
		if txn.Error != nil {
			alog.Logger().Errorln(txn.Error)
		}
	}()
	return accountsCh
}

func (s *service) AccountKeyExists(publicKey string) <-chan bool {
	existsCh := make(chan bool, 1)
	go func() {
		exists := false
		defer func() {
			recoverPanicCloseCh(existsCh, exists, alog.Logger())
		}()
		txn := s.GormDB().Find(&Account{PublicKey: publicKey}, "public_key = ?", publicKey)
		if txn.Error != nil {
			alog.Logger().Errorln(txn.Error)
		}
		exists = txn.Error == nil && txn.RowsAffected == 1
	}()
	return existsCh
}

func (s *service) Contacts(accountPublicKey string, offset, limit int) <-chan []Contact {
	contactsCh := make(chan []Contact, 1)
	contacts := make([]Contact, 0)
	go func() {
		defer func() {
			recoverPanicCloseCh(contactsCh, contacts, alog.Logger())
		}()
		txn := s.GormDB().Order("updated_at desc").Offset(offset).Limit(limit).Find(&contacts, "account_public_key = ?", accountPublicKey)
		if txn.Error != nil {
			alog.Logger().Errorln(txn.Error)
		}
	}()
	return contactsCh
}

func (s *service) Messages(contactKey string, offset int, limit int) <-chan []Message {
	messagesChan := make(chan []Message, 1)
	messages := make([]Message, 0)
	accountKey := s.Account().PublicKey
	go func() {
		defer func() {
			recoverPanicCloseCh(messagesChan, messages, alog.Logger())
		}()
		txn := s.GormDB().Order("created desc").Offset(offset).Limit(limit).Find(&messages, "account_public_key = ? and contact_public_key = ?", accountKey, contactKey)
		if txn.Error != nil {
			alog.Logger().Errorln(txn)
		}
	}()
	return messagesChan
}

// saveMessage saves message to the database
func (s *service) saveMessage(msg Message) <-chan error {
	errCh := make(chan error, 1)
	go func() {
		var err error
		defer func() {
			recoverPanicCloseCh(errCh, err, alog.Logger())
		}()
		foundMessage := Message{}
		// first find the same message that needs to be replaced, hash should be empty
		queryString := "id = ? and " +
			"account_public_key = ? and contact_public_key = ?"
		txn := s.GormDB().Take(
			&foundMessage,
			queryString,
			msg.ID,
			msg.AccountPublicKey,
			msg.ContactPublicKey,
		)
		if txn.Error != nil {
			alog.Logger().Println(txn.Error)
		}
		var isDuplicate bool
		// If message is found, then it needs to be replaced to avoid duplicated messages
		if txn.RowsAffected > 0 {
			foundMessage.Created = msg.Created
			txn = s.GormDB().Save(&foundMessage)
			isDuplicate = true
		} else {
			txn = s.GormDB().Create(&msg)
		}
		if txn.Error != nil {
			err = txn.Error
			alog.Logger().Errorln(txn.Error)
		}
		// After saving/updating new message, we update contact to update contact's Updated_At
		if !isDuplicate {
			ct := Contact{}
			txn = s.GormDB().Take(&ct, "account_public_key = ? and public_key = ?", msg.AccountPublicKey, msg.ContactPublicKey)
			if txn.Error != nil {
				err = txn.Error
				alog.Logger().Errorln(txn)
			}
			if txn.RowsAffected == 1 {
				lastMessage := <-s.LastMessage(msg.ContactPublicKey)
				timeVal, _ := time.Parse(time.RFC3339, lastMessage.Created)
				if timeVal.After(ct.UpdatedAt) {
					ct.UpdatedAt = timeVal
					txn.Save(&ct)
				}
			}
		}
		eventData := MessagesCountChangedEventData{
			AccountPublicKey: msg.AccountPublicKey,
			ContactPublicKey: msg.ContactPublicKey,
		}
		event := Event{Data: eventData, Topic: MessagesCountChangedEventTopic}
		s.eventBroker.Fire(event)
	}()
	return errCh
}

func (s *service) SaveContact(publicKey string) <-chan error {
	errCh := make(chan error, 1)
	a := s.Account()
	if a.PublicKey == "" {
		errCh <- errors.New("current id is nil")
		close(errCh)
		return errCh
	}
	_, err := GetPublicKeyFromStr(publicKey, libcrypto.Secp256k1)
	if err != nil {
		errCh <- err
		close(errCh)
		return errCh
	}
	go func() {
		var err error
		defer func() {
			recoverPanicCloseCh(errCh, err, alog.Logger())
		}()
		var txn *gorm.DB
		ct := Contact{
			PublicKey:        publicKey,
			AccountPublicKey: a.PublicKey,
		}
		txn = s.GormDB().Save(&ct)
		if txn.Error != nil {
			err = txn.Error
			alog.Logger().Errorln(txn)
			return
		}
		eventData := ContactsChangeEventData{
			AccountPublicKey: a.PublicKey,
		}
		event := Event{Data: eventData, Topic: ContactsChangedEventTopic}
		s.eventBroker.Fire(event)
	}()
	return errCh
}

func (s *service) SetAsCurrentAccount(account Account) <-chan error {
	errCh := make(chan error, 1)
	if exists := <-s.AccountKeyExists(account.PublicKey); !exists {
		errCh <- errors.New("could not find the account")
		close(errCh)
		return errCh
	}
	go func() {
		var err error
		defer func() {
			recoverPanicCloseCh(errCh, err, alog.Logger())
		}()
		txn := s.GormDB().Find(&account, "public_key = ?", account.PublicKey)
		if txn.Error != nil {
			err = txn.Error
		}
		s.setAccount(account)
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
		acc := s.Account()
		var currentAccountDeleted bool
		for _, eachAccount := range accounts {
			if acc.PublicKey == eachAccount.PublicKey {
				currentAccountDeleted = true
			}
			s.GormDB().Where("public_key = ?", eachAccount.PublicKey).Delete(&eachAccount)
			s.GormDB().Where("account_public_key = ?", eachAccount.PublicKey).Delete(&Contact{})
			s.GormDB().Where("account_public_key = ?", eachAccount.PublicKey).Delete(&Message{})
		}
		accs := <-s.reqAccountsFromDB()
		if currentAccountDeleted {
			if len(accs) > 0 {
				s.setAccount(accs[0])
			} else {
				s.setAccount(Account{})
			}
			s.eventBroker.Fire(Event{
				Data:  ContactsChangeEventData{},
				Topic: ContactsChangedEventTopic,
			})
		}
		s.eventBroker.Fire(Event{
			Data:  AccountsChangedEventData{},
			Topic: AccountsChangedEventTopic,
		})
	}()
	return errCh
}

func (s *service) AccountsCount() <-chan int64 {
	countCh := make(chan int64, 1)
	go func() {
		count := int64(0)
		defer func() {
			recoverPanicCloseCh(countCh, count, alog.Logger())
		}()
		if s.GormDB() == nil {
			return
		}
		txn := s.GormDB().Model(&Account{}).Count(&count)
		if txn.Error != nil {
			alog.Logger().Println(txn.Error)
		}
	}()
	return countCh
}

func (s *service) ContactsCount(accountPublicKey string) <-chan int64 {
	countCh := make(chan int64, 1)
	go func() {
		count := int64(0)
		defer func() {
			recoverPanicCloseCh(countCh, count, alog.Logger())
		}()
		if s.GormDB() == nil {
			return
		}
		txn := s.GormDB().Model(&Contact{}).Where(map[string]interface{}{
			"account_public_key": accountPublicKey,
		}).Count(&count)
		if txn.Error != nil {
			alog.Logger().Println(txn.Error)
		}
	}()
	return countCh
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
		for _, eachContact := range contacts {
			s.GormDB().Where("account_public_key = ? and public_key = ?", eachContact.AccountPublicKey, eachContact.PublicKey).Delete(&Contact{})
			s.GormDB().Where("account_public_key = ? and contact_public_key = ?", eachContact.AccountPublicKey, eachContact.PublicKey).Delete(&Message{})
		}
		s.eventBroker.Fire(Event{
			Data: ContactsChangeEventData{
				AccountPublicKey: s.Account().PublicKey,
			},
			Topic: ContactsChangedEventTopic,
		})
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
		a := s.Account()
		txn := s.GormDB().Order("created desc").First(&msg,
			"account_public_key = ? and contact_public_key = ?", a.PublicKey, contactAddr)
		if txn.Error != nil {
			alog.Logger().Println(txn.Error)
		}
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
		a := s.Account()
		if a.PublicKey == "" {
			return
		}
		txn := s.GormDB().Model(&Message{}).Where(map[string]interface{}{
			"account_public_key": a.PublicKey,
			"contact_public_key": contactAddr,
			"from":               contactAddr,
			"to":                 a.PublicKey,
			"read":               false,
		}).Count(&count)
		if txn.Error != nil {
			alog.Logger().Println(txn.Error)
		}
	}()
	return countCh
}

func (s *service) MessagesCount(contactPublicKey string) <-chan int64 {
	countCh := make(chan int64, 1)
	go func() {
		count := int64(0)
		defer func() {
			recoverPanicCloseCh(countCh, count, alog.Logger())
		}()
		a := s.Account()
		if a.PublicKey == "" {
			return
		}
		txn := s.GormDB().Model(&Message{}).Where(map[string]interface{}{
			"account_public_key": a.PublicKey,
			"contact_public_key": contactPublicKey,
		}).Count(&count)
		if txn.Error != nil {
			alog.Logger().Println(txn.Error)
		}
	}()
	return countCh
}

func (s *service) MarkPrevMessagesAsRead(publicKey string) <-chan error {
	created := time.Now().UTC().Format(time.RFC3339)
	errCh := make(chan error, 1)
	go func() {
		var err error
		defer func() {
			recoverPanicCloseCh(errCh, err, alog.Logger())
		}()
		a := s.Account()
		if a.PublicKey == "" {
			err = errors.New("account addr is empty")
			return
		}
		txn := s.GormDB().Model(Message{}).Where(
			"`from` = ? and `to` = ? and created < ?",
			publicKey, a.PublicKey, created,
		).Updates(Message{Read: true})
		if txn.Error != nil {
			err = txn.Error
			alog.Logger().Println(txn.Error)
		}
		//eventData := MessagesStateChangedEventData{
		//	AccountPublicKey: a.PublicKey,
		//	ContactPublicKey: publicKey,
		//}
		//event := Event{Data: eventData, Topic: MessagesStateChangedEventTopic}
		//s.eventBroker.Fire(event)
	}()
	return errCh
}
