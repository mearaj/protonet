package view

import (
	"gioui.org/f32"
	"gioui.org/layout"
	"gioui.org/op/clip"
	"gioui.org/op/paint"
	"gioui.org/text"
	"gioui.org/unit"
	"gioui.org/widget"
	"gioui.org/widget/material"
	"gioui.org/x/component"
	"golang.org/x/exp/shiny/materialdesign/icons"
	"golang.org/x/image/colornames"
	"image/color"
	"protonet.live/database"
	"protonet.live/service"
	"time"
)

type ChatItem struct {
	icon        *widget.Icon
	iconDone    *widget.Icon
	iconDoneAll *widget.Icon
	//iconNotDone *widget.Icon
	msg         *database.TxtMsg
	cs          *service.TxtChatService
	th          material.Theme
}

func NewChatItem(msg *database.TxtMsg, cs *service.TxtChatService,
	th material.Theme) *ChatItem {
	ci := &ChatItem{}
	ci.iconDone, _ = widget.NewIcon(icons.ActionCheckCircle)
	ci.iconDoneAll, _ = widget.NewIcon(icons.ActionDoneAll)
	//ci.iconNotDone, _ = widget.NewIcon(icons.NavigationCancel)
	ci.icon = ci.iconDone
	ci.msg = msg
	ci.cs = cs
	ci.th = th
	return ci
}

