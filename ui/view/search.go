package view

import (
	"gioui.org/layout"
	"gioui.org/widget"
	"gioui.org/widget/material"
	"gioui.org/x/component"
	"github.com/mearaj/protonet/assets/fonts"
	"github.com/mearaj/protonet/ui/fwk"
	"golang.org/x/exp/shiny/materialdesign/icons"
	"time"
)

type Search struct {
	initialized bool
	widget.Clickable
	EditorAnimated
	Icon *widget.Icon
	*material.Theme
}

func (s *Search) Layout(gtx fwk.Gtx) fwk.Dim {
	if !s.initialized {
		if s.Icon == nil {
			icon, _ := widget.NewIcon(icons.ActionSearch)
			s.Icon = icon
		}
		if s.Theme == nil {
			s.Theme = fonts.NewTheme()
		}
		if s.Animation.Duration == time.Duration(0) {
			s.Animation.Duration = time.Millisecond * 150
			s.Animation.State = component.Invisible
		}
		s.EditorAnimated.SingleLine = true
		s.initialized = true
	}
	if s.Clicked() {
		if !s.Animating() {
			s.EditorAnimated.Focus()
		}
		s.Animation.ToggleVisibility(gtx.Now)
	}
	flex := layout.Flex{Alignment: layout.Middle}
	return flex.Layout(gtx,
		layout.Flexed(1, func(gtx layout.Context) layout.Dimensions {
			return s.EditorAnimated.Layout(gtx)
		}),
		layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			return material.IconButton(s.Theme, &s.Clickable, s.Icon, "").Layout(gtx)
		}),
	)
}
