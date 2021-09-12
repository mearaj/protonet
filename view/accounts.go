package view

import (
	"gioui.org/layout"
	"gioui.org/unit"
	"gioui.org/widget"
	"gioui.org/widget/material"
	"gioui.org/x/component"
	log "github.com/sirupsen/logrus"
	"golang.org/x/exp/shiny/materialdesign/icons"
	"image"
	"image/color"
	"protonet.live/database"
)

type AccountsView struct {
	nav               *Navigator
	list              layout.List
	listItems         []*AccountListItem
	th                material.Theme
	barActions        []component.AppBarAction
	overflowActions   []component.OverflowAction
	radioButtonsGroup widget.Enum
	plusBtn,
	overflowState *widget.Clickable
	Addrs    string
	Redraw   bool
	eventKey uint8
	plusIcon,
	menuIcon *widget.Icon
	user *database.Account
}

func NewAccountsView(nav *Navigator) (acsv *AccountsView) {
	acsv = &AccountsView{}
	acsv.nav = nav
	acsv.th = nav.Theme
	acsv.user = Nav.ChatService.GetCurrentUser()
	acsv.listItems = []*AccountListItem{}
	acsv.list.Axis = layout.Vertical
	acsv.plusBtn = &widget.Clickable{}
	acsv.plusIcon, _ = widget.NewIcon(icons.ContentAdd)
	acsv.menuIcon, _ = widget.NewIcon(icons.NavigationMenu)
	acsv.overflowActions = make([]component.OverflowAction, 0, 0)
	acsv.barActions = make([]component.AppBarAction, 0, 0)
	acsv.SetBarActions()
	acsv.AccountListItems()
	if nav.ChatService.IsServiceReady() {
		acsv.radioButtonsGroup.Value = Nav.ChatService.GetCurrentUser().ID
	}
	return acsv
}

func (acsv *AccountsView) Layout(gtx C) (d D) {
	if acsv.radioButtonsGroup.Value == "" {
		acsv.radioButtonsGroup.Value = Nav.ChatService.GetCurrentUser().ID
	}
	if len(acsv.listItems) != len(Nav.ChatService.GetAccounts()) {
		acsv.AccountListItems()
		//InvalidateWindows()
	}

	// FixMe Disable for performance reason and wrongly interpreted by users
	//if acsv.plusBtn.Clicked() && Nav.ChatService.IsServiceReady() {
	//	 Nav.ChatService.CreateNewAccount()
	//}

	d = layout.Flex{Axis: layout.Vertical, Alignment: layout.Start}.Layout(gtx,
		layout.Rigid(acsv.setBar),
		layout.Flexed(1, acsv.drawAccountList),
	)

	if acsv.radioButtonsGroup.Changed() {
		acsv.user = Nav.ChatService.GetAccounts()[acsv.radioButtonsGroup.Value]
		if Nav.ChatService.GetCurrentUser().ID != acsv.user.ID {
			Nav.ChatService.SetCurrentUser(acsv.user)
			acsv.user = Nav.ChatService.GetCurrentUser()
		}
	}
	acsv.initiateAppBarListener(gtx)
	return d
}

func (acsv *AccountsView) setBar(gtx C) D {
	acsv.nav.AppBar.Title = "Accounts"
	acsv.nav.AppBar.NavigationIcon = acsv.menuIcon
	d := acsv.nav.AppBar.Layout(gtx, &acsv.th)
	return d
}

func (acsv *AccountsView) AccountListItems() {
	acsv.listItems = make([]*AccountListItem, 0, len(Nav.ChatService.GetAccounts()))
	var index int
	for _, Account := range Nav.ChatService.GetAccountsArray() {
		listItem := NewAccountListItem(acsv.nav, index, &acsv.radioButtonsGroup, Account)
		index++
		acsv.listItems = append(acsv.listItems, listItem)
	}
}

func (acsv *AccountsView) SetBarActions() {
	acsv.barActions = []component.AppBarAction{
		component.SimpleIconAction(acsv.plusBtn, acsv.plusIcon,
			component.OverflowAction{
				Name: "Create New Account",
				Tag:  &acsv.plusBtn,
			},
		),
	}

	acsv.nav.AppBar.SetActions(acsv.barActions, acsv.overflowActions)
}

func (acsv *AccountsView) drawAccountList(gtx C) D {
	gtx.Constraints.Min.Y = 0
	acsv.list.Axis = layout.Vertical
	return acsv.list.Layout(gtx, len(acsv.listItems), acsv.drawAccountListItem)
}

func (acsv *AccountsView) drawAccountListItem(gtx C, index int) D {
	gtx.Constraints.Max.Y = gtx.Px(unit.Dp(100))
	gtx.Constraints.Min = gtx.Constraints.Max
	liItem := acsv.listItems[index]
	dimensions := layout.Flex{Axis: layout.Vertical}.Layout(gtx,
		layout.Flexed(1, func(gtx C) D {
			return liItem.Layout(gtx)
		}),
		layout.Rigid(func(gtx C) D {
			return component.Rect{Color: color.NRGBA{A: 255},
				Size:  image.Point{X: gtx.Constraints.Max.X, Y: gtx.Px(unit.Dp(1.0))},
				Radii: 0,
			}.Layout(gtx)
		}),
	)
	if liItem.Clicked() && !liItem.Changed() && Nav.ChatService.IsServiceReady() {
		if !liItem.radioClickable.Clicked() && !liItem.Changed() {
			acsv.nav.PushView(NewAccountView(acsv.nav, liItem.Account))
		}
	}
	return dimensions
}

func (acsv *AccountsView) initiateAppBarListener(gtx C) {
	for _, event := range acsv.nav.AppBar.Events(gtx) {
		switch event := event.(type) {
		case component.AppBarNavigationClicked:
			log.Printf("%#v\n", acsv.nav.ModalNavDrawer)
			if Nav.ChatService.IsServiceReady() {
				acsv.nav.ModalNavDrawer.Appear(gtx.Now)
			}
		case component.AppBarContextMenuDismissed:
			log.Printf("Context menu dismissed: %v", event)
		case component.AppBarOverflowActionClicked:
			log.Printf("Overflow action selected: %v", event)
		}
	}
}
