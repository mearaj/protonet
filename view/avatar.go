package view

import (
	"gioui.org/op"
	"gioui.org/unit"
	"gioui.org/widget/material"
	"gioui.org/x/component"
	"image"
	"image/color"
)

func drawAvatar(gtx C, initials string, bgColor color.NRGBA, textTheme material.Theme) D {
	d := component.Rect{
		Color: bgColor,
		Size:  image.Point{X: gtx.Dp(48), Y: gtx.Dp(48)},
		Radii: int(float32(gtx.Dp(48)) / 2.0),
	}.Layout(gtx)
	macro2 := op.Record(gtx.Ops)
	d2 := material.Label(&textTheme, unit.Sp(20), initials).Layout(gtx)
	macro2.Stop()
	op.Offset(image.Point{
		X: int(float32(d.Size.X-d2.Size.X) / 2.0),
		Y: int(float32(d.Size.Y-d2.Size.Y) / 2.0),
	}).Add(gtx.Ops)
	material.Label(&textTheme, unit.Sp(20), initials).Layout(gtx)
	return d
}
