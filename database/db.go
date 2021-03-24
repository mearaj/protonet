package database

import "time"


type Database struct {
	// Refers to indexeddb of type js.Value
	indexedDb interface{}
	// db refers to database created from indexedDb of type js.Value
	db interface{}

	isDatabaseReady bool
}

func NewDatabase() (*Database, <-chan struct{}){
	db := &Database{}
	initChan := make(chan struct{})
	db.open()
	go func() {
		for {
			time.Sleep(time.Second / 10)
			switch db.isDatabaseReady {
			case true:
				initChan <- struct{}{}
				return
			case false:
			}
		}
	}()
	return db,initChan
}

func (db *Database) IsDatabaseReady()  bool {
	return db.isDatabaseReady
}
