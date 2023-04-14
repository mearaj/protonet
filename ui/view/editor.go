package view

import (
	"gioui.org/layout"
	"gioui.org/op"
	"gioui.org/op/clip"
	"gioui.org/widget"
	"gioui.org/widget/material"
	"gioui.org/x/component"
	"github.com/mearaj/protonet/assets/fonts"
	"github.com/mearaj/protonet/ui/fwk"
	"image"
	"time"
)

type EditorAnimated struct {
	initialized bool
	fwk.Animation
	widget.Editor
	*material.Theme
	layout.Inset
}

func (e *EditorAnimated) Layout(gtx fwk.Gtx) fwk.Dim {
	if !e.initialized {
		if e.Theme == nil {
			e.Theme = fonts.NewTheme()
		}
		if e.Animation.Duration == time.Duration(0) {
			e.Animation.Duration = time.Millisecond * 150
			e.Animation.State = component.Invisible
		}
		if e.Inset == (layout.Inset{}) {
			e.Inset = layout.Inset{Left: 16, Right: 16}
		}
		e.Editor.SingleLine = true
		e.initialized = true
	}
	return e.Inset.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
		progress := e.Animation.Revealed(gtx)
		if progress == 0 {
			return layout.Dimensions{}
		}
		inset := layout.Inset{Top: 8, Bottom: 8, Left: 12, Right: 12}
		rec := op.Record(gtx.Ops)
		dims := inset.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
			return material.Editor(e.Theme, &e.Editor, "").Layout(gtx)
		})
		call := rec.Stop()
		radii := 8
		dims.Size.X = int(float32(dims.Size.X) * progress)
		//dims.Size.Y = int(float32(dims.Size.Y) * progress)
		radii = int(float32(radii) * progress)
		rect := component.Rect{Color: e.Theme.Bg, Size: dims.Size, Radii: radii}
		rect.Layout(gtx)
		rRect := clip.RRect{Rect: image.Rect(0, 0,
			dims.Size.X,
			dims.Size.Y),
			SE: radii, SW: radii, NW: radii, NE: radii}.Push(gtx.Ops)
		call.Add(gtx.Ops)
		rRect.Pop()
		return dims
	})
}
