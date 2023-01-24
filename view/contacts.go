package view

import (
	"gioui.org/layout"
	"gioui.org/widget"
	"gioui.org/widget/material"
	"gioui.org/x/component"
	"github.com/mearaj/protonet/jni"
	log "github.com/sirupsen/logrus"
	"golang.org/x/exp/shiny/materialdesign/icons"
	"image"
	"image/color"
	"runtime"
)

type ContactsView struct {
	nav                 *Navigator
	list                layout.List
	listItems           []*ListItem
	selectedRoomIndex   int
	selectedRoomChanged bool
	th                  material.Theme
	barActions          []component.AppBarAction
	overflowActions     []component.OverflowAction
	plusBtn,
	overflowState *widget.Clickable
	Addrs    string
	Redraw   bool
	eventKey uint8
	plusIcon,
	menuIcon *widget.Icon

	// Tags to identify events
	keyCopyId             uint8
	keySelectAll          uint8
	keyDeleteSelected     uint8
	keyClearSelection     uint8
	keyAtLeastOneSelected uint8
	keyShareMyID          uint8
}

func NewContactsView(nav *Navigator) (crs *ContactsView) {
	crs = &ContactsView{}
	crs.nav = nav
	crs.th = nav.Theme
	crs.listItems = []*ListItem{}
	crs.list.Axis = layout.Vertical
	crs.plusBtn = &widget.Clickable{}
	crs.plusIcon, _ = widget.NewIcon(icons.ContentAdd)
	crs.menuIcon, _ = widget.NewIcon(icons.NavigationMenu)
	crs.SetBarActions()
	crs.createChatRoomListItems()
	return crs
}

func (ctsV *ContactsView) Layout(gtx C) (d D) {
	//if Nav.ChatService.IsServiceReady() {
	//	ctsV.createChatRoomListItems()
	//	InvalidateWindows()
	//}
	d = layout.Flex{Axis: layout.Vertical, Alignment: layout.Start}.Layout(gtx,
		layout.Rigid(ctsV.setBar),
		layout.Flexed(1, ctsV.drawChatRoomList))
	if ctsV.plusBtn.Clicked() {
		ctsV.nav.PushView(NewContactView(ctsV.nav, ""))
	}

	if ctsV.isAnyItemChanged() {
		ctsV.SetBarActions()
		InvalidateWindows()
	}
	ctsV.initiateAppBarListener(gtx)
	if Nav.ChatService.IsServiceReady() && len(ctsV.listItems) != len(Nav.ChatService.GetTxtChatServicesMap()) {
		ctsV.createChatRoomListItems()
	}
	return d
}

func (ctsV *ContactsView) setBar(gtx C) D {
	ctsV.nav.AppBar.Title = "Contacts"
	ctsV.nav.AppBar.NavigationIcon = ctsV.menuIcon
	d := ctsV.nav.AppBar.Layout(gtx, &ctsV.th, "App Bar", "App Bar Overflow")
	return d
}

func (ctsV *ContactsView) createChatRoomListItems() {
	ctsV.listItems = make([]*ListItem, 0, len(Nav.ChatService.GetTxtChatServicesMap()))
	var index int
	for _, txtChatService := range Nav.ChatService.GetTxtChatServicesMap() {
		listItem := NewContactListItem(ctsV.nav, index, txtChatService)
		index++
		ctsV.listItems = append(ctsV.listItems, listItem)
	}
}

func (ctsV *ContactsView) SetBarActions() {
	ctsV.barActions = []component.AppBarAction{
		component.SimpleIconAction(ctsV.plusBtn, ctsV.plusIcon,
			component.OverflowAction{
				Name: "Create Chat Room",
				Tag:  &ctsV.plusBtn,
			},
		),
	}

	copyAction := component.OverflowAction{Name: "Copy / Share IDHex", Tag: &ctsV.keyCopyId}
	selectAllAction := component.OverflowAction{Name: "Select All", Tag: &ctsV.keySelectAll}
	deleteAction := component.OverflowAction{Name: "Delete Selected", Tag: &ctsV.keyDeleteSelected}
	clearAction := component.OverflowAction{Name: "Clear", Tag: &ctsV.keyClearSelection}

	ctsV.overflowActions = []component.OverflowAction{
		copyAction,
		selectAllAction,
	}
	if ctsV.isAllSelected() {
		ctsV.overflowActions = []component.OverflowAction{
			copyAction,
			deleteAction,
			clearAction,
		}
	} else if ctsV.atLeastOneSelected() {
		ctsV.overflowActions = []component.OverflowAction{
			copyAction,
			selectAllAction,
			deleteAction,
			clearAction,
		}
	}

	ctsV.nav.AppBar.SetActions(ctsV.barActions, ctsV.overflowActions)
}

