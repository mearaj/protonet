package model

import "errors"

var ErrInvalidAccount = errors.New("invalid account")
var ErrInvalidMessage = errors.New("invalid message")
var ErrInvalidContact = errors.New("invalid contact")

const KeySeparator = "[]"
const KeyPrefixAccounts = "accounts"
const KeyPrefixMessages = "messages"
const KeyPrefixContacts = "contacts"
