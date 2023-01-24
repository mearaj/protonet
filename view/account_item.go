package view

import (
	"gioui.org/io/pointer"
	"gioui.org/layout"
	"gioui.org/op"
	"gioui.org/op/clip"
	"gioui.org/text"
	"gioui.org/unit"
	"gioui.org/widget"
	"gioui.org/widget/material"
	"gioui.org/x/component"
	"github.com/mearaj/protonet/database"
	"image"
	"image/color"
)

type AccountListItem struct {
	material.Theme
	widget.Bool
	Hovered  bool
	Selected bool
	widget.Clickable
	component.AlphaPalette
	*database.Account
	group                widget.Enum
	radioButtonsGroup    *widget.Enum
	radioClickable       widget.Clickable
	contactNameClickable widget.Clickable
	index                int
}

func NewAccountListItem(nav *Navigator, index int, radioButtonsGroup *widget.Enum, Account *database.Account) *AccountListItem {
	li := &AccountListItem{}
	li.Theme = nav.Theme
	li.index = index
	li.radioButtonsGroup = radioButtonsGroup
	li.AlphaPalette = nav.AlphaPalette
	li.Account = Account
	return li
}

func (li *AccountListItem) Layout(gtx C) D {
	events := gtx.Events(li)
	for _, event := range events {
		switch event := event.(type) {
		case pointer.Event:
			switch event.Type {
			case pointer.Enter:
				li.Hovered = true
			case pointer.Leave:
				li.Hovered = false
			case pointer.Cancel:
				li.Hovered = false
			}
		}
	}

	stackPass := pointer.PassOp{}.Push(gtx.Ops)
	areaStack := clip.Rect(image.Rectangle{
		Max: gtx.Constraints.Max,
	}).Push(gtx.Ops)
	pointer.InputOp{
		Tag:   li,
		Types: pointer.Enter | pointer.Leave | pointer.Cancel,
	}.Add(gtx.Ops)
	stackPass.Pop()
	areaStack.Pop()
	return material.Clickable(gtx, &li.Clickable, func(gtx C) D {
		return li.layoutContent(gtx)
	})

}

func (li *AccountListItem) layoutContent(gtx C) D {
	gtx.Constraints.Min = gtx.Constraints.Max
	labelColor := li.Theme.Fg
	macro := op.Record(gtx.Ops)
	nameBgColor := database.GetRandomNRGBA(li.index)
	//if li.Hovered {
	//	labelColor = li.Theme.Bg
	//	checkboxWidget.IconColor = li.Theme.Bg
	//	nameBgColor.A = 37
	//}
	d := layout.Flex{Spacing: layout.SpaceBetween, Alignment: layout.Middle}.Layout(gtx,
		layout.Rigid(func(gtx C) D {
			return layout.Inset{Left: unit.Dp(16.0)}.Layout(gtx, func(gtx C) D {
				return layout.Flex{Axis: layout.Vertical, Alignment: layout.Middle,
					Spacing: layout.SpaceSides}.Layout(gtx,
					layout.Rigid(func(gtx C) D {
						textTheme := li.Theme
						textTheme.Fg = color.NRGBA{R: 255, G: 255, B: 255, A: 255}
						initials := database.GetInitialsFromName(li.Account.Name)
						return drawAvatar(gtx, initials, nameBgColor, textTheme)
					}),
				)
			})
		}),
		layout.Flexed(1.0, func(gtx C) D {
			return layout.Inset{
				Left: unit.Dp(16.0),
			}.Layout(gtx, func(gtx C) D {
				return layout.Flex{Axis: layout.Vertical,
					Alignment: layout.Start,
					Spacing:   layout.SpaceSides,
				}.Layout(gtx,
					layout.Rigid(func(gtx C) D {
						label := material.Label(&li.Theme, unit.Sp(14), li.Account.Name)
						label.Color = labelColor
						label.Font.Weight = text.Bold
						return layout.Inset{
							Bottom: unit.Dp(8.0),
						}.Layout(gtx, func(gtx C) D {
							return component.TruncatingLabelStyle(label).Layout(gtx)
						})
					}),
					layout.Rigid(func(gtx C) D {
						label := material.Label(&li.Theme, unit.Sp(14), li.Account.PvtKeyHex)
						label.Color = labelColor
						label.Font.Weight = text.Bold
						return layout.Inset{
							Bottom: unit.Dp(8.0),
						}.Layout(gtx, func(gtx C) D {
							return component.TruncatingLabelStyle(label).Layout(gtx)
						})
					}),
					layout.Rigid(func(gtx C) D {
						pubKey := li.Account.PubKeyHex
						label := material.Label(&li.Theme, unit.Sp(14), pubKey)
						label.Color = labelColor
						label.Font.Weight = text.Bold
						return layout.Inset{
							Bottom: unit.Dp(8.0),
						}.Layout(gtx, func(gtx C) D {
							return component.TruncatingLabelStyle(label).Layout(gtx)
						})
					}),
				)
			})
		}),
		layout.Rigid(func(gtx C) D {
			return material.Clickable(gtx, &li.radioClickable, func(gtx C) D {
				gtx.Constraints.Min.Y = gtx.Constraints.Max.Y
				return layout.Inset{Left: unit.Dp(16.0), Right: unit.Dp(8.0)}.Layout(gtx,
					func(gtx C) D {
						return layout.Flex{Alignment: layout.Middle,
							Spacing: layout.SpaceSides}.Layout(gtx,
							layout.Rigid(material.RadioButton(&li.Theme, li.radioButtonsGroup, li.Account.ID, "").Layout),
						)

					})
			})
		}),
	)
	operations := macro.Stop()
	fill := color.NRGBA{
		R: 255,
		G: 255,
		B: 255,
		A: 255,
	}
	if li.Selected {
		fill.R = 0
		fill.G = 0
		fill.B = 0
	}
	component.Rect{Color: fill,
		Size:  image.Point{X: d.Size.X, Y: d.Size.Y},
		Radii: 0,
	}.Layout(gtx)
	operations.Add(gtx.Ops)
	return d
}
