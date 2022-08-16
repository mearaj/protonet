package view

import (
	"gioui.org/layout"
	"gioui.org/text"
	"gioui.org/unit"
	"gioui.org/widget"
	"gioui.org/widget/material"
	"gioui.org/x/component"
	"github.com/mearaj/protonet/assets/fonts"
	. "github.com/mearaj/protonet/ui/fwk"
	"golang.org/x/exp/shiny/materialdesign/colornames"
	"golang.org/x/exp/shiny/materialdesign/icons"
	"image/color"
	"time"
)

type AuthAccount struct {
	Manager
	buttonNewAccount *IconButton
	*material.Theme
	*widget.Icon
	inActiveTh      *material.Theme
	iconCreateNewID *widget.Icon
	AuthFormView    View
	*ModalContent
}

//
//func NewAuthAccount(manager Manager) *AuthAccount {
//	acc := AuthAccount{Manager: manager, Theme: manager.Theme()}
//	acc.AuthFormView = NewAuthForm(manager, acc.onSuccess)
//	acc.ModalContent = NewModalContent(func() {
//		acc.Modal().Dismiss(nil)
//	})
//	return &acc
//}

func (na *AuthAccount) Layout(gtx Gtx) Dim {
	if na.Theme == nil {
		na.Theme = fonts.NewTheme()
	}
	if na.Icon == nil {
		na.Icon, _ = widget.NewIcon(icons.ActionAccountCircle)
	}
	if na.inActiveTh == nil {
		inActiveTh := *fonts.NewTheme()
		inActiveTh.ContrastBg = color.NRGBA(colornames.Grey500)
		na.inActiveTh = &inActiveTh
	}
	if na.iconCreateNewID == nil {
		na.iconCreateNewID, _ = widget.NewIcon(icons.ContentCreate)
	}
	if na.buttonNewAccount == nil {
		na.buttonNewAccount = &IconButton{
			Theme: na.Theme,
			Icon:  na.Icon,
			Text:  "Unlock Account",
		}
	}

	flex := layout.Flex{Axis: layout.Vertical, Spacing: layout.SpaceSides, Alignment: layout.Middle}
	gtx.Constraints.Min.Y = gtx.Constraints.Max.Y
	if na.buttonNewAccount.Button.Clicked() {
		na.Modal().Show(na.drawModalContent, nil, Animation{
			Duration: time.Millisecond * 250,
			State:    component.Invisible,
			Started:  time.Time{},
		})
	}
	d := flex.Layout(gtx,
		layout.Rigid(func(gtx Gtx) Dim {
			return DrawProtonetImageCenter(gtx, na.Theme)
		}),
		layout.Rigid(layout.Spacer{Height: unit.Dp(16)}.Layout),
		layout.Rigid(func(gtx Gtx) Dim {
			return layout.Center.Layout(gtx, func(gtx Gtx) Dim {
				bdy := material.Body1(na.Theme, "Private key required for access")
				bdy.Alignment = text.Middle
				bdy.Font.Weight = text.Black
				bdy.Color = color.NRGBA{R: 102, G: 117, B: 127, A: 255}
				return bdy.Layout(gtx)
			})
		}),
		layout.Rigid(layout.Spacer{Height: unit.Dp(16)}.Layout),
		layout.Rigid(func(gtx Gtx) Dim {
			return layout.Flex{Spacing: layout.SpaceSides}.Layout(gtx, layout.Rigid(func(gtx layout.Context) layout.Dimensions {
				gtx.Constraints.Max.X = gtx.Dp(250)
				return na.buttonNewAccount.Layout(gtx)
			}))
		}),
		layout.Rigid(layout.Spacer{Height: unit.Dp(16)}.Layout),
	)
	return d
}

func (na *AuthAccount) drawModalContent(gtx Gtx) Dim {
	gtx.Constraints.Max.X = int(float32(gtx.Constraints.Max.X) * 0.85)
	gtx.Constraints.Max.Y = int(float32(gtx.Constraints.Max.Y) * 0.85)
	return na.ModalContent.DrawContent(gtx, na.Theme, na.AuthFormView.Layout)
}

func (na *AuthAccount) onSuccess() {
	na.Modal().Dismiss(func() {
		na.Manager.NavigateToUrl(ChatPageURL, nil)
	})
}
