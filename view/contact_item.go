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
	"github.com/mearaj/protonet/service"
	"golang.org/x/image/colornames"
	"image"
	"image/color"
	"time"
)

type ListItem struct {
	material.Theme
	widget.Bool
	Hovered  bool
	Selected bool
	widget.Clickable
	component.AlphaPalette
	checkbox             *widget.Bool
	checkboxClickable    widget.Clickable
	contactNameClickable widget.Clickable
	index                int
	cs                   *service.TxtChatService
	lastMessage          string
}

func NewContactListItem(nav *Navigator, index int, cs *service.TxtChatService) *ListItem {
	li := &ListItem{}
	li.Theme = nav.Theme
	li.index = index
	li.cs = cs
	li.AlphaPalette = nav.AlphaPalette
	li.checkbox = new(widget.Bool)
	return li
}

func (li *ListItem) Layout(gtx C) D {
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

	passStack := pointer.PassOp{}.Push(gtx.Ops)
	areaStack := clip.Rect(image.Rectangle{
		Max: gtx.Constraints.Max,
	}).Push(gtx.Ops)
	pointer.InputOp{
		Tag:   li,
		Types: pointer.Enter | pointer.Leave | pointer.Cancel,
	}.Add(gtx.Ops)
	passStack.Pop()
	areaStack.Pop()
	return material.Clickable(gtx, &li.Clickable, func(gtx C) D {
		return li.layoutContent(gtx)
	})

}

func (li *ListItem) layoutContent(gtx C) D {
	gtx.Constraints.Min = gtx.Constraints.Max
	labelColor := li.Theme.Fg
	checkboxWidget := material.CheckBox(&li.Theme, li.checkbox, "")
	checkboxWidget.Size = unit.Dp(32.0)
	macro := op.Record(gtx.Ops)
	nameBgColor := database.GetRandomNRGBA(li.index)
	if li.Hovered || li.checkbox.Value {
		labelColor = li.Theme.Bg
		checkboxWidget.IconColor = li.Theme.Bg
		nameBgColor.A = 37
	}
	d := layout.Flex{Spacing: layout.SpaceBetween, Alignment: layout.Middle}.Layout(gtx,
		layout.Rigid(func(gtx C) D {
			return layout.Inset{Left: unit.Dp(16.0)}.Layout(gtx, func(gtx C) D {
				return layout.Flex{Axis: layout.Vertical, Alignment: layout.Middle,
					Spacing: layout.SpaceSides}.Layout(gtx,
					layout.Rigid(func(gtx C) D {
						textTheme := li.Theme
						textTheme.Fg = color.NRGBA{R: 255, G: 255, B: 255, A: 255}
						initials := database.GetInitialsFromName(li.cs.GetClient().Name)
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
						label := material.Label(&li.Theme, unit.Sp(14), li.cs.GetClient().Name)
						label.Color = labelColor
						label.Font.Weight = text.Bold
						return layout.Inset{
							Bottom: unit.Dp(8.0),
						}.Layout(gtx, func(gtx C) D {
							return component.TruncatingLabelStyle(label).Layout(gtx)
						})
					}),
					layout.Rigid(func(gtx C) D {
						if arr := li.cs.TextMessagesToArray(); len(arr) > 0 {
							txtMsg := arr[len(arr)-1].Message
							if li.lastMessage != txtMsg {
								li.lastMessage = txtMsg
							}
						} else {
							return D{}
						}
						label := material.Label(&li.Theme, unit.Sp(14), li.lastMessage)
						label.Color = labelColor
						label.Font.Weight = text.Bold
						return layout.Inset{
							Bottom: unit.Dp(8.0),
						}.Layout(gtx, func(gtx C) D {
							return component.TruncatingLabelStyle(label).Layout(gtx)
						})
					}),
					layout.Rigid(func(gtx C) D {
						var txtMsg string
						if arr := li.cs.TextMessagesToArray(); len(arr) > 0 {
							timeVal := time.Unix(arr[len(arr)-1].Timestamp, 0)
							txtMsg = timeVal.Format("Mon Jan 2 15:04 2006")
						} else {
							return D{}
						}
						label := material.Label(&li.Theme, unit.Sp(14), txtMsg)
						label.Color = color.NRGBA{
							R: colornames.Gray.R,
							G: colornames.Gray.G,
							B: colornames.Gray.B,
							A: colornames.Gray.A,
						}
						label.Font.Weight = text.Bold
						label.Font.Style = text.Italic
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
			return material.Clickable(gtx, &li.checkboxClickable, func(gtx C) D {
				gtx.Constraints.Min.Y = gtx.Constraints.Max.Y
				return layout.Inset{Left: unit.Dp(16.0), Right: unit.Dp(8.0)}.Layout(gtx,
					func(gtx C) D {
						return layout.Flex{Alignment: layout.Middle,
							Spacing: layout.SpaceSides}.Layout(gtx,
							layout.Rigid(func(gtx C) D {
								return checkboxWidget.Layout(gtx)
							}),
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
	if li.Hovered || li.Selected || li.checkbox.Value {
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
