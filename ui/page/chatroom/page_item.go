package chatroom

import (
	"gioui.org/layout"
	"gioui.org/op"
	"gioui.org/op/clip"
	"gioui.org/text"
	"gioui.org/unit"
	"gioui.org/widget"
	"gioui.org/widget/material"
	"gioui.org/x/component"
	"github.com/mearaj/audio"
	"github.com/mearaj/protonet/alog"
	"github.com/mearaj/protonet/assets/fonts"
	"github.com/mearaj/protonet/internal/chat"
	. "github.com/mearaj/protonet/ui/fwk"
	"golang.org/x/exp/shiny/materialdesign/icons"
	"image"
	"math"
)

type PageItem struct {
	chat.Message
	*material.Theme
	btnPlayPauseIcon widget.Clickable
	playIcon         *widget.Icon
	stopIcon         *widget.Icon
	accountPublicKey string
	player           *audio.RawPlayer
}

func (p *PageItem) Layout(gtx Gtx) (d Dim) {
	if p.Message.Text == "" && len(p.Message.Audio) == 0 {
		return d
	}
	if p.Theme == nil {
		p.Theme = fonts.NewTheme()
	}
	if p.playIcon == nil {
		p.playIcon, _ = widget.NewIcon(icons.AVPlayArrow)
	}
	if p.stopIcon == nil {
		p.stopIcon, _ = widget.NewIcon(icons.AVStop)
	}

	isMe := p.accountPublicKey == p.Message.Sender
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
				timeVal := p.Message.CreatedAt
				txtMsg := timeVal.Local().Format("Mon, Jan 2, 3:04 PM")
				label := material.Label(p.Theme, p.Theme.TextSize*0.70, txtMsg)
				label.Color = p.Theme.ContrastBg
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
							iconColor := p.Theme.ContrastBg
							switch p.Message.State {
							case chat.MessageStateReceived:
								icon, _ = widget.NewIcon(icons.ActionDone)
							case chat.MessageStateRead:
								icon, _ = widget.NewIcon(icons.ActionDoneAll)
							}
							return icon.Layout(gtx, iconColor)
						}
						return Dim{}
					}),
					layout.Rigid(func(gtx Gtx) Dim {
						if p.Message.Text != "" {
							macro := op.Record(gtx.Ops)
							inset := layout.UniformInset(unit.Dp(12))
							d := inset.Layout(gtx, func(gtx Gtx) Dim {
								flex := layout.Flex{}
								gtx.Constraints.Min.X = 0
								return flex.Layout(gtx,
									layout.Rigid(func(gtx Gtx) Dim {
										gtx.Constraints.Max.X = int(float32(gtx.Constraints.Max.X) / 1.5)
										bd := material.Body1(p.Theme, p.Message.Text)
										return bd.Layout(gtx)
									}))
							})
							call := macro.Stop()
							bgColor := p.Theme.ContrastBg
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
						} else if len(p.Message.Audio) != 0 {
							btn := &p.btnPlayPauseIcon
							icon := p.playIcon
							descText := "Play Audio Message"
							p.handlePlayPauseClick(gtx)
							isPlaying := p.player != nil &&
								(p.player.State() == audio.RawPlayerStatePlaying)
							if isPlaying {
								icon = p.stopIcon
								descText = "Pause Audio Message"
							}
							button := material.IconButton(p.Theme, btn, icon, descText)
							button.Size = unit.Dp(32)
							button.Background = p.Theme.Palette.ContrastBg
							button.Color = p.Theme.Palette.ContrastFg
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

func (p *PageItem) handlePlayPauseClick(gtx Gtx) {
	if p.btnPlayPauseIcon.Clicked() {
		go func() {
			if p.player == nil {
				var err error
				p.player, err = audio.NewRawPlayer(p.Message.Audio, 0, 0)
				if err != nil {
					alog.Logger().Errorln(err)
				} else {
					_ = p.player.Play()
				}
			} else {
				state := p.player.State()
				if state == audio.RawPlayerStatePlaying {
					_ = p.player.Stop()
				} else {
					_ = p.player.Play()
				}
			}
		}()
	}
}
