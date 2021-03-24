// +build !js

package database

import (
	"path/filepath"
	log "github.com/sirupsen/logrus"
)

func (db *Database) LoadAccountsFromDisk() <-chan Accounts {
	accounts := Accounts{}
	accountsChan := make(chan Accounts,1)
	go func() {
		allPaths, err := GetFileInfosFromDir(AccountsDir)
		if err != nil {
			log.Println("error in LoadAccountsFromDisk, err:", err)
		}

		for _, eachPath := range allPaths {
			if eachPath.IsDir() {
				subDirPath := filepath.Join(AccountsDir, eachPath.Name())
				if _, ok := accounts[eachPath.Name()]; !ok {
					account := &Account{}
					err = LoadStructFromFile(subDirPath, AccountsFile, account)
					if err != nil {
						log.Println(err)
						continue
					}
					accounts[account.ID] = account
				}
			}
		}
		accountsChan <- accounts
	}()
	return accountsChan
}

func (db *Database) SaveAccountsToDisk(accounts Accounts) {
	for _, acc := range accounts {
		db.SaveAccountToDisk(acc)
	}
}
func (db *Database) SaveAccountToDisk(acc *Account) error {
	dirPath := filepath.Join(AccountsDir, acc.ID)
	return SaveStructToFile(dirPath, AccountsFile, acc)
}

func (db *Database) SaveUserToDisk(user *Account) error {
	return SaveStructToFile(AccountDir, AccountFile, user)
}

func (db *Database) LoadUserFromDisk() <-chan *Account {
	userChan := make(chan *Account,1)
	go func() {
		user := &Account{}
		err := LoadStructFromFile(AccountDir, AccountFile, user)
		if err != nil {
			log.Println(err)
		}
		userChan <- user
	}()
	return userChan
}
