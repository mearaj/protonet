//go:build !js
// +build !js

package database

import (
	log "github.com/sirupsen/logrus"
	"path"
	"path/filepath"
	"time"
)

func (db *Database) LoadTxtMsgsFromDisk(userID string, contactID string) <-chan TxtMsgs {
	txtMsgsChan := make(chan TxtMsgs)
	messages := make(TxtMsgs)
	go func() {
		dirPath := filepath.Join(TextMessagesDir, userID,
			contactID)
		allPaths, err := GetFileInfosFromDir(dirPath)
		if err != nil {
			log.Println("error in SaveUpdateTxtMsgToDisk, err:", err)
		}

		for _, eachPath := range allPaths {
			if eachPath.IsDir() {
				subDirPath := filepath.Join(dirPath, eachPath.Name())
				subPaths, err := GetFileInfosFromDir(subDirPath)
				for _, eachSubPath := range subPaths {
					message := &TxtMsg{}
					err = LoadStructFromFile(subDirPath, eachSubPath.Name(), message)
					if err != nil {
						log.Println(err)
						continue
					}
					messages[message.ID] = message
				}
			}
		}
		select {
		case txtMsgsChan <- messages:
		default:
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
	dirPath := filepath.Join(TextMessagesDir, userID, contactID)

	subPath := path.Join(dirPath, time.Unix(message.Timestamp, 0).Format("2006Jan2"))
	SaveStructToFile(subPath, message.ID, message)
}
