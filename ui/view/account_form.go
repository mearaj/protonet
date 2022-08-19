package view

import (
	"gioui.org/io/clipboard"
	"gioui.org/layout"
	"gioui.org/op"
	"gioui.org/op/clip"
	"gioui.org/op/paint"
	"gioui.org/text"
	"gioui.org/unit"
	"gioui.org/widget"
	"gioui.org/widget/material"
	"github.com/mearaj/protonet/assets/fonts"
	. "github.com/mearaj/protonet/ui/fwk"
	"golang.org/x/exp/shiny/materialdesign/colornames"
	"golang.org/x/exp/shiny/materialdesign/icons"
	"image"
	"image/color"
	"strings"
)

type accountForm struct {
	Manager
	Theme                 *material.Theme
	InActiveTheme         *material.Theme
	iconCreateNewID       *widget.Icon
	iconImportFile        *widget.Icon
	pvtKeyStr             string
	title                 string
	importLabelText       string
	btnClear              IconButton
	btnNewID              IconButton
	btnSubmitImportKey    IconButton
	btnPasteKey           IconButton
	navigationIcon        *widget.Icon
	iDDetailsView         AccountDetails
	errorCreateNewID      error
	errorCreateNewIDChan  chan error
	errorImportKey        error
	errorImportKeyChan    chan error
	creatingNewID         bool
	submittingImportedKey bool
	OnSuccess             func()
	*ModalContent
}

func NewAccountFormView(manager Manager, OnSuccess func()) View {
	clearIcon, _ := widget.NewIcon(icons.ContentClear)
	navIcon, _ := widget.NewIcon(icons.NavigationArrowBack)
	iconCreateNewID, _ := widget.NewIcon(icons.ActionDone)
	iconImportFile, _ := widget.NewIcon(icons.FileFileUpload)
	pasteIcon, _ := widget.NewIcon(icons.ContentContentPaste)
	th := manager.Theme()
	errorTh := *fonts.NewTheme()
	errorTh.ContrastBg = color.NRGBA(colornames.Red500)
	inActiveTh := *fonts.NewTheme()
	inActiveTh.ContrastBg = color.NRGBA(colornames.Grey500)
	s := accountForm{
		Manager:              manager,
		Theme:                th,
		InActiveTheme:        &inActiveTh,
		title:                "Account",
		navigationIcon:       navIcon,
		iconCreateNewID:      iconCreateNewID,
		iconImportFile:       iconImportFile,
		importLabelText:      "Import Key",
		OnSuccess:            OnSuccess,
		errorCreateNewIDChan: make(chan error, 1),
		errorImportKeyChan:   make(chan error, 1),
		btnSubmitImportKey: IconButton{
			Theme: th,
			Icon:  iconCreateNewID,
			Text:  "Submit",
		},
		btnPasteKey: IconButton{
			Theme: th,
			Icon:  pasteIcon,
			Text:  "Paste",
		},
		btnNewID: IconButton{
			Theme: th,
			Icon:  iconCreateNewID,
			Text:  "Auto Create New Account",
		},
		btnClear: IconButton{
			Theme: th,
			Icon:  clearIcon,
			Text:  "Clear",
		},
		iDDetailsView: AccountDetails{
			Theme:   th,
			Manager: manager,
		},
	}
	s.ModalContent = NewModalContent(func() {
		s.Modal().Dismiss(nil)
		s.creatingNewID = false
		s.submittingImportedKey = false
		if s.OnSuccess != nil {
			s.OnSuccess()
		}
	})
	return &s
}

