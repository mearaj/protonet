package view

import (
	"gioui.org/layout"
	"gioui.org/unit"
	"gioui.org/widget"
	"gioui.org/widget/material"
	"gioui.org/x/component"
	log "github.com/sirupsen/logrus"
	"golang.org/x/exp/shiny/materialdesign/icons"
	"protonet.live/database"
	"protonet.live/jni"
	"runtime"
)

type AccountDetailView struct {
	nav             *Navigator
	list            layout.List
	th              material.Theme
	barActions      []component.AppBarAction
	overflowActions []component.OverflowAction
	overflowState   widget.Clickable
	backIcon        *widget.Icon
	submitIcon      *widget.Icon
	Account         *database.Account
	nameTextField   component.TextField
	copyBtn,
	saveBtn,
	copyPvtKey,
	copyPubKey,
	cancelBtn widget.Clickable
}

func NewAccountView(nav *Navigator, acc *database.Account) (accv *AccountDetailView) {
	accv = &AccountDetailView{}
	accv.nav = nav
	accv.th = nav.Theme
	accv.backIcon, _ = widget.NewIcon(icons.NavigationArrowBack)
	accv.list.Axis = layout.Vertical
	accv.list.Alignment = layout.End
	accv.submitIcon, _ = widget.NewIcon(icons.ContentSend)
	accv.nameTextField = component.TextField{}
	accv.Account = acc
	accv.overflowActions = []component.OverflowAction{
		{
			Name: "Example 1",
			Tag:  &accv.overflowState,
		},
		{
			Name: "Example 2",
			Tag:  &accv.overflowActions,
		},
	}
	accv.SetBarActions()
	accv.nameTextField.SetText(accv.Account.Name)
	return accv
}

func (sv *AccountDetailView) Layout(gtx C) (d D) {
	sv.handleAppBarEvents(gtx)

	fl := layout.Flex{
		Axis:      layout.Vertical,
		Spacing:   layout.SpaceBetween,
		Alignment: layout.Start,
		WeightSum: 1.0,
	}
	d = fl.Layout(gtx,
		layout.Rigid(sv.setBar),
		layout.Rigid(sv.drawAccountViewList),
	)
	if sv.saveBtn.Clicked() {
		privateName := sv.nameTextField.Text()
		sv.Account.Name = privateName
		Nav.ChatService.SaveAccountToDisk(sv.Account)
		sv.nav.NavDrawer.Subtitle = privateName
		//InvalidateWindows()
		//switch runtime.GOOS {
		//case "android":
		//	jni.OpenImage()
		//}
	} else if sv.cancelBtn.Clicked() {
		sv.nameTextField.SetText(sv.Account.Name)
	}
	if sv.copyBtn.Clicked() {
		Window.WriteClipboard(sv.Account.ID)
		switch runtime.GOOS {
		case "android":
			jni.ShareStringWith(sv.Account.ID)
		}
	}
	if sv.copyPvtKey.Clicked() {
		Window.WriteClipboard(sv.Account.PvtKeyHex)
	}
	if sv.copyPubKey.Clicked() {
		Window.WriteClipboard(sv.Account.PubKeyHex)
	}

	return
}

func (sv *AccountDetailView) setBar(gtx C) D {
	sv.nav.AppBar.Title = sv.Account.Name
	sv.nav.AppBar.NavigationIcon = sv.backIcon
	d := sv.nav.AppBar.Layout(gtx, &sv.th)
	return d
}

func (sv *AccountDetailView) SetBarActions() {
	sv.nav.AppBar.SetActions(sv.barActions, sv.overflowActions)
}

func (sv *AccountDetailView) handleAppBarEvents(gtx C) {
	for _, event := range sv.nav.AppBar.Events(gtx) {
		switch event := event.(type) {
		case component.AppBarNavigationClicked:
			log.Println("App Menu clicked from accounts view")
			sv.nav.PopView()
		case component.AppBarContextMenuDismissed:
			log.Printf("Context menu dismissed: %v", event)
		case component.AppBarOverflowActionClicked:
			log.Printf("Overflow action selected: %v", event)
		default:
		}
	}
}

