package view

import (
	"gioui.org/layout"
	"gioui.org/unit"
	"gioui.org/widget"
	"gioui.org/widget/material"
	. "github.com/mearaj/protonet/ui/fwk"
	"image"
)

type IconButton struct {
	*material.Theme
	Button widget.Clickable
	Icon   *widget.Icon
	Text   string
	layout.Inset
}

func (b *IconButton) Layout(gtx Gtx) Dim {
	btnLayoutStyle := material.ButtonLayout(b.Theme, &b.Button)
	btnLayoutStyle.CornerRadius = unit.Dp(8)
	return btnLayoutStyle.Layout(gtx, func(gtx Gtx) Dim {
		inset := b.Inset
		if b.Inset == (layout.Inset{}) {
			inset = layout.UniformInset(unit.Dp(12))
		}
		return inset.Layout(gtx, func(gtx Gtx) Dim {
			iconAndLabel := layout.Flex{Alignment: layout.Middle, Spacing: layout.SpaceSides}
			textIconSpacer := unit.Dp(5)

			layIcon := layout.Rigid(func(gtx Gtx) Dim {
				return layout.Inset{Right: textIconSpacer}.Layout(gtx, func(gtx Gtx) Dim {
					var d Dim
					if b.Icon != nil {
						size := gtx.Dp(56.0 / 2.5)
						d = Dim{Size: image.Pt(size, size)}
						gtx.Constraints = layout.Exact(d.Size)
						d = b.Icon.Layout(gtx, b.Theme.ContrastFg)
					}
					return d
				})
			})

			layLabel := layout.Rigid(func(gtx Gtx) Dim {
				return layout.Inset{Left: textIconSpacer}.Layout(gtx, func(gtx Gtx) Dim {
					l := material.Label(b.Theme, b.Theme.TextSize, b.Text)
					l.Color = b.Theme.Palette.ContrastFg
					return l.Layout(gtx)
				})
			})

			return iconAndLabel.Layout(gtx, layIcon, layLabel)
		})
	})
}