func (p *accountForm) Layout(gtx Gtx) Dim {
	if p.Theme == nil {
		p.Theme = fonts.NewTheme()
	}

	inset := layout.UniformInset(unit.Dp(16))
	flex := layout.Flex{Axis: layout.Vertical, Alignment: layout.Start}
	d := flex.Layout(gtx,
		layout.Rigid(func(gtx Gtx) Dim {
			inset := inset
			return inset.Layout(gtx, p.drawImportKeyTextField)
		}),
		layout.Rigid(func(gtx Gtx) Dim {
			gtx.Constraints.Min.X = gtx.Constraints.Max.X
			bd := material.Body1(p.Theme, "Or")
			bd.Font.Weight = text.Bold
			bd.Alignment = text.Middle
			bd.TextSize = unit.Sp(20)
			return bd.Layout(gtx)
		}),
		layout.Rigid(func(gtx Gtx) Dim {
			inset := inset
			return inset.Layout(gtx, p.drawAutoCreateField)
		}),
	)
	if p.creatingNewID || p.submittingImportedKey {
		layout.Stack{}.Layout(gtx,
			layout.Stacked(func(gtx layout.Context) layout.Dimensions {
				loader := Loader{}
				gtx.Constraints.Max, gtx.Constraints.Min = d.Size, d.Size
				return layout.Flex{Alignment: layout.Middle, Spacing: layout.SpaceSides}.Layout(gtx,
					layout.Rigid(func(gtx Gtx) Dim {
						return loader.Layout(gtx)
					}))
			}),
		)

		select {
		case p.errorCreateNewID = <-p.errorCreateNewIDChan:
			p.creatingNewID = false
			if p.errorCreateNewID == nil {
				// p.account.PrivateKey = p.Service().Account().PrivateKey
				if p.OnSuccess != nil {
					p.OnSuccess()
				}
			} else {
				p.Snackbar().Show(p.errorCreateNewID.Error(), &widget.Clickable{}, color.NRGBA{}, "")
				p.errorCreateNewID = nil
			}
		default:
		}

		select {
		case p.errorImportKey = <-p.errorImportKeyChan:
			p.submittingImportedKey = false
			if p.errorImportKey == nil {
				//p.account.PrivateKey = p.Service().Account().PrivateKey
				if p.OnSuccess != nil {
					p.OnSuccess()
				}
			}
		default:
		}

		return d
	}
	return d
}

