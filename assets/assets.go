package assets

import (
	"bytes"
	_ "embed"
	"image"
	_ "image/png"
)

//go:embed appicon.png

// AppIcon is encoded protonet's icon in png format
var AppIcon []byte

// AppIconImage is decoded Image representing AppIcon
var AppIconImage, _, _ = image.Decode(bytes.NewReader(AppIcon))
