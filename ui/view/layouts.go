package view

import (
	"gioui.org/layout"
	"gioui.org/text"
	"gioui.org/unit"
	"gioui.org/widget"
	"gioui.org/widget/material"
	"gioui.org/x/component"
	. "github.com/mearaj/protonet/ui/fwk"
	"golang.org/x/exp/shiny/materialdesign/colornames"
	"image/color"
)

// DrawAppBarLayout reusable function to draw consistent AppBar
func DrawAppBarLayout(gtx Gtx, th *material.Theme, widget layout.Widget) Dim {
	gtx.Constraints.Max.Y = gtx.Dp(56)
	component.Rect{Size: gtx.Constraints.Max, Color: th.ContrastBg}.Layout(gtx)
	inset := layout.Inset{Left: unit.Dp(8), Right: unit.Dp(8)}
	return inset.Layout(gtx, widget)
}

type PromptContent struct {
	*material.Theme
	btnYes      *widget.Clickable
	btnNo       *widget.Clickable
	HeaderTxt   string
	ContentText string
}

func NewPromptContent(theme *material.Theme, headerText string, contentText string, btnYes *widget.Clickable, btnNo *widget.Clickable) View {
	return &PromptContent{
		Theme:       theme,
		btnYes:      btnYes,
		btnNo:       btnNo,
		HeaderTxt:   headerText,
		ContentText: contentText,
	}
}

func (p *PromptContent) Layout(gtx Gtx) Dim {
	gtx.Constraints.Min.X = gtx.Constraints.Max.X
	inset := layout.UniformInset(unit.Dp(16))
	d := inset.Layout(gtx, func(gtx Gtx) Dim {
		return layout.Flex{Axis: layout.Vertical}.Layout(gtx,
			layout.Rigid(func(gtx Gtx) Dim {
				if p.HeaderTxt == "" {
					return Dim{}
				}
				bd := material.Body1(p.Theme, p.HeaderTxt)
				bd.Font.Weight = text.Bold
				bd.Alignment = text.Middle
				return bd.Layout(gtx)
			}),
			layout.Rigid(func(gtx Gtx) Dim {
				return layout.Spacer{Height: unit.Dp(8)}.Layout(gtx)
			}),
			layout.Rigid(func(gtx Gtx) Dim {
				if p.ContentText == "" {
					return Dim{}
				}
				bd := material.Body1(p.Theme, p.ContentText)
				bd.Alignment = text.Middle
				return bd.Layout(gtx)
			}),
			layout.Rigid(layout.Spacer{Height: unit.Dp(16)}.Layout),
			layout.Rigid(func(gtx Gtx) Dim {
				return layout.Flex{Spacing: layout.SpaceSides, Alignment: layout.Middle}.Layout(gtx,
					layout.Rigid(func(gtx Gtx) Dim {
						btn := material.Button(p.Theme, p.btnYes, "Yes")
						btn.Background = color.NRGBA(colornames.Red500)
						return btn.Layout(gtx)
					}),
					layout.Rigid(layout.Spacer{Width: unit.Dp(16)}.Layout),
					layout.Rigid(func(gtx Gtx) Dim {
						btn := material.Button(p.Theme, p.btnNo, "No")
						btn.Background = color.NRGBA(colornames.Green500)
						return btn.Layout(gtx)
					}),
				)
			}),
		)
	})
	return d
}
