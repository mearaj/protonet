package view

import (
	"gioui.org/layout"
	"gioui.org/widget/material"
	"github.com/mearaj/protonet/assets/fonts"
	. "github.com/mearaj/protonet/ui/fwk"
	"image"
)

type Loader struct {
	*material.Theme
	loader material.LoaderStyle
	Size   image.Point
}

func (l *Loader) Layout(gtx Gtx) Dim {
	var th *material.Theme
	if l.Theme == nil {
		l.Theme = fonts.NewTheme()
	}
	th = l.Theme
	return layout.Flex{Alignment: layout.Middle,
		Axis:    layout.Vertical,
		Spacing: layout.SpaceSides}.Layout(gtx,
		layout.Flexed(1.0, func(gtx Gtx) Dim {
			return layout.Center.Layout(gtx,
				func(gtx Gtx) Dim {
					if l.Size == (image.Point{}) {
						l.Size = image.Point{X: gtx.Dp(56), Y: gtx.Dp(56)}
					}
					gtx.Constraints.Min = l.Size
					l.loader.Color = th.ContrastBg
					return l.loader.Layout(gtx)
				},
			)
		}),
	)
}
