package db

import (
	"errors"
)

const KeySeparator = "[]"
const KeyPrefixAccounts = "accounts"
const KeyPrefixMessages = "messages"
const KeyPrefixContacts = "contacts"

var ErrInvalidKey = errors.New("invalid key")
var ErrInvalidAccount = errors.New("invalid account")
var ErrInvalidMessage = errors.New("invalid message")
var ErrInvalidContact = errors.New("invalid contact")
var ErrAccountDoesNotExist = errors.New("account does not exists")
var ErrPasswdNotSet = errors.New("password is not set")
var ErrPasswdAlreadyExist = errors.New("password already exist")
var ErrPasswdCannotBeEmpty = errors.New("password cannot be empty")

const MaxNumOfPasswdChars = 32
const passwdPadCharacter = "0"