func (ci *ChatItem) Layout(gtx C) D {
	labelColor := color.NRGBA{
		R: 0,
		G: 0,
		B: 0,
		A: 255,
	}
	rigidLayout := layout.Rigid(func(gtx C) D {
		label := material.Body1(&ci.th, ci.msg.Message)
		label.Color = labelColor
		label.Font.Weight = text.Bold
		leftSpace := unit.Dp(8.0)
		rightSpace := unit.Dp(8.0)
		label.Alignment = text.Start

		d := layout.Inset{
			Top:    unit.Dp(8.0),
			Left:   leftSpace,
			Right:  rightSpace,
			Bottom: unit.Dp(8.0),
		}.Layout(gtx, func(gtx C) D {
			gtx.Constraints.Max.X = gtx.Constraints.Max.X - 100
			return layout.Flex{Spacing: layout.SpaceBetween}.Layout(gtx,
				layout.Rigid(label.Layout),
			)
		})

		fill := color.NRGBA{
			R: colornames.Blue.R,
			G: colornames.Blue.G,
			B: 255,
			A: 50,
		}
		if ci.msg.CreatorID == ci.cs.User.ID {
			fill = color.NRGBA{
				R: colornames.Green.R,
				G: 255,
				B: colornames.Green.B,
				A: 50,
			}
		}
		path := clip.Path{}
		path.Begin(gtx.Ops)
		path.Move(f32.Pt(0, 0))
		maxX := float32(d.Size.X)
		maxY := float32(d.Size.Y)
		if ci.msg.CreatorID == ci.cs.User.ID {
			path.Line(f32.Pt(maxX+40, 0))
			path.Line(f32.Pt(-32, 16))
			//path.Quad(f32.Pt(16.0, 0), f32.Pt(16.0, 16))
			path.Line(f32.Pt(0, maxY-32))
			path.Quad(f32.Pt(0, 16), f32.Pt(-16.0, 16))
			path.Line(f32.Pt(-maxX+16, 0))
			path.Quad(f32.Pt(-16, 0), f32.Pt(-16, -16))
			path.Line(f32.Pt(0, -maxY+32))
			path.Quad(f32.Pt(0, -16), f32.Pt(16, -16))
		} else {
			path.Line(f32.Pt(maxX, 0))
			path.Quad(f32.Pt(16.0, 0), f32.Pt(16.0, 16))
			path.Line(f32.Pt(0, maxY-32))
			path.Quad(f32.Pt(0, 16), f32.Pt(-16.0, 16))
			path.Line(f32.Pt(-maxX+8, 0))
			path.Quad(f32.Pt(-16, 0), f32.Pt(-16, -16))
			path.Line(f32.Pt(0, -maxY+32))
			path.Line(f32.Pt(-32, -16))
			path.Line(f32.Pt(32, 0))
			//path.Quad(f32.Pt(0, -16), f32.Pt(16, -16))
		}
		path.Close()
		clip.Outline{Path: path.End()}.Op().Add(gtx.Ops)
		paint.ColorOp{Color: fill}.Add(gtx.Ops)
		paint.PaintOp{}.Add(gtx.Ops)

		//component.Rect{
		//	Color: fill,
		//	Size:  image.Point{X: maxX, Y: maxY},
		//	Radii: 16,
		//}.Layout(gtx)

		return d
	})

	d := layout.Inset{
		Top:    unit.Dp(16.0),
		Left:   unit.Dp(8.0),
		Right:  unit.Dp(8.0),
		Bottom: unit.Dp(16.0),
	}.Layout(gtx, func(gtx C) D {
		gtx.Constraints.Min.X = gtx.Constraints.Max.X
		userAvatarLayout := layout.Rigid(func(gtx C) D {
			if ci.msg.CreatorID != ci.cs.User.ID {
				return D{}
			}

			fill := color.NRGBA{
				R: colornames.Green.R,
				G: colornames.Green.G,
				B: colornames.Green.B,
				A: 220,
			}
			textTheme := ci.th
			textTheme.Fg = color.NRGBA{R: 255, G: 255, B: 255, A: 255}
			name := database.GetInitialsFromName(ci.cs.User.Name)
			return drawAvatar(gtx, name, fill, textTheme)
		})
		contactAvatarLayout := layout.Rigid(func(gtx C) D {
			if ci.msg.CreatorID != ci.cs.GetClient().IDStr {
				return D{}
			}
			fill := color.NRGBA{
				R: colornames.Blue.R,
				G: colornames.Blue.G,
				B: colornames.Blue.B,
				A: 220,
			}
			textTheme := ci.th
			textTheme.Fg = color.NRGBA{R: 255, G: 255, B: 255, A: 255}
			name := database.GetInitialsFromName(ci.cs.GetClient().Name)
			return drawAvatar(gtx, name, fill, textTheme)
		})
		spacer := layout.Rigid(func(gtx C) D {
			return layout.Spacer{
				Width: unit.Dp(16.0),
			}.Layout(gtx)
		})
		_ = spacer
		if ci.msg.CreatorID == ci.cs.GetClient().IDStr {
			return layout.Flex{Axis: layout.Vertical, Alignment: layout.Start}.Layout(gtx,
				layout.Rigid(func(gtx C) D {
					return layout.Flex{Spacing: layout.SpaceEnd}.Layout(gtx,
						layout.Rigid(func(gtx C) D {
							timeVal := time.Unix(ci.msg.Timestamp,0)
							txtMsg := timeVal.Format("Mon Jan 2 15:04 2006")
							label := material.Label(&ci.th, unit.Sp(12), txtMsg)
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
				}),
				layout.Rigid(func(gtx C) D {
					return layout.Flex{Spacing: layout.SpaceEnd}.Layout(gtx,
						layout.Rigid(func(gtx C) D {
							d := layout.Flex{Alignment: layout.Start,
								Axis:    layout.Vertical,
								Spacing: layout.SpaceBetween}.Layout(gtx,
								contactAvatarLayout,
							)
							return d
						}),
						rigidLayout,
					)
				}),
			)
		} else {
			return layout.Flex{Axis: layout.Vertical}.Layout(gtx,
				layout.Rigid(func(gtx C) D {
					return layout.Flex{Spacing: layout.SpaceStart}.Layout(gtx,
						layout.Rigid(func(gtx C) D {
							timeVal := time.Unix( ci.msg.Timestamp,0)
							txtMsg := timeVal.Format("Mon Jan 2 15:04 2006")
							label := material.Label(&ci.th, unit.Sp(12), txtMsg)
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
				}),
				layout.Rigid(func(gtx C) D {
					return layout.Flex{Spacing: layout.SpaceStart}.Layout(gtx,
						layout.Rigid(func(gtx C) D {
							// ci.icon = ci.iconNotDone
							colorGreen := color.NRGBA{
								R: colornames.Green.R,
								G: colornames.Green.G,
								B: colornames.Green.B,
								A: 255,
							}
							if !ci.msg.AckReceivedOrSent && !ci.msg.ReadAckReceivedOrSent {
								return D{}
							}
							if ci.msg.AckReceivedOrSent || ci.msg.ReadAckReceivedOrSent  {
								ci.icon = ci.iconDone
								//ci.icon.Color = colorGreen
								if ci.msg.ReadAckReceivedOrSent {
									ci.icon = ci.iconDoneAll
									//ci.icon.Color = colorGreen
								}
								return layout.Flex{Alignment: layout.Middle,
									Axis:    layout.Vertical,
									Spacing: layout.SpaceEvenly}.Layout(gtx,
									layout.Rigid(func(gtx C) D {
										return ci.icon.Layout(gtx, color.NRGBA{})
									}),
								)
							}
							return layout.Flex{Alignment: layout.Middle,
								Axis:    layout.Vertical,
								Spacing: layout.SpaceEvenly}.Layout(gtx,
								layout.Rigid(func(gtx C) D {
									loader := material.Loader(&ci.th)
									loader.Color = colorGreen
									return loader.Layout(gtx)
								}),
							)
						}),
						spacer,
						rigidLayout,
						layout.Rigid(func(gtx C) D {
							d := layout.Flex{Alignment: layout.Start,
								Axis:    layout.Vertical,
								Spacing: layout.SpaceBetween}.Layout(gtx,
								userAvatarLayout,
							)
							return d
						}),
					)
				}),
			)
		}
	})

	return d
}