func (ctsV *ContactsView) drawChatRoomList(gtx C) D {
	gtx.Constraints.Min.Y = 0
	ctsV.list.Axis = layout.Vertical
	return ctsV.list.Layout(gtx, len(ctsV.listItems), ctsV.drawChatRoomListItem)
}

func (ctsV *ContactsView) drawChatRoomListItem(gtx C, index int) D {
	gtx.Constraints.Max.Y = gtx.Dp(100)
	gtx.Constraints.Min = gtx.Constraints.Max
	liItem := ctsV.listItems[index]
	dimensions := layout.Flex{Axis: layout.Vertical}.Layout(gtx,
		layout.Flexed(1, func(gtx C) D {
			return liItem.Layout(gtx)
		}),
		layout.Rigid(func(gtx C) D {
			return component.Rect{Color: color.NRGBA{A: 255},
				Size:  image.Point{X: gtx.Constraints.Max.X, Y: gtx.Dp(1.0)},
				Radii: 0,
			}.Layout(gtx)
		}),
	)
	if liItem.Clicked() && !liItem.Changed() {
		if !liItem.checkboxClickable.Clicked() && !liItem.Changed() {
			cs := ctsV.listItems[index].cs
			Nav.PushView(NewChatRoom(Nav, cs))
		}
	}
	return dimensions
}

func (ctsV *ContactsView) initiateAppBarListener(gtx C) {
	for _, event := range ctsV.nav.AppBar.Events(gtx) {
		switch event := event.(type) {
		case component.AppBarNavigationClicked:
			log.Printf("%#v\n", ctsV.nav.ModalNavDrawer)
			ctsV.nav.ModalNavDrawer.Appear(gtx.Now)
		case component.AppBarContextMenuDismissed:
			log.Printf("Context menu dismissed: %v", event)
		case component.AppBarOverflowActionClicked:
			log.Printf("Overflow action selected: %v", event)
			if event.Tag == &ctsV.keyCopyId {
				Window.WriteClipboard(Nav.ChatService.GetCurrentUser().ID)
				switch runtime.GOOS {
				case "android":
					jni.ShareStringWith(Nav.ChatService.GetCurrentUser().ID)
				}
			} else if event.Tag == &ctsV.keySelectAll {
				ctsV.selectAllItems()
				ctsV.SetBarActions()
			} else if event.Tag == &ctsV.keyDeleteSelected {
				ctsV.deleteSelectedItems()
				ctsV.SetBarActions()
			} else if event.Tag == &ctsV.keyClearSelection {
				ctsV.clearSelectedItems()
				ctsV.SetBarActions()
			}
		}
	}
}

func (ctsV *ContactsView) selectAllItems() {
	if len(ctsV.listItems) > 0 {
		for _, item := range ctsV.listItems {
			item.checkbox.Value = true
		}
	}
}

func (ctsV *ContactsView) isAllSelected() bool {
	if len(ctsV.listItems) > 0 {
		for _, item := range ctsV.listItems {
			if !item.checkbox.Value {
				return false
			}
		}
	} else {
		return false
	}
	return true
}

func (ctsV *ContactsView) atLeastOneSelected() bool {
	if len(ctsV.listItems) > 0 {
		for _, item := range ctsV.listItems {
			if item.checkbox.Value {
				log.Println("returning true")
				return true
			}
		}
	}
	return false
}

func (ctsV *ContactsView) deleteSelectedItems() {
	if len(ctsV.listItems) > 0 {
		clientIDs := make([]string, 0, 1)
		for _, item := range ctsV.listItems {
			if item.checkbox.Value {
				clientIDs = append(clientIDs, item.cs.GetClient().IDStr)
			}
		}
		if len(clientIDs) > 0 {
			Nav.ChatService.DeleteTxtChatServicesMapItems(clientIDs)
		}
		//InvalidateWindows()
	}
}
func (ctsV *ContactsView) clearSelectedItems() {
	if len(ctsV.listItems) > 0 {
		for _, item := range ctsV.listItems {
			item.checkbox.Value = false
		}
	}
}

func (ctsV *ContactsView) isAnyItemChanged() bool {
	if len(ctsV.listItems) > 0 {
		for _, item := range ctsV.listItems {
			if item.checkbox.Changed() {
				return true
			}
		}
	}
	return false
}
