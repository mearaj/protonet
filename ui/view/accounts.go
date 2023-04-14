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
	"github.com/mearaj/protonet/alog"
	"github.com/mearaj/protonet/assets"
	"github.com/mearaj/protonet/internal/wallet"
	. "github.com/mearaj/protonet/ui/fwk"
	"golang.org/x/exp/shiny/materialdesign/colornames"
	"image"
	"image/color"
)

type accountsView struct {
	layout.List
	Manager
	theme                 *material.Theme
	title                 string
	accountsItems         []*accountsItem
	currentAccountLayout  layout.List
	enum                  widget.Enum
	accountChangeCallback func()
}

func NewAccountsView(manager Manager, accountChangeCallback func()) View {
	errorTh := *manager.Theme()
	errorTh.ContrastBg = color.NRGBA(colornames.Red500)
	p := accountsView{
		Manager:               manager,
		theme:                 manager.Theme(),
		title:                 "Accounts",
		List:                  layout.List{Axis: layout.Vertical},
		accountsItems:         []*accountsItem{},
		accountChangeCallback: accountChangeCallback,
	}
	return &p
}

func (p *accountsView) Layout(gtx Gtx) Dim {
	a, _ := wallet.GlobalWallet.Account()
	p.enum.Value = a.PublicKey
	flex := layout.Flex{Axis: layout.Vertical,
		Spacing:   layout.SpaceEnd,
		Alignment: layout.Start,
	}

	d := flex.Layout(gtx,
		layout.Rigid(p.drawIdentitiesItems),
	)

	return d
}

func (p *accountsView) drawIdentitiesItems(gtx Gtx) Dim {
	if p.isProcessingRequired() {
		accs, _ := wallet.GlobalWallet.Accounts()
		p.accountsItems = make([]*accountsItem, 0, len(accs))
		for _, userID := range accs {
			p.accountsItems = append(p.accountsItems, &accountsItem{
				Theme:   p.theme,
				Manager: p.Manager,
				Account: userID,
				Enum:    &p.enum,
			})
		}
	}
	return layout.Flex{Axis: layout.Vertical}.Layout(gtx,
		layout.Rigid(func(gtx Gtx) Dim {
			inset := layout.UniformInset(unit.Dp(16))
			return inset.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
				flex := layout.Flex{Alignment: layout.Middle}
				a, _ := wallet.GlobalWallet.Account()
				d := flex.Layout(gtx,
					layout.Rigid(func(gtx layout.Context) layout.Dimensions {
						var img image.Image
						var err error
						img, _, err = image.Decode(bytes.NewReader(a.PublicImage))
						if err != nil {
							alog.Logger().Errorln(err)
						}
						if img == nil {
							img = assets.AppIconImage
						}
						radii := gtx.Dp(24)
						gtx.Constraints.Max.X, gtx.Constraints.Max.Y = radii*2, radii*2
						bounds := image.Rect(0, 0, radii*2, radii*2)
						clipOp := clip.UniformRRect(bounds, radii).Push(gtx.Ops)
						imgOps := paint.NewImageOp(img)
						imgWidget := widget.Image{Src: imgOps, Fit: widget.Contain, Position: layout.Center, Scale: 0}
						d := imgWidget.Layout(gtx)
						clipOp.Pop()
						return d
					}),
					layout.Rigid(func(gtx Gtx) Dim {
						gtx.Constraints.Min.X = gtx.Constraints.Max.X
						return p.currentAccountLayout.Layout(gtx, 1, func(gtx layout.Context, index int) layout.Dimensions {
							flex := layout.Flex{Spacing: layout.SpaceSides, Alignment: layout.Start, Axis: layout.Vertical}
							inset := layout.Inset{Right: unit.Dp(8), Left: unit.Dp(8)}
							d := inset.Layout(gtx, func(gtx Gtx) Dim {
								d := flex.Layout(gtx,
									layout.Rigid(func(gtx Gtx) Dim {
										b := material.Body1(p.Theme(), a.PublicKey)
										b.Font.Weight = text.Bold
										return b.Layout(gtx)
									}),
									//layout.Rigid(func(gtx Gtx) Dim {
									//	b := material.Body1(p.Theme(), strings.Trim(string(p.currentAccount.Contents), "\n"))
									//	b.Color = color.NRGBA(colornames.Grey600)
									//	return b.Layout(gtx)
									//}),
								)
								return d
							})
							return d
						})
					}),
				)
				return d
			})
		}),
		layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			return p.List.Layout(gtx, len(p.accountsItems), func(gtx Gtx, index int) (d Dim) {
				accountItem := p.accountsItems[index]
				if accountItem.Clickable.Pressed() {
					_ = wallet.GlobalWallet.AddUpdateAccount(&accountItem.Account)
					if p.accountChangeCallback != nil {
						p.accountChangeCallback()
					}
				}
				return p.accountsItems[index].Layout(gtx)
			})
		}),
	)
}

// isProcessingRequired
func (p *accountsView) isProcessingRequired() bool {
	accs, _ := wallet.GlobalWallet.Accounts()
	isRequired := len(accs) != len(p.accountsItems)
	return isRequired
}
