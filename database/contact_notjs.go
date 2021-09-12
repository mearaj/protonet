// +build !js

package database

import (
	log "github.com/sirupsen/logrus"
	"path/filepath"
)

func (db *Database) LoadContactsFromDisk(userID string)  <-chan Contacts {
	contactsChan := make(chan Contacts,1)
	go func() {
		contacts := Contacts{}
		dirPath := filepath.Join(ContactsDir, userID)
		allPaths, err := GetFileInfosFromDir(dirPath)
		if err != nil {
			log.Println("error in LoadContactsFromDisk, err:", err)
		}
		for _, eachPath := range allPaths {
			if eachPath.IsDir() {
				eachContact := &Contact{}
				contactPath := filepath.Join(dirPath, eachPath.Name())
				err = LoadStructFromFile(contactPath, ContactsFile, eachContact)
				if err != nil {
					log.Println("error in LoadContactsFromDisk, err:", err)
					continue
				}
				contacts[eachContact.IDStr] = eachContact
			}
		}
		contactsChan<- contacts
	}()
	return contactsChan
}

func (db *Database) SaveAllContactsToDisk(userID string, contacts Contacts) {
	for _, eachContact := range contacts {
		db.SaveContactToDisk(userID, eachContact)
	}
}

func (db *Database) SaveContactToDisk(userID string, eachContact *Contact) {
	dirPath := filepath.Join(ContactsDir, userID, eachContact.IDStr)
	SaveStructToFile(dirPath, ContactsFile, eachContact)
}

func (db *Database) DeleteContact(userID string, contact *Contact)  {
	DeleteDirIfExist(filepath.Join(TextMessagesDir, userID, contact.IDStr))
	DeleteDirIfExist(filepath.Join(ContactsDir, userID))
}

func ContactsToArray(contacts Contacts) []*Contact {
	contactsArr := make([]*Contact, 0, len(contacts))
	for _, contact := range contacts {
		contactsArr = append(contactsArr, contact)
	}
	return contactsArr
}
