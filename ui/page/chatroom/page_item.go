package chatroom

import (
	"fmt"
	"gioui.org/layout"
	"gioui.org/op"
	"gioui.org/op/clip"
	"gioui.org/text"
	"gioui.org/unit"
	"gioui.org/widget"
	"gioui.org/widget/material"
	"gioui.org/x/component"
	"github.com/mearaj/protonet/assets/fonts"
	"github.com/mearaj/protonet/service"
	. "github.com/mearaj/protonet/ui/fwk"
	"golang.org/x/exp/shiny/materialdesign/icons"
	"image"
	"math"
	"time"
)

type PageItem struct {
	service.Message
	*material.Theme
	btnPlayIcon  widget.Clickable
	btnPauseIcon widget.Clickable
	playIcon     *widget.Icon
	pauseIcon    *widget.Icon
	isPlaying    bool
}

func (c *PageItem) Layout(gtx Gtx) (d Dim) {
	if c.Message.Text == "" && len(c.Message.Audio) == 0 {
		return d
	}
	if c.Theme == nil {
		c.Theme = fonts.NewTheme()
	}
	if c.playIcon == nil {
		c.playIcon, _ = widget.NewIcon(icons.AVPlayArrow)
	}
	if c.pauseIcon == nil {
		c.playIcon, _ = widget.NewIcon(icons.AVPlayArrow)
	}

	isMe := c.Message.AccountPublicKey == c.Message.From
	inset := layout.Inset{Top: unit.Dp(24), Bottom: unit.Dp(0)}
	d = inset.Layout(gtx, func(gtx Gtx) Dim {
		gtx.Constraints.Min.X = gtx.Constraints.Max.X
		flex := layout.Flex{Axis: layout.Vertical}
		if isMe {
			flex.Alignment = layout.End
		} else {
			flex.Alignment = layout.Start
		}
		d := flex.Layout(gtx,
			layout.Rigid(func(gtx Gtx) Dim {
				timeVal, _ := time.Parse(time.RFC3339, c.Message.Created)
				txtMsg := timeVal.Local().Format("Mon, Jan 2, 3:04 PM")
				label := material.Label(c.Theme, c.Theme.TextSize*0.70, txtMsg)
				label.Color = c.Theme.ContrastBg
				label.Color.A = uint8(int(math.Abs(float64(label.Color.A)-50)) % 256)
				label.Font.Weight = text.Bold
				label.Font.Style = text.Italic
				inset = layout.Inset{Bottom: unit.Dp(8.0)}
				d := inset.Layout(gtx, func(gtx Gtx) Dim {
					flex := layout.Flex{}
					if isMe {
						flex.Spacing = layout.SpaceStart
					} else {
						flex.Spacing = layout.SpaceEnd
					}
					return flex.Layout(gtx,
						layout.Rigid(func(gtx Gtx) Dim {
							return component.TruncatingLabelStyle(label).Layout(gtx)
						}))
				})
				return d
			}),
			layout.Rigid(func(gtx layout.Context) layout.Dimensions {
				flex := layout.Flex{Spacing: layout.SpaceEnd, Alignment: layout.Middle}
				if isMe {
					flex.Spacing = layout.SpaceStart
				}
				return flex.Layout(gtx,
					layout.Rigid(func(gtx layout.Context) layout.Dimensions {
						if isMe {
							icon, _ := widget.NewIcon(icons.AlertErrorOutline)
							iconColor := c.Theme.ContrastBg
							switch c.Message.State {
							case service.MessageReceivedSent:
								icon, _ = widget.NewIcon(icons.ActionDone)
							case service.MessageRead:
								icon, _ = widget.NewIcon(icons.ActionDoneAll)
							default:
								return Dim{}
							}
							return icon.Layout(gtx, iconColor)
						}
						return Dim{}
					}),
					layout.Rigid(func(gtx Gtx) Dim {
						if c.Message.Text != "" {
							macro := op.Record(gtx.Ops)
							inset := layout.UniformInset(unit.Dp(12))
							d := inset.Layout(gtx, func(gtx Gtx) Dim {
								flex := layout.Flex{}
								gtx.Constraints.Min.X = 0
								return flex.Layout(gtx,
									layout.Rigid(func(gtx Gtx) Dim {
										gtx.Constraints.Max.X = int(float32(gtx.Constraints.Max.X) / 1.5)
										bd := material.Body1(c.Theme, c.Message.Text)
										return bd.Layout(gtx)
									}))
							})
							call := macro.Stop()
							bgColor := c.Theme.ContrastBg
							bgColor.A = 50
							radius := gtx.Dp(16)
							sE, sW, nW, nE := radius, radius, radius, radius
							if isMe {
								nE = 0
							} else {
								nW = 0
							}
							clipOp := clip.RRect{Rect: image.Rectangle{
								Max: image.Point{X: d.Size.X, Y: d.Size.Y},
							}, SE: sE, SW: sW, NW: nW, NE: nE}.Push(gtx.Ops)
							component.Rect{Color: bgColor, Size: d.Size}.Layout(gtx)
							call.Add(gtx.Ops)
							clipOp.Pop()
							return d
						} else if len(c.Message.Audio) != 0 {
							if c.btnPlayIcon.Clicked() {
								fmt.Println("clicked")
							}
							icon := c.playIcon
							descText := "Play Audio Message"
							btn := &c.btnPlayIcon
							if c.isPlaying {
								icon = c.pauseIcon
								descText = "Pause Audio Message"
								btn = &c.btnPauseIcon
							}
							button := material.IconButton(c.Theme, btn, icon, descText)
							button.Size = unit.Dp(40)
							button.Background = c.Theme.Palette.ContrastBg
							button.Color = c.Theme.Palette.ContrastFg
							button.Inset = layout.UniformInset(unit.Dp(8))
							return button.Layout(gtx)
						}
						return Dim{}
					}),
				)
			}),
		)
		return d
	})
	return d
}
