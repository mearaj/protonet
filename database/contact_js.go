package database

import (
	"syscall/js"
)

func (db *Database) LoadContactsFromDisk(userID string) <-chan Contacts {
	contactsChan := make(chan Contacts)
	go func() {
		contacts := Contacts{}
		if db.db == nil {
			contactsChan <- contacts
			return
		}
		dbi := db.db.(js.Value)
		if !dbi.IsUndefined() && !dbi.IsNull() {
			req := dbi.Call("transaction", js.ValueOf(ContactsDir), js.ValueOf("readwrite")).Call("objectStore", js.ValueOf(ContactsDir)).Call("get", js.ValueOf(userID))
			req.Set("onsuccess", js.FuncOf(func(this js.Value, args []js.Value) interface{} {
				acc := args[0].Get("target").Get("result")
				if !acc.IsNull() && !acc.IsUndefined() {
					cts := acc.Get("contacts")
					keysJs := js.Global().Get("Object").Call("keys", cts)
					for j := 0; j < keysJs.Get("length").Int(); j++ {
						ct := cts.Get(keysJs.Index(j).String())
						contact := Contact{
							PublicKey: ct.Get("PubKeyHex").String(),
							IDStr:     ct.Get("ID").String(),
							Name:      ct.Get("Name").String(),
							Image:     []byte(ct.Get("PvtImg").String()),
						}
						contacts[contact.IDStr] = &contact
					}
				} else {
					req = dbi.Call("transaction", js.ValueOf(ContactsDir), js.ValueOf("readwrite")).Call("objectStore", js.ValueOf(ContactsDir))
					req.Call("put", js.ValueOf(map[string]interface{}{
						"ID":       userID,
						"contacts": js.ValueOf(map[string]interface{}{}),
					}))
				}
				contactsChan <- contacts
				return nil
			}))

		} else {
			contactsChan <- contacts
		}
	}()
	return contactsChan
}

func (db *Database) SaveAllContactsToDisk(userID string, contacts Contacts) {
	if db.db == nil {
		return
	}
	dbi := db.db.(js.Value)
	if !dbi.IsUndefined() && !dbi.IsNull() {
		req := dbi.Call("transaction", js.ValueOf(ContactsDir)).Call("objectStore", js.ValueOf(ContactsDir)).Call("get", js.ValueOf(userID))
		accountVal := js.ValueOf(map[string]interface{}{
			"ID":       userID,
			"contacts": contacts,
		})
		req.Call("put", accountVal)
	}
}

func (db *Database) SaveContactToDisk(userID string, eachContact *Contact) {
	if db.db == nil {
		return
	}
	dbi := db.db.(js.Value)
	if !dbi.IsUndefined() && !dbi.IsNull() {
		req := dbi.Call("transaction", js.ValueOf(ContactsDir)).Call("objectStore", js.ValueOf(ContactsDir)).Call("get", js.ValueOf(userID))
		req.Set("onsuccess", js.FuncOf(func(this js.Value, args []js.Value) interface{} {
			acc := args[0].Get("target").Get("result")
			if acc.IsUndefined() || acc.IsNull() {
				req.Call("put", js.ValueOf(map[string]interface{}{
					"ID":       userID,
					"contacts": js.ValueOf(map[string]interface{}{}),
				}))
				acc = args[0].Get("target").Get("result")
			}
			contactsJs := acc.Get("contacts")
			foundContact := js.ValueOf(map[string]interface{}{
				"PubKeyHex": eachContact.PublicKey,
				"ID":        eachContact.IDStr,
				"Name":      eachContact.Name,
				"Image":     string(eachContact.Image),
			})
			contactsJs.Set(eachContact.IDStr, foundContact)
			req = dbi.Call("transaction", js.ValueOf(ContactsDir), js.ValueOf("readwrite")).Call("objectStore", js.ValueOf(ContactsDir))
			req.Call("put", js.ValueOf(map[string]interface{}{
				"ID":       userID,
				"contacts": contactsJs,
			}))
			return nil
		}))
	}
}

// Todo TextMessages related to this contact should also be deleted
func (db *Database) DeleteContact(userID string, eachContact *Contact) {
	if db.db == nil {
		return
	}
	dbi := db.db.(js.Value)
	if !dbi.IsUndefined() && !dbi.IsNull() {
		req := dbi.Call("transaction", js.ValueOf(ContactsDir)).Call("objectStore", js.ValueOf(ContactsDir)).Call("get", js.ValueOf(userID))
		req.Set("onsuccess", js.FuncOf(func(this js.Value, args []js.Value) interface{} {
			acc := args[0].Get("target").Get("result")
			if acc.IsUndefined() || acc.IsNull() {
				req.Call("put", js.ValueOf(map[string]interface{}{
					"ID":       userID,
					"contacts": js.ValueOf(map[string]interface{}{}),
				}))
				acc = args[0].Get("target").Get("result")
			}
			contactsJs := acc.Get("contacts")
			foundContact := contactsJs.Get(eachContact.IDStr)
			if foundContact.IsUndefined() || foundContact.IsNull() {
				return nil
			}
			contactsJs.Delete(eachContact.IDStr)
			req = dbi.Call("transaction", js.ValueOf(ContactsDir), js.ValueOf("readwrite")).Call("objectStore", js.ValueOf(ContactsDir))
			req.Call("put", js.ValueOf(map[string]interface{}{
				"ID":       userID,
				"contacts": contactsJs,
			}))
			return nil
		}))
	}
}

func ContactsToArray(contacts Contacts) []*Contact {
	contactsArr := make([]*Contact, 0, len(contacts))
	for _, contact := range contacts {
		contactsArr = append(contactsArr, contact)
	}
	return contactsArr
}