func (sv *AccountDetailView) drawPeerIDField(gtx C) D {
	return layout.UniformInset(unit.Dp(16.0)).Layout(gtx, func(gtx C) D {
		return layout.Flex{Axis: layout.Vertical}.Layout(gtx,
			layout.Rigid(func(gtx C) D {
				return layout.Inset{
					Top:    unit.Dp(8.0),
					Right:  unit.Dp(0.0),
					Bottom: unit.Dp(8.0),
					Left:   unit.Dp(0.0),
				}.Layout(gtx, func(gtx C) D {
					return layout.Flex{Alignment: layout.Middle}.Layout(gtx,
						layout.Rigid(func(gtx C) D {
							return material.Label(&sv.th, unit.Dp(24), "Peer/Public ID").Layout(gtx)
						}),
						layout.Rigid(func(gtx C) D {
							return layout.Inset{Left: unit.Dp(32.0)}.Layout(gtx,
								func(gtx C) D {
									return material.Button(&sv.th,
										&sv.copyBtn, "Copy / Share").Layout(gtx)
								},
							)
						}),
					)
				})
			}),
			layout.Rigid(func(gtx C) D {
				return material.Label(&sv.th, unit.Dp(16), sv.Account.ID).Layout(gtx)
			}),
			layout.Rigid(func(gtx C) D {
				return layout.Inset{
					Top:    unit.Dp(8.0),
					Right:  unit.Dp(0.0),
					Bottom: unit.Dp(8.0),
					Left:   unit.Dp(0.0),
				}.Layout(gtx, func(gtx C) D {
					return layout.Flex{Alignment: layout.Middle}.Layout(gtx,
						layout.Rigid(func(gtx C) D {
							return material.Label(&sv.th, unit.Dp(24), "Private Key").Layout(gtx)
						}),
						layout.Rigid(func(gtx C) D {
							return layout.Inset{Left: unit.Dp(32.0)}.Layout(gtx,
								func(gtx C) D {
									return material.Button(&sv.th,
										&sv.copyPvtKey, "Copy").Layout(gtx)
								},
							)
						}),
					)
				})
			}),
			layout.Rigid(func(gtx C) D {
				return material.Label(&sv.th, unit.Dp(16), sv.Account.PvtKeyHex).Layout(gtx)
			}),
			layout.Rigid(func(gtx C) D {
				return layout.Inset{
					Top:    unit.Dp(8.0),
					Right:  unit.Dp(0.0),
					Bottom: unit.Dp(8.0),
					Left:   unit.Dp(0.0),
				}.Layout(gtx, func(gtx C) D {
					return layout.Flex{Alignment: layout.Middle}.Layout(gtx,
						layout.Rigid(func(gtx C) D {
							return material.Label(&sv.th, unit.Dp(24), "Public Key").Layout(gtx)
						}),
						layout.Rigid(func(gtx C) D {
							return layout.Inset{Left: unit.Dp(32.0)}.Layout(gtx,
								func(gtx C) D {
									return material.Button(&sv.th,
										&sv.copyPubKey, "Copy").Layout(gtx)
								},
							)
						}),
					)
				})
			}),
			layout.Rigid(func(gtx C) D {
				return material.Label(&sv.th, unit.Dp(16), sv.Account.PubKeyHex).Layout(gtx)
			}),
		)

	})
}

func (sv *AccountDetailView) drawAccountViewList(gtx C) D {
	return layout.Flex{}.Layout(gtx,
		layout.Flexed(1.0, func(gtx C) D {
			return layout.Inset{
				Top:    unit.Dp(8),
				Right:  unit.Dp(8),
				Bottom: unit.Dp(8),
				Left:   unit.Dp(8),
			}.Layout(gtx, func(gtx C) D {
				return sv.list.Layout(gtx, 3, sv.drawAccountViewListItem)
			})
		}))
}
func (sv *AccountDetailView) drawAccountViewListItem(gtx C, index int) D {
	var labelText string
	var labelHintText string
	var textField *component.TextField
	switch index {
	case 0:
		return sv.drawPeerIDField(gtx)
	case 1:
		textField = &sv.nameTextField
		labelText = "Your Name"
		labelHintText = "Enter your name"
	case 2:
		return sv.drawSubmitButtons(gtx)
	}

	d := layout.UniformInset(unit.Dp(16.0)).Layout(gtx,
		func(gtx C) D {
			return layout.Flex{
				Axis:      layout.Vertical,
				Spacing:   layout.SpaceStart,
				Alignment: layout.Middle}.Layout(gtx,
				layout.Rigid(func(gtx C) D {
					return layout.Flex{
						Axis:      layout.Horizontal,
						Spacing:   layout.SpaceBetween,
						Alignment: layout.Middle}.Layout(gtx,
						layout.Flexed(1.0, func(gtx C) D {
							return layout.Inset{
								Top:    unit.Dp(0),
								Right:  unit.Dp(0),
								Bottom: unit.Dp(8.0),
								Left:   unit.Dp(0),
							}.Layout(gtx, func(gtx C) D {
								return material.Label(&sv.th, unit.Dp(16.0), labelText).Layout(gtx)
							})
						}),
					)
				}),
				layout.Rigid(func(gtx C) D {
					return layout.Flex{
						Axis:      layout.Horizontal,
						Spacing:   layout.SpaceBetween,
						Alignment: layout.Middle}.Layout(gtx,
						layout.Flexed(1.0, func(gtx C) D {
							return textField.Layout(gtx,
								&sv.th, labelHintText)
						}),
					)
				}),
			)
		})
	return d
}

func (sv *AccountDetailView) drawSubmitButtons(gtx C) D {
	gtx.Constraints.Min.X = gtx.Constraints.Max.X
	return layout.Inset{
		Top:    unit.Dp(16.0),
		Right:  unit.Dp(16.0),
		Bottom: unit.Dp(16.0),
		Left:   unit.Dp(16.0),
	}.Layout(gtx,
		func(gtx C) D {
			return layout.Flex{
				Axis:      layout.Horizontal,
				Spacing:   layout.SpaceSides,
				Alignment: layout.Middle}.Layout(gtx,
				layout.Rigid(func(gtx C) D {
					return layout.Inset{
						Top:    unit.Dp(0.0),
						Right:  unit.Dp(8.0),
						Bottom: unit.Dp(0.0),
						Left:   unit.Dp(0.0),
					}.Layout(gtx,
						func(gtx C) D {
							return material.Button(&sv.th,
								&sv.saveBtn, "Save").Layout(gtx)
						})
				}),
				layout.Rigid(func(gtx C) D {
					return layout.Inset{
						Top:    unit.Dp(0.0),
						Right:  unit.Dp(0.0),
						Bottom: unit.Dp(0.0),
						Left:   unit.Dp(8.0),
					}.Layout(gtx,
						func(gtx C) D {
							return material.Button(&sv.th,
								&sv.cancelBtn, "Cancel").Layout(gtx)
						})
				}),
			)
		},
	)
}
