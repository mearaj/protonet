package view

import (
	"gioui.org/layout"
	"gioui.org/text"
	"gioui.org/unit"
	"gioui.org/widget/material"
	"github.com/mearaj/protonet/assets/fonts"
	. "github.com/mearaj/protonet/ui/fwk"
)

type Greetings struct {
	*material.Theme
}

func NewGreetings(theme *material.Theme) Greetings {
	return Greetings{Theme: theme}
}

func (cp *Greetings) Layout(gtx Gtx) Dim {
	if cp.Theme == nil {
		cp.Theme = fonts.NewTheme()
	}

	flex := layout.Flex{Axis: layout.Vertical, Spacing: layout.SpaceSides, Alignment: layout.Middle}
	gtx.Constraints.Min.Y = gtx.Constraints.Max.Y
	return flex.Layout(gtx,
		layout.Rigid(func(gtx Gtx) Dim {
			return DrawProtonetImageCenter(gtx, cp.Theme)
		}),
		layout.Rigid(layout.Spacer{Height: unit.Dp(16)}.Layout),
		layout.Rigid(func(gtx Gtx) Dim {
			return layout.Center.Layout(gtx, func(gtx Gtx) Dim {
				bdy := material.Body1(cp.Theme, "Welcome to Protonet !")
				bdy.Alignment = text.Middle
				bdy.Font.Weight = text.Black
				return bdy.Layout(gtx)
			})
		}),
		layout.Rigid(layout.Spacer{Height: unit.Dp(16)}.Layout),
	)
}
