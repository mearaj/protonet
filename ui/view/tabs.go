package view

import (
	"gioui.org/layout"
	"gioui.org/op/clip"
	"gioui.org/op/paint"
	"gioui.org/unit"
	"gioui.org/widget"
	"gioui.org/widget/material"
	"github.com/mearaj/protonet/assets/fonts"
	"github.com/mearaj/protonet/ui/fwk"
	"image"
)

type Tabs struct {
	List     layout.List
	Tabs     []Tab
	Selected int
	Slider   Slider
	Header   func(gtx fwk.Gtx, index int) fwk.Dim
	Body     func(gtx fwk.Gtx, index int) fwk.Dim
}
type Tab struct {
	Clickable widget.Clickable
	*material.Theme
}

func (t *Tabs) Layout(gtx fwk.Gtx) fwk.Dim {
	if len(t.Tabs) == 0 {
		return fwk.Dim{}
	}
	return layout.Flex{Axis: layout.Vertical}.Layout(gtx,
		layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			return t.List.Layout(gtx, len(t.Tabs), func(gtx layout.Context, index int) layout.Dimensions {
				tab := &t.Tabs[index]
				if tab.Clickable.Clicked() {
					if t.Selected < index {
						t.Slider.PushLeft()
					} else if t.Selected > index {
						t.Slider.PushRight()
					}
					t.Selected = index
				}
				var tabWidth int
				return layout.Stack{Alignment: layout.S}.Layout(gtx,
					layout.Stacked(func(gtx fwk.Gtx) fwk.Dim {
						dims := material.Clickable(gtx, &tab.Clickable, func(gtx fwk.Gtx) fwk.Dim {
							if t.Header != nil {
								return t.Header(gtx, index)
							}
							return fwk.Dim{}
						})
						tabWidth = dims.Size.X
						return dims
					}),
					layout.Stacked(func(gtx fwk.Gtx) fwk.Dim {
						if t.Selected != index {
							return layout.Dimensions{}
						}
						tabHeight := gtx.Dp(unit.Dp(4))
						tabRect := image.Rect(0, 0, tabWidth, tabHeight)
						if tab.Theme == nil {
							tab.Theme = fonts.NewTheme()
						}
						paint.FillShape(gtx.Ops, tab.Palette.ContrastBg, clip.Rect(tabRect).Op())
						return layout.Dimensions{
							Size: image.Point{X: tabWidth, Y: tabHeight},
						}
					}),
				)
			})
		}),
		layout.Flexed(1, func(gtx fwk.Gtx) fwk.Dim {
			return t.Slider.Layout(gtx, func(gtx fwk.Gtx) fwk.Dim {
				if t.Body != nil {
					return t.Body(gtx, t.Selected)
				}
				return fwk.Dim{}
			})
		}),
	)
}
