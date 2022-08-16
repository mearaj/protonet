package service

const (
	// DBPathCfgDir app directory (indexedDB, gormDB)
	DBPathCfgDir = "protonet.x"

	// DBPathFileName database file name (gormDB)
	DBPathFileName = "protonet.db"

	// DBPathAccountsDir contains all the user accounts (indexedDB)
	DBPathAccountsDir = "accounts"

	// DBPathCurrentAccountDir contains the current Account in the ui (indexedDB)
	DBPathCurrentAccountDir = "account"

	// DBPathContactsDir contains Account directory to which it belongs to and
	// then all the contacts inside that directory (indexedDB)
	DBPathContactsDir = "contacts"

	// DBPathMessagesDir contains Account directory which in turn
	// contains Contact directory to which it belongs to and
	// then all the messages inside that directory (indexedDB)
	DBPathMessagesDir = "messages"
)

const AppIndexedDBVersion int64 = 3