func (p *accountForm) drawImportKeyTextField(gtx Gtx) Dim {
	if p.btnPasteKey.Button.Clicked() {
		clipboard.ReadOp{Tag: &p.btnPasteKey}.Add(gtx.Ops)
	}
	for _, e := range gtx.Events(&p.btnPasteKey) {
		switch e := e.(type) {
		case clipboard.Event:
			_ = e
			p.pvtKeyStr = e.Text
			// Clear the clipboard
			clipboard.WriteOp{Text: ""}.Add(gtx.Ops)
			p.errorImportKey = nil
		}
	}

	if p.btnClear.Button.Clicked() {
		p.pvtKeyStr = ""
		p.errorImportKey = nil
	}

	if p.btnSubmitImportKey.Button.Clicked() && !p.submittingImportedKey {
		p.submittingImportedKey = true
		p.createAccountFromPvtKeyHexStr()
	}
	flex := layout.Flex{Axis: layout.Vertical}
	return flex.Layout(gtx,
		layout.Rigid(func(gtx Gtx) Dim {
			gtx.Constraints.Min.X = gtx.Constraints.Max.X
			var txt string
			txt = strings.TrimSpace(p.pvtKeyStr)
			txtColor := p.Theme.Fg
			if txt == "" {
				txt = "Paste key file contents here"
				txtColor = color.NRGBA(colornames.Grey500)
			}
			if p.errorImportKey != nil {
				txt = p.errorImportKey.Error()
				txtColor = color.NRGBA(colornames.Red500)
			}
			inset := layout.UniformInset(unit.Dp(16))
			mac := op.Record(gtx.Ops)
			d := inset.Layout(gtx,
				func(gtx layout.Context) layout.Dimensions {
					lbl := material.Label(p.Theme, p.Theme.TextSize, txt)
					lbl.MaxLines = 10
					lbl.Color = txtColor
					return lbl.Layout(gtx)
				})
			stop := mac.Stop()
			bounds := image.Rect(0, 0, d.Size.X, d.Size.Y)
			rect := clip.UniformRRect(bounds, gtx.Dp(4))
			paint.FillShape(gtx.Ops,
				p.Theme.Fg,
				clip.Stroke{Path: rect.Path(gtx.Ops), Width: float32(gtx.Dp(1))}.Op(),
			)
			stop.Add(gtx.Ops)
			return d
		}),
		layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			gtx.Constraints.Min.X = gtx.Constraints.Max.X
			mobileWidth := gtx.Dp(350)
			flex := layout.Flex{Spacing: layout.SpaceBetween}
			spacerLayout := layout.Spacer{Width: unit.Dp(16)}
			submitLayout := layout.Flexed(1, func(gtx layout.Context) layout.Dimensions {
				return p.btnSubmitImportKey.Layout(gtx)
			})
			pasteLayout := layout.Flexed(1, func(gtx Gtx) Dim {
				return p.btnPasteKey.Layout(gtx)
			})
			clearLayout := layout.Flexed(1, func(gtx Gtx) Dim {
				return p.btnClear.Layout(gtx)
			})
			if gtx.Constraints.Max.X <= mobileWidth {
				flex.Axis = layout.Vertical
				spacerLayout.Width = 0
				spacerLayout.Height = 8
				submitLayout = layout.Rigid(func(gtx Gtx) Dim {
					return p.btnSubmitImportKey.Layout(gtx)
				})
				pasteLayout = layout.Rigid(func(gtx Gtx) Dim {
					return p.btnPasteKey.Layout(gtx)
				})
				clearLayout = layout.Rigid(func(gtx Gtx) Dim {
					return p.btnClear.Layout(gtx)
				})
			}
			inset := layout.Inset{Top: unit.Dp(16)}
			return inset.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
				return flex.Layout(gtx,
					submitLayout,
					layout.Rigid(spacerLayout.Layout),
					pasteLayout,
					layout.Rigid(spacerLayout.Layout),
					clearLayout,
				)
			})
		}),
	)

}

func (p *accountForm) drawAutoCreateField(gtx Gtx) Dim {
	var button *IconButton
	if p.errorCreateNewID != nil {
		button = &IconButton{
			Theme: p.InActiveTheme,
			Icon:  p.iconCreateNewID,
			Text:  "Auto Create New Account",
		}
	} else {
		button = &p.btnNewID
	}
	if p.btnNewID.Button.Clicked() && !p.creatingNewID {
		p.creatingNewID = true
		p.autoCreateAccount()
	}
	gtx.Constraints.Min.X = gtx.Constraints.Max.X
	flex := layout.Flex{Axis: layout.Vertical}
	return flex.Layout(gtx,
		layout.Rigid(func(gtx Gtx) Dim {
			flex := layout.Flex{Spacing: layout.SpaceEnd}
			inset := layout.Inset{Top: unit.Dp(16)}
			gtx.Constraints.Min.X = gtx.Constraints.Max.X
			return inset.Layout(gtx, func(gtx Gtx) Dim {
				return flex.Layout(gtx,
					layout.Rigid(func(gtx Gtx) Dim {
						gtx.Constraints.Min.X = gtx.Constraints.Max.X
						return button.Layout(gtx)
					}),
				)
			})
		}),
	)
}

func (p *accountForm) createAccountFromPvtKeyHexStr() {
	p.submittingImportedKey = true
	go func() {
		p.errorImportKeyChan <- <-p.Service().CreateAccount(p.pvtKeyStr)
		p.Window().Invalidate()
	}()
}

func (p *accountForm) autoCreateAccount() {
	p.creatingNewID = true
	go func() {
		p.errorCreateNewIDChan <- <-p.Service().AutoCreateAccount()
		p.Window().Invalidate()
	}()
}
