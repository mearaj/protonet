package view

import (
	"gioui.org/layout"
	"gioui.org/op"
	"gioui.org/text"
	"gioui.org/unit"
	"gioui.org/widget"
	"gioui.org/widget/material"
	"gioui.org/x/component"
	. "github.com/mearaj/protonet/ui/fwk"
	"golang.org/x/exp/shiny/materialdesign/colornames"
	"golang.org/x/exp/shiny/materialdesign/icons"
	"image"
	"image/color"
)

type ModalContent struct {
	btnClose     widget.Clickable
	iconClose    *widget.Icon
	OnCloseClick func()
	layout.List
}

func NewModalContent(onCloseClick func()) *ModalContent {
	iconClear, _ := widget.NewIcon(icons.ContentClear)
	return &ModalContent{
		iconClose:    iconClear,
		OnCloseClick: onCloseClick,
		List:         layout.List{Axis: layout.Vertical},
	}
}

func (m *ModalContent) DrawContent(gtx Gtx, theme *material.Theme, contentWidget layout.Widget) Dim {
	if m.iconClose == nil {
		m.iconClose, _ = widget.NewIcon(icons.ContentClear)
	}
	if m.btnClose.Clicked() {
		if m.OnCloseClick != nil {
			m.OnCloseClick()
		}
	}
	mac := op.Record(gtx.Ops)
	d := layout.Flex{Axis: layout.Vertical}.Layout(gtx,
		layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			gtx.Constraints.Min.X = gtx.Constraints.Max.X
			vert := unit.Dp(16)
			horiz := unit.Dp(8)
			inset := layout.Inset{Top: vert, Bottom: vert, Right: horiz, Left: horiz}
			return inset.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
				return layout.Flex{Spacing: layout.SpaceBetween, Alignment: layout.Middle}.Layout(gtx,
					layout.Rigid(func(gtx layout.Context) layout.Dimensions {
						btn := material.IconButtonStyle{
							Icon:        m.iconClose,
							Button:      &m.btnClose,
							Description: "close backdrop",
						}
						btn.Inset = layout.UniformInset(unit.Dp(4))
						btn.Size = unit.Dp(24)
						btn.Background = theme.ContrastBg
						btn.Color = theme.ContrastFg
						return btn.Layout(gtx)
					}),
					layout.Rigid(layout.Spacer{Width: unit.Dp(8)}.Layout),
					layout.Rigid(func(gtx layout.Context) layout.Dimensions {
						bd := material.Body1(theme, "Protonet")
						bd.TextSize = unit.Sp(18)
						bd.Font.Weight = text.ExtraBold
						bd.Color = theme.ContrastBg
						return bd.Layout(gtx)
					}),
					layout.Rigid(layout.Spacer{Width: unit.Dp(8)}.Layout),
					layout.Rigid(func(gtx layout.Context) layout.Dimensions {
						btn := material.IconButtonStyle{
							Icon:        m.iconClose,
							Button:      &m.btnClose,
							Description: "close backdrop",
						}
						btn.Inset = layout.UniformInset(unit.Dp(4))
						btn.Size = unit.Dp(24)
						btn.Background = theme.ContrastBg
						btn.Color = theme.ContrastFg
						return btn.Layout(gtx)
					}),
				)
			})
		}),
		layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			return component.Rect{
				Color: color.NRGBA(colornames.Grey300),
				Size:  image.Point{Y: gtx.Dp(1), X: gtx.Constraints.Max.X},
				Radii: 0,
			}.Layout(gtx)
		}),
		layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			return m.List.Layout(gtx, 1, func(gtx layout.Context, index int) layout.Dimensions {
				return contentWidget(gtx)
			})
		}),
	)
	call := mac.Stop()
	component.Rect{Color: theme.Bg, Size: d.Size, Radii: gtx.Dp(8)}.Layout(gtx)
	call.Add(gtx.Ops)
	return d
}
