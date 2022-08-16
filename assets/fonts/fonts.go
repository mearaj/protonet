package fonts

import (
	_ "embed"
	"gioui.org/font/opentype"
	"gioui.org/text"
	"gioui.org/widget/material"
	"image/color"
)

//go:embed custom-black.ttf

var customBlack []byte

//go:embed custom-black-italic.ttf

var customBlackItalic []byte

//go:embed custom-bold.ttf

var customBold []byte

//go:embed custom-bold-italic.ttf

var customBoldItalic []byte

//go:embed custom-light.ttf

var customLight []byte

//go:embed custom-light-italic.ttf

var customLightItalic []byte

//go:embed custom-medium.ttf

var customMedium []byte

//go:embed custom-medium-italic.ttf

var customMediumItalic []byte

//go:embed custom-regular.ttf

var customRegular []byte

//go:embed custom-regular-italic.ttf

var customRegularItalic []byte

//go:embed custom-thin.ttf

var customThin []byte

//go:embed custom-thin-italic.ttf

var customThinItalic []byte

var black, _ = opentype.Parse(customBlack)
var blackItalic, _ = opentype.Parse(customBlackItalic)
var bold, _ = opentype.Parse(customBold)
var boldItalic, _ = opentype.Parse(customBoldItalic)
var light, _ = opentype.Parse(customLight)
var lightItalic, _ = opentype.Parse(customLightItalic)
var medium, _ = opentype.Parse(customMedium)
var mediumItalic, _ = opentype.Parse(customMediumItalic)
var regular, _ = opentype.Parse(customRegular)
var regularItalic, _ = opentype.Parse(customRegularItalic)
var thin, _ = opentype.Parse(customThin)
var thinItalic, _ = opentype.Parse(customThinItalic)

var BlackFont = text.Font{Weight: text.Black, Style: text.Regular}
var blackItalicFont = text.Font{Weight: text.Black, Style: text.Italic}
var boldFont = text.Font{Weight: text.Bold, Style: text.Regular}
var boldItalicFont = text.Font{Weight: text.Bold, Style: text.Italic}
var lightFont = text.Font{Weight: text.Light, Style: text.Regular}
var lightItalicFont = text.Font{Weight: text.Light, Style: text.Italic}
var mediumFont = text.Font{Weight: text.Medium, Style: text.Regular}
var mediumItalicFont = text.Font{Weight: text.Medium, Style: text.Italic}
var regularFont = text.Font{Weight: text.Normal, Style: text.Regular}
var regularItalicFont = text.Font{Weight: text.Normal, Style: text.Italic}
var thinFont = text.Font{Weight: text.Thin, Style: text.Regular}
var thinItalicFont = text.Font{Weight: text.Thin, Style: text.Italic}

var collection = []text.FontFace{
	{Font: BlackFont, Face: black},
	{Font: blackItalicFont, Face: blackItalic},
	{Font: boldFont, Face: bold},
	{Font: boldItalicFont, Face: boldItalic},
	{Font: lightFont, Face: light},
	{Font: lightItalicFont, Face: lightItalic},
	{Font: mediumFont, Face: medium},
	{Font: mediumItalicFont, Face: mediumItalic},
	{Font: regularFont, Face: regular},
	{Font: regularItalicFont, Face: regularItalic},
	{Font: thinFont, Face: thin},
	{Font: thinItalicFont, Face: thinItalic},
}
var AppColor = color.NRGBA{R: 102, G: 117, B: 127, A: 255}

func NewTheme() *material.Theme {
	th := material.NewTheme(collection)
	th.Bg.R = 245
	th.Bg.G = 245
	th.Bg.B = 255
	return th
}
