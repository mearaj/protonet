package view

import (
	"gioui.org/f32"
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
		Size:  image.Point{X: gtx.Px(unit.Dp(48)), Y: gtx.Px(unit.Dp(48))},
		Radii: float32(gtx.Px(unit.Dp(48)) / 2),
	}.Layout(gtx)
	macro2 := op.Record(gtx.Ops)
	d2 := material.Label(&textTheme, unit.Dp(20), initials).Layout(gtx)
	macro2.Stop()
	op.Offset(f32.Point{
		X: float32(d.Size.X-d2.Size.X) / 2,
		Y: float32(d.Size.Y-d2.Size.Y) / 2,
	}).Add(gtx.Ops)
	material.Label(&textTheme, unit.Dp(20), initials).Layout(gtx)
	return d
}
