package db

import (
	"bytes"
	"encoding/gob"
	"github.com/mearaj/protonet/alog"
	_ "image/gif"
	_ "image/jpeg"
	_ "image/png"
)

// DecodeToStruct strc is a pointer to a struct
func DecodeToStruct(strc interface{}, s []byte) (err error) {
	dec := gob.NewDecoder(bytes.NewReader(s))
	err = dec.Decode(strc)
	if err != nil {
		alog.Logger().Errorln(err)
		return
	}
	return nil
}

// Ref https://gist.github.com/SteveBate/042960baa7a4795c3565
func EncodeToBytes(str interface{}) []byte {
	buf := bytes.Buffer{}
	enc := gob.NewEncoder(&buf)
	err := enc.Encode(str)
	if err != nil {
		alog.Logger().Errorln(err)
		return nil
	}
	return buf.Bytes()
}
