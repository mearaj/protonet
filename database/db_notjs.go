// +build !js

package database

func (db *Database) open()  {
	db.isDatabaseReady = true
}