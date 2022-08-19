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
	"golang.org/x/exp/shiny/materialdesign/colornames"
	"golang.org/x/exp/shiny/materialdesign/icons"
	"image/color"
	"strings"
)

type passwordForm struct {
	Manager
	Theme                        *material.Theme
	inputPassword                component.TextField
	inputPasswordStr             string
	inputRepeatPassword          component.TextField
	inputRepeatPasswordStr       string
	buttonSubmit                 IconButton
	btnClearPassword             widget.Clickable
	btnClearRepeatPassword       widget.Clickable
	errorAuth                    error
	errorAuthChan                chan error
	authenticating               bool
	OnSuccess                    func()
	inActiveTheme                *material.Theme
	buttonShowHidePassword       widget.Clickable
	buttonShowHideRepeatPassword widget.Clickable
	isFetchingAccountsCount      bool
	fetchingAccountsCountCh      chan int64
	accountsCount                int64
	initialized                  bool
	layout.List
}

func NewPasswordForm(manager Manager, OnSuccess func()) *passwordForm {
	iconSubmit, _ := widget.NewIcon(icons.ActionDone)
	inActiveTheme := fonts.NewTheme()
	inActiveTheme.ContrastBg = color.NRGBA(colornames.Grey500)
	passForm := passwordForm{
		Manager:                 manager,
		Theme:                   manager.Theme(),
		OnSuccess:               OnSuccess,
		inActiveTheme:           inActiveTheme,
		errorAuthChan:           make(chan error),
		fetchingAccountsCountCh: make(chan int64, 10),
		buttonSubmit: IconButton{
			Theme: manager.Theme(),
			Icon:  iconSubmit,
			Text:  "Submit",
		},
	}
	return &passForm
}

func (p *passwordForm) Layout(gtx Gtx) Dim {
	if !p.initialized {
		p.List.Axis = layout.Vertical
		p.fetchAccountsCount()
		p.initialized = true
	}
	shouldBreak := false
	for {
		select {
		case p.accountsCount = <-p.fetchingAccountsCountCh:
			p.isFetchingAccountsCount = false
		default:
			shouldBreak = true
		}
		if shouldBreak {
			break
		}
	}

	if p.Theme == nil {
		p.Theme = fonts.NewTheme()
	}
	if p.inputPassword.Text() != p.inputPasswordStr ||
		p.inputRepeatPassword.Text() != p.inputRepeatPasswordStr {
		p.errorAuth = nil
		p.inputPassword.ClearError()
		p.inputRepeatPassword.ClearError()
	}

	p.inputPasswordStr = p.inputPassword.Text()
	p.inputRepeatPasswordStr = p.inputRepeatPassword.Text()

	inset := layout.UniformInset(unit.Dp(16))
	flex := layout.Flex{Axis: layout.Vertical, Alignment: layout.Start}
	d := flex.Layout(gtx,
		layout.Rigid(func(gtx Gtx) Dim {
			return inset.Layout(gtx, p.drawPasswordTextField)
		}),
	)
	if p.authenticating {
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
		case p.errorAuth = <-p.errorAuthChan:
			p.authenticating = false
			p.Window().Invalidate()
			if p.errorAuth == nil {
				p.inputPassword.ClearError()
				if p.OnSuccess != nil {
					p.OnSuccess()
				}
			} else {
				p.inputPassword.SetError(p.errorAuth.Error())
				p.inputRepeatPassword.SetError(p.errorAuth.Error())
			}
		default:
		}
		return d
	}
	return d
}

func (p *passwordForm) fetchAccountsCount() {
	if !p.isFetchingAccountsCount {
		p.isFetchingAccountsCount = true
		go func() {
			p.fetchingAccountsCountCh <- <-p.Service().AccountsCount()
			p.Window().Invalidate()
		}()
	}
}

