package view

import (
	"gioui.org/layout"
	"gioui.org/op"
	"gioui.org/widget"
	"gioui.org/x/component"
	. "github.com/mearaj/protonet/ui/fwk"
	"image"
	"image/color"
	"time"
)

type appModal struct {
	onBackdropClick func()
	widget          layout.Widget
	btnBackdrop     widget.Clickable
	Animation       component.VisibilityAnimation
	afterDismiss    func()
}

type modalsStack struct {
	Modals []*appModal
}

func (s *modalsStack) Show(widget layout.Widget, onBackdropClickCallback func(), animation Animation) {
	if len(s.Modals) == 0 {
		s.Modals = make([]*appModal, 0)
	}
	modal := appModal{
		onBackdropClick: onBackdropClickCallback,
		widget:          widget,
		Animation:       animation,
	}
	s.Modals = append(s.Modals, &modal)
	s.Modals[len(s.Modals)-1].Show(widget)
}

func (s *modalsStack) Dismiss(afterDismiss func()) {
	stackSize := len(s.Modals)
	if stackSize > 0 {
		s.Modals[stackSize-1].Dismiss(afterDismiss)
	}
}

func (s *modalsStack) Layout(gtx Gtx) Dim {
	for _, modal := range s.Modals {
		modal.Layout(gtx)
	}
	return Dim{Size: gtx.Constraints.Max}
}

func NewModalStack() Modal {
	return &modalsStack{}
}

func (m *appModal) Show(widget layout.Widget) {
	m.widget = widget
	m.Animation.Appear(time.Now())
}

func (m *appModal) Layout(gtx Gtx) Dim {
	if m.btnBackdrop.Clicked() {
		if m.onBackdropClick != nil {
			m.onBackdropClick()
		} else {
			m.Animation.Disappear(gtx.Now)
		}
	}
	var finalPosY int
	d := layout.Stack{Alignment: layout.N}.Layout(gtx,
		layout.Stacked(func(gtx layout.Context) layout.Dimensions {
			return m.btnBackdrop.Layout(gtx,
				func(gtx layout.Context) layout.Dimensions {
					if m.Animation.Revealed(gtx) == 0 || m.widget == nil {
						return Dim{}
					}
					return component.Rect{Size: gtx.Constraints.Max, Color: color.NRGBA{A: 200}}.Layout(gtx)
				},
			)
		}),
		layout.Stacked(func(gtx layout.Context) layout.Dimensions {
			state := m.Animation.State
			progress := m.Animation.Revealed(gtx)
			switch {
			case state == component.Invisible, progress == 0, m.widget == nil:
				if m.afterDismiss != nil && !m.Animation.Animating() {
					m.afterDismiss()
					m.afterDismiss = nil
				}
				return Dim{}
			case state == component.Visible, state == component.Appearing, state == component.Disappearing:
				// record Widget's dimension
				macro := op.Record(gtx.Ops)
				clickable := widget.Clickable{}
				d := clickable.Layout(gtx, m.widget)
				call := macro.Stop()
				finalPosY = -d.Size.Y + int(float32((gtx.Constraints.Max.Y+d.Size.Y)/2)*progress)
				op.Offset(image.Point{
					X: 0,
					Y: finalPosY,
				}).Add(gtx.Ops)
				call.Add(gtx.Ops)
			}
			d := m.widget(gtx)
			return d
		}),
	)
	return d
}

func (m *appModal) Dismiss(afterDismiss func()) {
	m.Animation.Disappear(time.Now())
	m.afterDismiss = afterDismiss
}
