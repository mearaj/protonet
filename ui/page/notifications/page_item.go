package notifications

import (
	"gioui.org/io/pointer"
	"gioui.org/layout"
	"gioui.org/op"
	"gioui.org/text"
	"gioui.org/unit"
	"gioui.org/widget"
	"gioui.org/widget/material"
	"gioui.org/x/component"
	"github.com/mearaj/protonet/service"
	. "github.com/mearaj/protonet/ui/fwk"
	"github.com/mearaj/protonet/ui/view"
	"golang.org/x/exp/shiny/materialdesign/colornames"
	"golang.org/x/exp/shiny/materialdesign/icons"
	"image"
	"image/color"
	"strings"
	"time"
)

type pageItem struct {
	*material.Theme
	widget.Clickable
	buttonIconMore        widget.Clickable
	btnSetCurrentIdentity widget.Clickable
	shouldCloseMenuItems  bool
	buttonIconMoreDim     Dim
	Manager
	service.Contact
	PressedStamp int64
	view.AvatarView
	iconMore                *widget.Icon
	menuVisibilityAnim      component.VisibilityAnimation
	settingAsPrimaryAccount bool
}

func (i *pageItem) Layout(gtx Gtx) Dim {
	if i.Theme == nil {
		i.Theme = i.Manager.Theme()
	}
	return i.layoutContent(gtx)
}
func (i *pageItem) layoutContent(gtx Gtx) Dim {
	if i.menuVisibilityAnim == (component.VisibilityAnimation{}) {
		i.menuVisibilityAnim = component.VisibilityAnimation{
			Duration: time.Millisecond * 250,
			State:    component.Invisible,
			Started:  time.Time{},
		}
	}

	for _, e := range gtx.Events(i) {
		switch e := e.(type) {
		case pointer.Event:
			switch e.Type {
			case pointer.Leave:
				if i.menuVisibilityAnim.Revealed(gtx) == 1 {
					i.menuVisibilityAnim.Disappear(gtx.Now)
				}
			}
		}
	}
	pointer.InputOp{Tag: i, Types: pointer.Leave}.Add(gtx.Ops)

	if i.buttonIconMore.Clicked() {
		i.menuVisibilityAnim.Appear(gtx.Now)
		i.shouldCloseMenuItems = false
	}

	if i.btnSetCurrentIdentity.Pressed() {
		i.shouldCloseMenuItems = false
	}

	if i.btnSetCurrentIdentity.Clicked() {
		i.shouldCloseMenuItems = false
		if !i.settingAsPrimaryAccount {
			i.settingAsPrimaryAccount = true
			// i.menuVisibilityAnim.State = component.Invisible
			go func() {
				i.Manager.Modal().Show(func(gtx layout.Context) layout.Dimensions {
					loader := view.Loader{}
					return loader.Layout(gtx)
				}, nil, Animation{
					Duration: time.Millisecond * 250,
					State:    component.Invisible,
					Started:  time.Time{},
				})
				//i.Manager.Service().SetAsCurrentAccount(i.Contact.PublicKey)
				i.settingAsPrimaryAccount = false
				i.Manager.Modal().Dismiss(nil)
			}()
		}
	}

	if i.Clickable.Pressed() {
		if i.PressedStamp == 0 {
			i.PressedStamp = time.Now().UnixMilli()
		}
	}
	if i.Clickable.Clicked() {
		if i.PressedStamp != 0 {
			diff := time.Now().UnixMilli() - i.PressedStamp
			if diff > 500 {
				i.Selected = !i.Selected
			}
		}
		i.PressedStamp = 0
	}

	btnStyle := material.ButtonLayoutStyle{Background: i.Theme.ContrastBg, Button: &i.Clickable}

	if i.Selected || i.Clickable.Hovered() {
		btnStyle.Background.A = 50
	} else {
		btnStyle.Background.A = 10
	}
	if i.AvatarView.Theme == nil {
		i.AvatarView.Theme = i.Theme
	}

	if i.shouldCloseMenuItems && !i.menuVisibilityAnim.Animating() {
		i.menuVisibilityAnim.Disappear(gtx.Now)
	}
	for _, e := range gtx.Events(i.Window()) {
		switch e := e.(type) {
		case pointer.Event:
			switch e.Type {
			case pointer.Press:
				i.shouldCloseMenuItems = true
			}
		}
	}

	gtx.Constraints.Min.X = gtx.Constraints.Max.X
	d := btnStyle.Layout(gtx, func(gtx Gtx) Dim {
		inset := layout.Inset{Top: unit.Dp(16), Bottom: unit.Dp(16), Left: unit.Dp(16), Right: unit.Dp(8)}
		d := inset.Layout(gtx, func(gtx Gtx) Dim {
			flex := layout.Flex{Spacing: layout.SpaceEnd, Alignment: layout.Start}
			d := flex.Layout(gtx,
				layout.Rigid(func(gtx Gtx) Dim {
					flex := layout.Flex{Spacing: layout.SpaceSides, Alignment: layout.Start, Axis: layout.Vertical}
					d := flex.Layout(gtx, layout.Rigid(i.AvatarView.Layout))
					return d
				}),
				layout.Rigid(func(gtx Gtx) Dim {
					gtx.Constraints.Max.X = gtx.Constraints.Max.X - gtx.Dp(80)
					gtx.Constraints.Min.X = gtx.Constraints.Max.X
					flex := layout.Flex{Spacing: layout.SpaceSides, Alignment: layout.Start, Axis: layout.Vertical}
					inset := layout.Inset{Right: unit.Dp(16), Left: unit.Dp(16)}
					d := inset.Layout(gtx, func(gtx Gtx) Dim {
						d := flex.Layout(gtx,
							layout.Rigid(func(gtx Gtx) Dim {
								b := material.Body1(i.Theme, i.Contact.PublicKey)
								b.Font.Weight = text.Bold
								return b.Layout(gtx)
							}),
							layout.Rigid(func(gtx Gtx) Dim {
								b := material.Body1(i.Theme, strings.Trim(string(i.Contact.PublicKey), "\n"))
								b.Color = color.NRGBA(colornames.Grey600)
								return b.Layout(gtx)
							}),
						)
						return d
					})
					return d
				}),
				layout.Rigid(func(gtx Gtx) Dim {
					flex := layout.Flex{Axis: layout.Vertical, Spacing: layout.SpaceBetween, Alignment: layout.Middle}
					return flex.Layout(gtx, layout.Rigid(func(gtx layout.Context) layout.Dimensions {
						if i.iconMore == nil {
							i.iconMore, _ = widget.NewIcon(icons.NavigationMoreVert)
						}
						button := material.IconButton(i.Theme, &i.buttonIconMore, i.iconMore, "Vertical Icon For Options")
						button.Size = unit.Dp(24)
						button.Background = color.NRGBA{}

						button.Color = i.Theme.ContrastBg
						button.Inset = layout.UniformInset(unit.Dp(16))
						i.buttonIconMoreDim = button.Layout(gtx)
						layout.Stack{Alignment: layout.NE}.Layout(gtx,
							layout.Stacked(func(gtx Gtx) Dim {
								progress := i.menuVisibilityAnim.Revealed(gtx)
								gtx.Constraints.Max.X = int(float32(gtx.Dp(300)) * progress)
								gtx.Constraints.Min.X = gtx.Constraints.Max.X
								posX := -gtx.Dp(46)
								posY := gtx.Dp(32)
								ops := op.Offset(image.Point{X: posX, Y: posY}).Push(gtx.Ops)
								macro := op.Record(gtx.Ops)
								d := i.drawMenuItems(gtx)
								d.Size.Y = int(float32(d.Size.Y) * progress)
								call := macro.Stop()
								component.Rect{Size: d.Size, Color: i.Theme.ContrastBg}.Layout(gtx)
								call.Add(gtx.Ops)
								ops.Pop()
								return d
							}),
						)
						return i.buttonIconMoreDim
					}),
						layout.Rigid(func(gtx layout.Context) layout.Dimensions {
							a := i.Manager.Service().Account()
							if a.PublicKey == i.PublicKey &&
								i.menuVisibilityAnim.Revealed(gtx) != 1 {
								icon, _ := widget.NewIcon(icons.ActionCheckCircle)
								return icon.Layout(gtx, i.Theme.ContrastBg)
							}
							return Dim{}
						}),
					)
				}),
			)
			return d
		})
		return d
	})

	return d
}

func (i *pageItem) drawMenuItems(gtx Gtx) Dim {
	inset := layout.UniformInset(unit.Dp(12))
	return layout.Flex{Axis: layout.Vertical}.Layout(gtx,
		layout.Rigid(func(gtx Gtx) Dim {
			btnStyle := material.ButtonLayoutStyle{Button: &i.btnSetCurrentIdentity}
			return btnStyle.Layout(gtx,
				func(gtx Gtx) Dim {
					inset := inset
					return inset.Layout(gtx, func(gtx Gtx) Dim {
						return layout.Flex{Spacing: layout.SpaceEnd}.Layout(gtx,
							layout.Flexed(1, func(gtx Gtx) Dim {
								if i.menuVisibilityAnim.Animating() {
									return Dim{}
								}
								bd := material.Body1(i.Theme, "Set as current account")
								bd.Color = i.Theme.ContrastFg
								bd.Alignment = text.Start
								return bd.Layout(gtx)
							}),
						)
					})
				},
			)
		}),
	)
}
