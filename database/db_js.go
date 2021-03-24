package database

import (
	log "github.com/sirupsen/logrus"
	"syscall/js"
)

var version int = 1

func (db *Database) open() {
	db.indexedDb = js.Global().Get("indexedDB")
	indexedDb := db.indexedDb.(js.Value)
	req := indexedDb.Call("open", js.ValueOf(AppDir), js.ValueOf(version))
	req.Set("onsuccess", js.FuncOf(db.addEventSuccess))
	req.Set("onupgradeneeded", js.FuncOf(db.addEventUpgradeNeeded))
}

func (db *Database) addEventSuccess(this js.Value, args []js.Value) interface{} {
	db.db = args[0].Get("target").Get("result")
	log.Println("event success called")
	db.isDatabaseReady = true
	return nil
}

func (db *Database) addEventUpgradeNeeded(this js.Value, args []js.Value) interface{} {
	log.Println("addEventUpgradeNeeded called")
	if db.db == nil {
		db.db = args[0].Get("target").Get("result")
	}
	idb := db.db.(js.Value)
	keypath := js.ValueOf(map[string]interface{}{
		"keyPath": "ID",
	})

	idb.Call("createObjectStore", js.ValueOf(AccountsDir), keypath)
	idb.Call("createObjectStore", js.ValueOf(AccountDir), keypath)
	idb.Call("createObjectStore", js.ValueOf(ContactsDir), keypath)
	idb.Call("createObjectStore", js.ValueOf(TextMessagesDir), keypath)
	return nil
}
