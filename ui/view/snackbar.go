package view

import (
	"gioui.org/layout"
	"gioui.org/op"
	"gioui.org/op/clip"
	"gioui.org/unit"
	"gioui.org/widget"
	"gioui.org/widget/material"
	"gioui.org/x/component"
	. "github.com/mearaj/protonet/ui/fwk"
	"image"
	"image/color"
	"strings"
	"time"
)

type snackBar struct {
	manager      Manager
	theme        *material.Theme
	txt          string
	actionTxt    string
	actionColor  color.NRGBA
	actionButton *widget.Clickable
	Animation    component.VisibilityAnimation
	startTime    int64
	duration     int64
}

func NewSnackBar(manager Manager) Snackbar {
	return &snackBar{manager: manager,
		theme:    manager.Theme(),
		duration: 3000,
		Animation: component.VisibilityAnimation{
			Duration: time.Millisecond * 500,
			State:    component.Invisible,
			Started:  time.Time{},
		}}
}

func (s *snackBar) Show(txt string, actionButton *widget.Clickable, actionColor color.NRGBA, actionText string) {
	s.txt = txt
	s.actionButton = actionButton
	s.actionTxt = actionText
	s.startTime = time.Now().UnixMilli()
	s.actionColor = actionColor
	if actionColor == (color.NRGBA{}) {
		s.actionColor = s.theme.ContrastBg
	}
	s.actionColor = s.theme.ContrastBg
	s.Animation.Appear(time.Now())
	s.manager.Window().Invalidate()
}

func (s *snackBar) Layout(gtx Gtx) (d Dim) {
	now := time.Now().UnixMilli()
	if s.startTime != 0 {
		diff := now - s.startTime
		if diff >= s.duration {
			s.startTime = 0
			s.Animation.Disappear(gtx.Now)
		}
	}
	layout.Stack{Alignment: layout.S}.Layout(gtx,
		layout.Stacked(func(gtx layout.Context) layout.Dimensions {
			progress := s.Animation.Revealed(gtx)
			minWidth := gtx.Dp(288)
			maxWidth := gtx.Dp(568)
			if s.manager.ShouldDrawSidebar() {
				gtx.Constraints.Min.X = minWidth
				gtx.Constraints.Max.X = maxWidth
			} else {
				gtx.Constraints.Min.X = gtx.Constraints.Max.X
			}
			inset := layout.Inset{Top: unit.Dp(14), Bottom: unit.Dp(14), Right: unit.Dp(24), Left: unit.Dp(24)}
			mac := op.Record(gtx.Ops)
			d = inset.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
				flex := layout.Flex{Alignment: layout.Middle, Spacing: layout.SpaceBetween}
				gtx.Constraints.Min.X = gtx.Constraints.Max.X
				return flex.Layout(gtx,
					layout.Rigid(func(gtx layout.Context) layout.Dimensions {
						label := material.Label(s.theme, unit.Sp(14.0), s.txt)
						label.MaxLines = 2
						label.Color = color.NRGBA{R: 255, G: 255, B: 255, A: 255}
						return component.TruncatingLabelStyle(label).Layout(gtx)
					}),
					layout.Rigid(layout.Spacer{Width: unit.Dp(24)}.Layout),
					layout.Rigid(func(gtx layout.Context) layout.Dimensions {
						if s.actionTxt == "" || s.actionButton == nil {
							return Dim{}
						}
						actionTxt := strings.ToUpper(s.actionTxt)
						actionColor := s.actionColor
						return material.Clickable(gtx, s.actionButton, func(gtx Gtx) Dim {
							return layout.UniformInset(unit.Dp(12)).Layout(gtx, func(gtx Gtx) Dim {
								flatBtnText := material.Body1(s.theme, actionTxt)
								flatBtnText.Color = actionColor
								return layout.Center.Layout(gtx, flatBtnText.Layout)
							})
						})
					}),
				)
			})
			stop := mac.Stop()
			visibleHeight := float32(d.Size.Y)
			offsetStack := op.Offset(image.Pt(0, int(visibleHeight-(visibleHeight*progress)))).Push(gtx.Ops)
			fillColor := color.NRGBA{R: 61, G: 61, B: 61, A: 255}
			component.Rect{Color: fillColor, Size: d.Size, Radii: gtx.Dp(2)}.Layout(gtx)
			stop.Add(gtx.Ops)
			areaStack := clip.Rect(image.Rectangle{Max: d.Size}).Push(gtx.Ops)
			areaStack.Pop()
			offsetStack.Pop()
			return d
		}),
	)
	return d
}
