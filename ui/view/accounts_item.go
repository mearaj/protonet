package view

import (
	"bytes"
	"gioui.org/layout"
	"gioui.org/op/clip"
	"gioui.org/op/paint"
	"gioui.org/text"
	"gioui.org/unit"
	"gioui.org/widget"
	"gioui.org/widget/material"
	"gioui.org/x/component"
	"github.com/mearaj/protonet/alog"
	"github.com/mearaj/protonet/assets"
	"github.com/mearaj/protonet/service"
	. "github.com/mearaj/protonet/ui/fwk"
	"golang.org/x/exp/shiny/materialdesign/icons"
	"image"
)

type accountsItem struct {
	*material.Theme
	widget.Clickable
	btnSetCurrentIdentity widget.Clickable
	Manager
	service.Account
	*widget.Enum
}

func (i *accountsItem) Layout(gtx Gtx) Dim {
	if i.Theme == nil {
		i.Theme = i.Manager.Theme()
	}
	return i.layoutContent(gtx)
}

func (i *accountsItem) IsSelected() bool {
	return i.Enum.Value == i.Account.PublicKey
}

func (i *accountsItem) layoutContent(gtx Gtx) Dim {
	if i.btnSetCurrentIdentity.Clicked() {
		i.Manager.Service().SetAsCurrentAccount(i.Account)
	}

	btnStyle := material.ButtonLayoutStyle{Background: i.Theme.ContrastBg, Button: &i.Clickable}

	if i.IsSelected() || i.Clickable.Hovered() {
		btnStyle.Background.A = 50
	} else {
		btnStyle.Background.A = 10
	}

	d := btnStyle.Layout(gtx, func(gtx Gtx) Dim {
		inset := layout.UniformInset(unit.Dp(16))
		d := inset.Layout(gtx, func(gtx Gtx) Dim {
			gtx.Constraints.Min.X = gtx.Constraints.Max.X
			flex := layout.Flex{Spacing: layout.SpaceBetween, Alignment: layout.Middle}
			d := flex.Layout(gtx,
				layout.Rigid(func(gtx Gtx) Dim {
					gtx.Constraints.Max.X = gtx.Constraints.Max.X - gtx.Dp(32)
					flex := layout.Flex{Alignment: layout.Middle}
					d := flex.Layout(gtx,
						layout.Rigid(func(gtx layout.Context) layout.Dimensions {
							var img image.Image
							var err error
							img, _, err = image.Decode(bytes.NewReader(i.Account.PublicImage))
							if err != nil {
								alog.Logger().Errorln(err)
							}
							if img == nil {
								img = assets.AppIconImage
							}
							radii := gtx.Dp(12)
							gtx.Constraints.Max.X, gtx.Constraints.Max.Y = radii*2, radii*2
							bounds := image.Rect(0, 0, radii*2, radii*2)
							clipOp := clip.UniformRRect(bounds, radii).Push(gtx.Ops)
							imgOps := paint.NewImageOp(img)
							imgWidget := widget.Image{Src: imgOps, Fit: widget.Contain, Position: layout.Center, Scale: 0}
							d := imgWidget.Layout(gtx)
							clipOp.Pop()
							return d
						}),
						layout.Rigid(layout.Spacer{Width: unit.Dp(8)}.Layout),
						layout.Rigid(func(gtx Gtx) Dim {
							label := material.Label(i.Theme, i.Theme.TextSize, i.Account.PublicKey)
							label.Font.Weight = text.Bold
							return component.TruncatingLabelStyle(label).Layout(gtx)
						}),
					)
					return d
				}),
				layout.Rigid(func(gtx layout.Context) layout.Dimensions {
					if i.IsSelected() {
						icon, _ := widget.NewIcon(icons.ToggleRadioButtonChecked)
						return icon.Layout(gtx, i.Theme.ContrastBg)
					}
					icon, _ := widget.NewIcon(icons.ToggleRadioButtonUnchecked)
					return icon.Layout(gtx, i.Theme.ContrastBg)
				}),
			)
			return d
		})
		return d
	})

	gtx.Constraints.Max.Y = d.Size.Y
	return d
}
