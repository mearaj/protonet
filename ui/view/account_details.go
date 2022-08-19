package view

import (
	"errors"
	"gioui.org/layout"
	"gioui.org/op"
	"gioui.org/op/clip"
	"gioui.org/op/paint"
	"gioui.org/unit"
	"gioui.org/widget"
	"gioui.org/widget/material"
	"gioui.org/x/component"
	"github.com/mearaj/protonet/assets/fonts"
	"github.com/mearaj/protonet/service"
	. "github.com/mearaj/protonet/ui/fwk"
	"golang.org/x/exp/shiny/materialdesign/colornames"
	"golang.org/x/exp/shiny/materialdesign/icons"
	"image"
	"image/color"
	"strings"
)

type AccountDetails struct {
	*material.Theme
	buttonCopyPvtKey        IconButton
	buttonCopyPubKey        IconButton
	buttonPrivateKeyVisible IconButton
	buttonPrivateKeyHidden  IconButton
	inputPassword           *component.TextField
	Account                 service.Account
	inputPasswordStr        string
	pvtKeyStr               string
	pvtKeyListLayout        layout.List
	pubKeyListLayout        layout.List
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

	return &accountDetails
}

func (ad *AccountDetails) Layout(gtx Gtx) Dim {
	if ad.Theme == nil {
		ad.Theme = fonts.NewTheme()
	}
	if ad.inputPassword.Text() != ad.inputPasswordStr {
		ad.inputPassword.ClearError()
	}
	ad.inputPasswordStr = ad.inputPassword.Text()

	inset := layout.UniformInset(unit.Dp(16))
	flex := layout.Flex{Axis: layout.Vertical, Alignment: layout.Start}
	d := flex.Layout(gtx,
		layout.Rigid(func(gtx Gtx) Dim {
			inset := inset
			return inset.Layout(gtx, ad.drawPasswordField)
		}),
		layout.Rigid(func(gtx Gtx) Dim {
			inset := inset
			return inset.Layout(gtx, ad.drawPvtKeyField)
		}),
		layout.Rigid(func(gtx Gtx) Dim {
			inset := inset
			return inset.Layout(gtx, ad.drawPubKeyField)
		}),
	)
	return d
}

func (ad *AccountDetails) drawPasswordField(gtx Gtx) Dim {
	if ad.buttonPrivateKeyHidden.Button.Clicked() {
		var err error
		if strings.TrimSpace(ad.inputPassword.Text()) == "" {
			err = errors.New("password is empty")
			ad.inputPassword.SetError(err.Error())
			ad.pvtKeyStr = ""
		} else {
			ad.pvtKeyStr, err = ad.Account.PrivateKey(ad.inputPasswordStr)
			if err != nil {
				ad.pvtKeyStr = ""
				ad.inputPassword.SetError(err.Error())
			}
		}
	}
	if ad.buttonPrivateKeyVisible.Button.Clicked() {
		ad.pvtKeyStr = ""
	}
	labelPasswordText := "Enter Password"
	flex := layout.Flex{Axis: layout.Vertical}
	return flex.Layout(gtx,
		layout.Rigid(func(gtx Gtx) Dim {
			th := *ad.Theme
			origSize := th.TextSize
			if strings.TrimSpace(ad.inputPassword.Text()) == "" && !ad.inputPassword.Focused() {
				th.TextSize = unit.Sp(12)
			} else {
				th.TextSize = origSize
			}
			return ad.inputPassword.Layout(gtx, &th, labelPasswordText)
		}),
		layout.Rigid(layout.Spacer{Height: unit.Dp(8)}.Layout),
		layout.Rigid(func(gtx Gtx) Dim {
			gtx.Constraints.Min.X = gtx.Constraints.Max.X
			btn := &ad.buttonPrivateKeyHidden
			if ad.pvtKeyStr != "" {
				btn = &ad.buttonPrivateKeyVisible
			}
			return btn.Layout(gtx)
		}),
	)
}

