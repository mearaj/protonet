package database

import (
	"syscall/js"
)

func (db *Database) LoadAccountsFromDisk() <-chan Accounts {
	accountsChan := make(chan Accounts, 1)
	go func() {
		accounts := Accounts{}
		if db.db == nil {
			accountsChan <- accounts
			return
		}
		dbi := db.db.(js.Value)
		if !dbi.IsUndefined() && !dbi.IsNull() {
			req := dbi.Call("transaction", js.ValueOf(AccountsDir),
			).Call("objectStore", js.ValueOf(AccountsDir),
			).Call("getAll")
			req.Set("onsuccess", js.FuncOf(func(this js.Value, args []js.Value) interface{} {
				accountsJS := args[0].Get("target").Get("result")
				sizeOfArray := accountsJS.Get("length").Int()
				for i := 0; i < sizeOfArray; i++ {
					accountJs := accountsJS.Index(i)
					account := Account{
						PvtKeyHex: accountJs.Get("PvtKeyHex").String(),
						PubKeyHex: accountJs.Get("PubKeyHex").String(),
						ID:        accountJs.Get("ID").String(),
						Name:      accountJs.Get("Name").String(),
						PvtImg:    []byte(accountJs.Get("PvtImg").String()),
						PubImg:    []byte(accountJs.Get("PubImg").String()),
					}
					accounts[account.ID] = &account
				}
				accountsChan <- accounts
				return nil
			}))
		} else {
			accountsChan <- accounts
		}
	}()
	return accountsChan
}

func (db *Database) SaveAccountsToDisk(accounts Accounts) {
	if db.db == nil {
		return
	}
	dbi := db.db.(js.Value)
	if !dbi.IsUndefined() && !dbi.IsNull() && len(accounts) != 0 {
		txn := dbi.Call("transaction", js.ValueOf(AccountsDir), js.ValueOf("readwrite"))
		store := txn.Call("objectStore", js.ValueOf(AccountsDir))
		for _, account := range accounts {
			val := js.ValueOf(map[string]interface{}{
				"PvtKeyHex": account.PvtKeyHex,
				"PubKeyHex": account.PubKeyHex,
				"ID":        account.ID,
				"Name":      account.Name,
				"PvtImg":    string(account.PvtImg),
				"PubImg":    string(account.PubImg),
			})
			store.Call("put", val)
		}
	}
}
func (db *Database) SaveAccountToDisk(account *Account) {
	if db.db == nil {
		return
	}
	dbi := db.db.(js.Value)
	if !dbi.IsUndefined() && !dbi.IsNull() {
		txn := dbi.Call("transaction", js.ValueOf(AccountsDir), js.ValueOf("readwrite"))
		store := txn.Call("objectStore", js.ValueOf(AccountsDir))
		val := js.ValueOf(map[string]interface{}{
			"PvtKeyHex": account.PvtKeyHex,
			"PubKeyHex": account.PubKeyHex,
			"ID":        account.ID,
			"Name":      account.Name,
			"PvtImg":    string(account.PvtImg),
			"PubImg":    string(account.PubImg),
		})
		store.Call("put", val)
	}
}

func (db *Database) SaveUserToDisk(user *Account) {
	if db.db == nil {
		return
	}
	dbi := db.db.(js.Value)
	if !dbi.IsUndefined() && !dbi.IsNull() && user != nil {
		txn := dbi.Call("transaction", js.ValueOf(AccountDir), js.ValueOf("readwrite"))
		store := txn.Call("objectStore", js.ValueOf(AccountDir)).Call("clear")
		store.Set("onsuccess", js.FuncOf(func(this js.Value, args []js.Value) interface{} {
			txn = dbi.Call("transaction", js.ValueOf(AccountDir), js.ValueOf("readwrite"))
			store = txn.Call("objectStore", js.ValueOf(AccountDir))
			userVal := map[string]interface{}{"PvtKeyHex": user.PvtKeyHex,
				"PubKeyHex": user.PubKeyHex,
				"ID":        user.ID,
				"Name":      user.Name,
				"PvtImg":    string(user.PvtImg),
				"PubImg":    string(user.PubImg),
			}
			store.Call("put", userVal)
			return nil
		}))
	}
}

func (db *Database) LoadUserFromDisk() <-chan *Account {
	user := &Account{}
	userChan := make(chan *Account)
	go func() {
		if db.db == nil {
			select {
			case userChan <- nil:
			default:
			}
			return
		}
		dbi := db.db.(js.Value)
		if !dbi.IsUndefined() && !dbi.IsNull() {
			req := dbi.Call("transaction", js.ValueOf(AccountDir),
			).Call("objectStore", js.ValueOf(AccountDir),
			).Call("getAll")
			req.Set("onsuccess", js.FuncOf(func(this js.Value, args []js.Value) interface{} {
				accountsJS := args[0].Get("target").Get("result")
				sizeOfArray := accountsJS.Get("length").Int()
				if sizeOfArray > 0 {
					accountJs := accountsJS.Index(0)
					user.PvtKeyHex = accountJs.Get("PvtKeyHex").String()
					user.PubKeyHex = accountJs.Get("PubKeyHex").String()
					user.ID = accountJs.Get("ID").String()
					user.Name = accountJs.Get("Name").String()
					user.PvtImg = []byte(accountJs.Get("PvtImg").String())
					user.PubImg = []byte(accountJs.Get("PubImg").String())
				}
				select {
				case userChan <- user:
				default:
				}
				return nil
			}))
		} else {
			select {
			case userChan <- user:
			default:
			}
		}
	}()
	return userChan
}
