package view

import (
	"gioui.org/io/clipboard"
	"gioui.org/io/key"
	"gioui.org/layout"
	"gioui.org/unit"
	"gioui.org/widget"
	"gioui.org/widget/material"
	"gioui.org/x/component"
	"github.com/mearaj/protonet/assets/fonts"
	"github.com/mearaj/protonet/service"
	. "github.com/mearaj/protonet/ui/fwk"
	"golang.org/x/exp/shiny/materialdesign/colornames"
	"golang.org/x/exp/shiny/materialdesign/icons"
	"image/color"
)

// contactForm Always call NewContactForm function to create contactForm
type contactForm struct {
	Manager
	Theme            *material.Theme
	inputNewChat     component.TextField
	inputNewChatStr  string
	buttonSubmit     IconButton
	buttonPasteKey   IconButton
	btnClear         IconButton
	errorNewChat     error
	errorNewChatChan chan error
	addingNewClient  bool
	contact          service.Contact
	OnSuccess        func(addr string)
	inActiveTheme    *material.Theme
}

// NewContactForm Always call this function to create contactForm
func NewContactForm(manager Manager, contact service.Contact, OnSuccess func(addr string)) *contactForm {
	iconSubmit, _ := widget.NewIcon(icons.ActionDone)
	inActiveTheme := fonts.NewTheme()
	inActiveTheme.ContrastBg = color.NRGBA(colornames.Grey500)
	iconPaste, _ := widget.NewIcon(icons.ContentContentPaste)
	iconClear, _ := widget.NewIcon(icons.ContentClear)
	contForm := contactForm{
		Manager:          manager,
		Theme:            manager.Theme(),
		contact:          contact,
		OnSuccess:        OnSuccess,
		inActiveTheme:    inActiveTheme,
		errorNewChatChan: make(chan error),
		buttonSubmit: IconButton{
			Theme: manager.Theme(),
			Icon:  iconSubmit,
			Text:  "Submit",
		},
		buttonPasteKey: IconButton{
			Theme: manager.Theme(),
			Icon:  iconPaste,
			Text:  "Paste",
		},
		btnClear: IconButton{
			Theme: manager.Theme(),
			Icon:  iconClear,
			Text:  "Clear",
		},
		inputNewChat: component.TextField{Editor: widget.Editor{
			SingleLine: true,
			Submit:     true,
			InputHint:  key.HintEmail,
		}},
	}
	contForm.inputNewChatStr = contact.PublicKey
	contForm.inputNewChat.SetText(contact.PublicKey)
	return &contForm
}

func (p *contactForm) Layout(gtx Gtx) Dim {
	if p.Theme == nil {
		p.Theme = fonts.NewTheme()
	}
	if p.inputNewChat.Text() != p.inputNewChatStr {
		p.errorNewChat = nil
		p.inputNewChat.ClearError()
	}
	p.inputNewChatStr = p.inputNewChat.Text()

	inset := layout.UniformInset(unit.Dp(16))
	flex := layout.Flex{Axis: layout.Vertical, Alignment: layout.Start}
	d := flex.Layout(gtx,
		layout.Rigid(func(gtx Gtx) Dim {
			return inset.Layout(gtx, p.drawNewChatTextField)
		}),
	)
	if p.addingNewClient {
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
		case p.errorNewChat = <-p.errorNewChatChan:
			p.addingNewClient = false
			p.Window().Invalidate()
			if p.errorNewChat == nil {
				addr := p.inputNewChat.Text()
				p.inputNewChat.ClearError()
				if p.OnSuccess != nil {
					p.OnSuccess(addr)
				}
			} else {
				p.inputNewChat.SetError(p.errorNewChat.Error())
			}
		default:
		}
		return d
	}
	return d
}

func (p *contactForm) drawNewChatTextField(gtx Gtx) Dim {
	labelHintText := "Enter/Paste contact's public address"

	if p.buttonPasteKey.Button.Clicked() {
		clipboard.ReadOp{Tag: &p.buttonPasteKey}.Add(gtx.Ops)
	}
	for _, e := range gtx.Events(&p.buttonPasteKey) {
		switch e := e.(type) {
		case clipboard.Event:
			p.inputNewChat.SetText(e.Text)
			// clear the clipboard
			clipboard.WriteOp{Text: ""}.Add(gtx.Ops)
			p.inputNewChat.ClearError()
			p.errorNewChat = nil
		}
	}
	if p.btnClear.Button.Clicked() {
		p.inputNewChat.SetText("")
		p.inputNewChat.ClearError()
		p.errorNewChat = nil
	}
	if p.buttonSubmit.Button.Clicked() && !p.addingNewClient {
		p.addingNewClient = true
		go func() {
			p.errorNewChatChan <- <-p.Service().SaveContact(p.inputNewChat.Text(), true)
		}()
	}
	gtx.Constraints.Min.X = gtx.Constraints.Max.X
	return layout.Flex{Axis: layout.Vertical}.Layout(gtx,
		layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			gtx.Constraints.Min.X = gtx.Constraints.Max.X
			return p.inputNewChat.Layout(gtx, p.Theme, labelHintText)
		}),
		layout.Rigid(func(gtx Gtx) Dim {
			gtx.Constraints.Min.X = gtx.Constraints.Max.X
			mobileWidth := gtx.Dp(350)
			flex := layout.Flex{Spacing: layout.SpaceBetween}
			spacerLayout := layout.Spacer{Width: unit.Dp(16)}
			submitLayout := layout.Flexed(1, func(gtx Gtx) Dim {
				return p.buttonSubmit.Layout(gtx)
			})
			pasteLayout := layout.Flexed(1, func(gtx Gtx) Dim {
				return p.buttonPasteKey.Layout(gtx)
			})
			clearLayout := layout.Flexed(1, func(gtx Gtx) Dim {
				return p.btnClear.Layout(gtx)
			})
			if gtx.Constraints.Max.X <= mobileWidth {
				flex.Axis = layout.Vertical
				spacerLayout.Width = 0
				spacerLayout.Height = 8
				submitLayout = layout.Rigid(func(gtx Gtx) Dim {
					return p.buttonSubmit.Layout(gtx)
				})
				pasteLayout = layout.Rigid(func(gtx Gtx) Dim {
					return p.buttonPasteKey.Layout(gtx)
				})
				clearLayout = layout.Rigid(func(gtx Gtx) Dim {
					return p.btnClear.Layout(gtx)
				})
			}
			inset := layout.Inset{Top: unit.Dp(16)}
			return inset.Layout(gtx, func(gtx Gtx) Dim {
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