func (ad *AccountDetails) drawPvtKeyField(gtx Gtx) Dim {
	if ad.buttonCopyPvtKey.Button.Clicked() {
		ad.Manager.Window().WriteClipboard(ad.pvtKeyStr)
	}
	flex := layout.Flex{Axis: layout.Vertical}
	return flex.Layout(gtx,
		layout.Rigid(func(gtx Gtx) Dim {
			gtx.Constraints.Min.X = gtx.Constraints.Max.X
			var txt string
			txt = strings.TrimSpace(ad.pvtKeyStr)
			txtColor := ad.Theme.Fg
			if txt == "" {
				txt = "Your Private Key"
				txtColor = color.NRGBA(colornames.Grey500)
			}
			inset := layout.UniformInset(unit.Dp(16))
			mac := op.Record(gtx.Ops)
			d := inset.Layout(gtx,
				func(gtx Gtx) Dim {
					lbl := material.Label(ad.Theme, ad.Theme.TextSize, txt)
					lbl.Color = txtColor
					return ad.pvtKeyListLayout.Layout(gtx, 1, func(gtx layout.Context, index int) layout.Dimensions {
						return lbl.Layout(gtx)
					})
				})
			stop := mac.Stop()
			bounds := image.Rect(0, 0, d.Size.X, d.Size.Y)
			rect := clip.UniformRRect(bounds, gtx.Dp(4))
			paint.FillShape(gtx.Ops,
				ad.Theme.Fg,
				clip.Stroke{Path: rect.Path(gtx.Ops), Width: float32(gtx.Dp(1))}.Op(),
			)
			stop.Add(gtx.Ops)
			return d
		}),
		layout.Rigid(layout.Spacer{Height: unit.Dp(8)}.Layout),
		layout.Rigid(func(gtx Gtx) Dim {
			gtx.Constraints.Min.X = gtx.Constraints.Max.X
			return ad.buttonCopyPvtKey.Layout(gtx)
		}),
	)
}

func (ad *AccountDetails) drawPubKeyField(gtx Gtx) Dim {
	publicKey := ad.Account.PublicKey
	if ad.buttonCopyPubKey.Button.Clicked() {
		ad.Manager.Window().WriteClipboard(publicKey)
	}
	flex := layout.Flex{Axis: layout.Vertical}
	return flex.Layout(gtx,
		layout.Rigid(func(gtx Gtx) Dim {
			gtx.Constraints.Min.X = gtx.Constraints.Max.X
			var txt string
			txt = publicKey
			txtColor := ad.Theme.Fg
			if txt == "" {
				txt = "Your Public Key"
				txtColor = color.NRGBA(colornames.Grey500)
			}
			inset := layout.UniformInset(unit.Dp(16))
			mac := op.Record(gtx.Ops)
			d := inset.Layout(gtx,
				func(gtx Gtx) Dim {
					lbl := material.Label(ad.Theme, ad.Theme.TextSize, txt)
					lbl.Color = txtColor
					return ad.pubKeyListLayout.Layout(gtx, 1, func(gtx layout.Context, index int) layout.Dimensions {
						return lbl.Layout(gtx)
					})
				})
			stop := mac.Stop()
			bounds := image.Rect(0, 0, d.Size.X, d.Size.Y)
			rect := clip.UniformRRect(bounds, gtx.Dp(4))
			paint.FillShape(gtx.Ops,
				ad.Theme.Fg,
				clip.Stroke{Path: rect.Path(gtx.Ops), Width: float32(gtx.Dp(1))}.Op(),
			)
			stop.Add(gtx.Ops)
			return d
		}),
		layout.Rigid(layout.Spacer{Height: unit.Dp(8)}.Layout),
		layout.Rigid(func(gtx Gtx) Dim {
			gtx.Constraints.Min.X = gtx.Constraints.Max.X
			return ad.buttonCopyPubKey.Layout(gtx)
		}),
	)
}
