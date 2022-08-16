package view

import (
	"errors"
	"gioui.org/layout"
	"gioui.org/unit"
	"gioui.org/widget"
	"gioui.org/widget/material"
	"gioui.org/x/component"
	"github.com/mearaj/protonet/assets/fonts"
	"github.com/mearaj/protonet/service"
	. "github.com/mearaj/protonet/ui/fwk"
	"golang.org/x/exp/shiny/materialdesign/icons"
	"strings"
)

type AccountDetails struct {
	*material.Theme
	buttonCopyPvtKey        IconButton
	buttonCopyPubKey        IconButton
	buttonPrivateKeyVisible IconButton
	buttonPrivateKeyHidden  IconButton
	inputPassword           *component.TextField
	inputPvtKey             *component.TextField
	inputPubKey             *component.TextField
	Account                 service.Account
	inputPasswordStr        string
	inputPvtKeyStr          string
	Manager
}

func NewAccountDetails(manager Manager, account service.Account) *AccountDetails {
	iconCopy, _ := widget.NewIcon(icons.ContentContentCopy)
	iconVisible, _ := widget.NewIcon(icons.ActionVisibility)
	iconHidden, _ := widget.NewIcon(icons.ActionVisibilityOff)
	accountDetails := AccountDetails{
		Theme:         manager.Theme(),
		Account:       account,
		Manager:       manager,
		inputPvtKey:   &component.TextField{Editor: widget.Editor{SingleLine: false}},
		inputPubKey:   &component.TextField{Editor: widget.Editor{SingleLine: false}},
		inputPassword: &component.TextField{Editor: widget.Editor{SingleLine: false}},
		buttonCopyPvtKey: IconButton{
			Theme: manager.Theme(),
			Icon:  iconCopy,
			Text:  "Copy Private Key",
		},
		buttonCopyPubKey: IconButton{
			Theme: manager.Theme(),
			Icon:  iconCopy,
			Text:  "Copy Public Key",
		},
		buttonPrivateKeyVisible: IconButton{
			Theme: manager.Theme(),
			Icon:  iconVisible,
			Text:  "Hide Private Key",
		},
		buttonPrivateKeyHidden: IconButton{
			Theme: manager.Theme(),
			Icon:  iconHidden,
			Text:  "Show Private Key",
		},
	}

	accountDetails.inputPubKey.SetText(account.PublicKey)
	return &accountDetails
}

func (i *AccountDetails) Layout(gtx Gtx) (d Dim) {
	if i.Theme == nil {
		i.Theme = fonts.NewTheme()
	}
	if i.inputPassword.Text() != i.inputPasswordStr {
		i.inputPassword.ClearError()
	}
	i.inputPasswordStr = i.inputPassword.Text()
	publicKey := i.Account.PublicKey

	inset := layout.UniformInset(unit.Dp(16))
	if i.buttonCopyPvtKey.Button.Clicked() {
		i.Manager.Window().WriteClipboard(i.inputPvtKey.Text())
	}
	if i.buttonCopyPubKey.Button.Clicked() {
		i.Manager.Window().WriteClipboard(publicKey)
	}

	if strings.TrimSpace(i.inputPvtKey.Text()) != strings.TrimSpace(i.inputPvtKeyStr) {
		i.inputPvtKey.SetText(i.inputPvtKeyStr)
	}
	if strings.TrimSpace(i.inputPubKey.Text()) != strings.TrimSpace(publicKey) {
		i.inputPubKey.SetText(publicKey)
	}

	labelPasswordText := "Enter Password"
	if i.buttonPrivateKeyHidden.Button.Clicked() {
		var err error
		if strings.TrimSpace(i.inputPassword.Text()) == "" {
			err = errors.New("password is empty")
			i.inputPassword.SetError(err.Error())
			i.inputPvtKeyStr = ""
			i.inputPvtKey.SetText("")
		} else {
			i.inputPvtKeyStr, err = i.Account.PrivateKey(i.inputPasswordStr)
			i.inputPvtKey.SetText(i.inputPvtKeyStr)
			if err != nil {
				i.inputPvtKeyStr = ""
				i.inputPvtKey.SetText("")
				i.inputPassword.SetError(err.Error())
			}
		}
	}
	if i.buttonPrivateKeyVisible.Button.Clicked() {
		i.inputPvtKeyStr = ""
		i.inputPvtKey.SetText("")
	}

	gtx.Constraints.Min.X = gtx.Constraints.Max.X
	return inset.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
		return layout.Flex{Axis: layout.Vertical, Alignment: layout.Middle}.Layout(gtx,
			layout.Rigid(func(gtx layout.Context) layout.Dimensions {
				flex := layout.Flex{Alignment: layout.Middle, Spacing: layout.SpaceBetween}
				return flex.Layout(gtx,
					layout.Rigid(func(gtx layout.Context) layout.Dimensions {
						gtx.Constraints.Max.X = gtx.Constraints.Max.X - gtx.Dp(200)
						th := *i.Theme
						origSize := th.TextSize
						if strings.TrimSpace(i.inputPassword.Text()) == "" && !i.inputPassword.Focused() {
							th.TextSize = unit.Sp(12)
						} else {
							th.TextSize = origSize
						}
						return i.inputPassword.Layout(gtx, &th, labelPasswordText)
					}),
					layout.Rigid(func(gtx layout.Context) layout.Dimensions {
						btn := &i.buttonPrivateKeyHidden
						if i.inputPvtKey.Text() != "" {
							btn = &i.buttonPrivateKeyVisible
						}
						return btn.Layout(gtx)
					}),
				)
			}),
			layout.Rigid(layout.Spacer{Height: unit.Dp(8)}.Layout),
			layout.Rigid(func(gtx Gtx) Dim {
				gtx.Constraints.Min.X = gtx.Constraints.Max.X
				return i.inputPvtKey.Layout(gtx, i.Theme, "Your Private Key")
			}),
			layout.Rigid(layout.Spacer{Height: unit.Dp(8)}.Layout),
			layout.Rigid(func(gtx layout.Context) layout.Dimensions {
				return layout.Center.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
					return i.buttonCopyPvtKey.Layout(gtx)
				})
			}),
			layout.Rigid(layout.Spacer{Height: unit.Dp(8)}.Layout),
			layout.Rigid(func(gtx Gtx) Dim {
				if strings.TrimSpace(publicKey) == "" {
					return Dim{}
				}
				gtx.Constraints.Min.X = gtx.Constraints.Max.X
				return i.inputPubKey.Layout(gtx, i.Theme, "Your Public Key")
			}),
			layout.Rigid(layout.Spacer{Height: unit.Dp(8)}.Layout),
			layout.Rigid(func(gtx layout.Context) layout.Dimensions {
				if strings.TrimSpace(publicKey) == "" {
					return Dim{}
				}
				return layout.Center.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
					return i.buttonCopyPubKey.Layout(gtx)
				})
			}),
		)
	})
}