func (p *passwordForm) drawPasswordTextField(gtx Gtx) Dim {
	labelPasswordText := "Set new password"
	labelRepeatPasswordText := "Re-enter password"
	if p.accountsCount > 0 {
		labelPasswordText = "Enter current password"
		labelRepeatPasswordText = "Re-enter password"
	}

	if p.btnClearPassword.Clicked() {
		p.inputPassword.SetText("")
		p.inputPassword.ClearError()
		p.errorAuth = nil
	}

	if p.btnClearRepeatPassword.Clicked() {
		p.inputRepeatPassword.SetText("")
		p.inputRepeatPassword.ClearError()
		p.errorAuth = nil
	}

	if p.buttonShowHidePassword.Clicked() {
		if p.inputPassword.Mask == '*' {
			p.inputPassword.Mask = '\x00'
		} else {
			p.inputPassword.Mask = '*'
		}
	}
	if p.buttonShowHideRepeatPassword.Clicked() {
		if p.inputRepeatPassword.Mask == '*' {
			p.inputRepeatPassword.Mask = '\x00'
		} else {
			p.inputRepeatPassword.Mask = '*'
		}
	}

	if p.buttonSubmit.Button.Clicked() && !p.authenticating {
		p.authenticating = true
		if strings.TrimSpace(p.inputPassword.Text()) != strings.TrimSpace(p.inputRepeatPassword.Text()) {
			go func() {
				p.errorAuthChan <- errors.New("Password mismatch!\n Please make sure password matches in both the inputs")
			}()
		} else {
			go func() {
				p.errorAuthChan <- <-p.Service().SetUserPassword(strings.TrimSpace(p.inputPassword.Text()))
			}()
		}
	}
	gtx.Constraints.Min = gtx.Constraints.Max
	return p.List.Layout(gtx, 1, func(gtx layout.Context, index int) layout.Dimensions {
		return layout.Flex{Axis: layout.Vertical, Alignment: layout.Middle, Spacing: layout.SpaceSides}.Layout(gtx,
			layout.Rigid(func(gtx Gtx) Dim {
				return DrawProtonetImageCenter(gtx, p.Theme)
			}),
			layout.Rigid(layout.Spacer{Height: unit.Dp(16)}.Layout),
			layout.Rigid(func(gtx layout.Context) layout.Dimensions {
				gtx.Constraints.Min.X = gtx.Constraints.Max.X
				return layout.Flex{Alignment: layout.Middle}.Layout(gtx,
					layout.Rigid(func(gtx layout.Context) layout.Dimensions {
						gtx.Constraints.Max.X = gtx.Constraints.Max.X - gtx.Dp(100)
						th := *p.Theme
						origSize := th.TextSize
						if strings.TrimSpace(p.inputPassword.Text()) == "" && !p.inputPassword.Focused() {
							th.TextSize = unit.Sp(12)
						} else {
							th.TextSize = origSize
						}
						return p.inputPassword.Layout(gtx, &th, labelPasswordText)
					}),
					layout.Rigid(layout.Spacer{Width: unit.Dp(16)}.Layout),
					layout.Rigid(func(gtx layout.Context) layout.Dimensions {
						icon, _ := widget.NewIcon(icons.ActionVisibility)
						if p.inputPassword.Editor.Mask == '*' {
							icon, _ = widget.NewIcon(icons.ActionVisibilityOff)
						}
						btn := material.IconButton(p.Theme,
							&p.buttonShowHidePassword, icon, "Show/Hide Password")
						btn.Size = unit.Dp(25)
						btn.Inset = layout.Inset{}
						return btn.Layout(gtx)
					}),
					layout.Rigid(layout.Spacer{Width: unit.Dp(16)}.Layout),
					layout.Rigid(func(gtx layout.Context) layout.Dimensions {
						clearIcon, _ := widget.NewIcon(icons.ContentClear)
						btn := material.IconButton(p.Theme,
							&p.btnClearPassword, clearIcon, "Clear Password")
						btn.Size = unit.Dp(25)
						btn.Inset = layout.Inset{}
						return btn.Layout(gtx)
					}),
				)
			}),
			layout.Rigid(func(gtx layout.Context) layout.Dimensions {
				gtx.Constraints.Min.X = gtx.Constraints.Max.X
				return layout.Flex{Alignment: layout.Middle}.Layout(gtx,
					layout.Rigid(func(gtx layout.Context) layout.Dimensions {
						gtx.Constraints.Max.X = gtx.Constraints.Max.X - gtx.Dp(100)
						th := *p.Theme
						origSize := th.TextSize
						if strings.TrimSpace(p.inputRepeatPassword.Text()) == "" && !p.inputRepeatPassword.Focused() {
							th.TextSize = unit.Sp(12)
						} else {
							th.TextSize = origSize
						}
						return p.inputRepeatPassword.Layout(gtx, &th, labelRepeatPasswordText)
					}),
					layout.Rigid(layout.Spacer{Width: unit.Dp(16)}.Layout),
					layout.Rigid(func(gtx layout.Context) layout.Dimensions {
						icon, _ := widget.NewIcon(icons.ActionVisibility)
						if p.inputRepeatPassword.Editor.Mask == '*' {
							icon, _ = widget.NewIcon(icons.ActionVisibilityOff)
						}
						btn := material.IconButton(p.Theme,
							&p.buttonShowHideRepeatPassword, icon, "Show/Hide Password")
						btn.Size = unit.Dp(25)
						btn.Inset = layout.Inset{}
						return btn.Layout(gtx)
					}),
					layout.Rigid(layout.Spacer{Width: unit.Dp(16)}.Layout),
					layout.Rigid(func(gtx layout.Context) layout.Dimensions {
						clearIcon, _ := widget.NewIcon(icons.ContentClear)
						btn := material.IconButton(p.Theme,
							&p.btnClearRepeatPassword, clearIcon, "Clear Password")
						btn.Size = unit.Dp(25)
						btn.Inset = layout.Inset{}
						return btn.Layout(gtx)
					}),
				)
			}),
			layout.Rigid(func(gtx Gtx) Dim {
				gtx.Constraints.Min.X = gtx.Constraints.Max.X
				mobileWidth := gtx.Dp(350)
				flex := layout.Flex{Spacing: layout.SpaceBetween}
				spacerLayout := layout.Spacer{Width: unit.Dp(16)}
				submitLayout := layout.Flexed(1, func(gtx Gtx) Dim {
					return p.buttonSubmit.Layout(gtx)
				})
				if gtx.Constraints.Max.X <= mobileWidth {
					flex.Axis = layout.Vertical
					spacerLayout.Width = 0
					spacerLayout.Height = 8
					submitLayout = layout.Rigid(func(gtx Gtx) Dim {
						return p.buttonSubmit.Layout(gtx)
					})
				}
				inset := layout.Inset{Top: unit.Dp(16)}
				return inset.Layout(gtx, func(gtx Gtx) Dim {
					return flex.Layout(gtx,
						submitLayout,
						layout.Rigid(spacerLayout.Layout),
					)
				})
			}),
		)
	})
}

func (p *passwordForm) OnDatabaseChange(event service.Event) {
	switch e := event.Data.(type) {
	case service.AccountsChangedEventData:
		_ = e
		p.fetchAccountsCount()
	}
}
