package view

import (
	"gioui.org/io/clipboard"
	"gioui.org/layout"
	"gioui.org/unit"
	"gioui.org/widget"
	"gioui.org/widget/material"
	"gioui.org/x/component"
	"github.com/libp2p/go-libp2p/core/peer"
	log "github.com/sirupsen/logrus"
	"golang.org/x/exp/shiny/materialdesign/icons"
	"image"
	"strings"
)

type ContactView struct {
	nav             *Navigator
	th              material.Theme
	barActions      []component.AppBarAction
	overflowActions []component.OverflowAction
	plusBtn,
	overflowState,
	saveBtn,
	pasteBtnID,
	pasteBtnName,
	lastPasteBtnClicked,
	cancelBtn *widget.Clickable
	inputIDField    component.TextField
	inputNameField  component.TextField
	Addrs           string
	eventKey        int
	backIcon        *widget.Icon
	clientID        string
	isAddingContact bool
}

func (crs *ContactView) Layout(gtx C) (d D) {
	crs.initiateAppBarListener(gtx)

	if crs.cancelBtn.Clicked() {
		crs.inputIDField.SetText("")
		crs.inputNameField.SetText("")
		crs.isAddingContact = false
		crs.nav.PopView()
	}

	if crs.saveBtn.Clicked() {
		if crs.isAddingContact {
			crs.isAddingContact = false
			//InvalidateWindows()
		} else {
			clientID := strings.TrimSpace(crs.inputIDField.Text())
			clientName := strings.TrimSpace(crs.inputNameField.Text())
			if clientID != "" {

				if _, err := peer.Decode(clientID); err != nil {
					log.Println("error in adding cs, error is err", err)
				} else {
					crs.isAddingContact = true
					go func() {
						err = Nav.ChatService.AddContact(clientID, clientName)
						if err != nil {
							log.Println(err)
						}
						crs.clientID = clientID
						crs.isAddingContact = false
						//InvalidateWindows()
						crs.nav.PopView()
					}()
				}
			}
		}
	}

	fl := layout.Flex{
		Axis:      layout.Vertical,
		Spacing:   0,
		Alignment: layout.Start,
		WeightSum: 1,
	}

	d = fl.Layout(gtx,
		layout.Rigid(crs.setBar),
		layout.Rigid(crs.drawTextField),
	)

	if crs.pasteBtnID.Clicked() {
		crs.lastPasteBtnClicked = crs.pasteBtnID
		clipboard.ReadOp{Tag: &crs.eventKey}.Add(gtx.Ops)
	} else if crs.pasteBtnName.Clicked() {
		crs.lastPasteBtnClicked = crs.pasteBtnName
		clipboard.ReadOp{Tag: &crs.eventKey}.Add(gtx.Ops)
	}

	if crs.lastPasteBtnClicked != nil {
		for _, e := range gtx.Events(&crs.eventKey) {
			switch e := e.(type) {
			case clipboard.Event:
				if crs.lastPasteBtnClicked == crs.pasteBtnID {
					crs.inputIDField.SetText(e.Text)
					crs.lastPasteBtnClicked = nil
					break
				}
				if crs.lastPasteBtnClicked == crs.pasteBtnName {
					crs.inputNameField.SetText(e.Text)
					crs.lastPasteBtnClicked = nil
					break
				}
			}
		}
	}

	return d
}

func NewContactView(nav *Navigator, clientID string) (crs *ContactView) {
	crs = &ContactView{}
	crs.nav = nav
	crs.th = nav.Theme
	crs.backIcon, _ = widget.NewIcon(icons.NavigationArrowBack)
	crs.inputIDField = component.TextField{}
	crs.inputNameField = component.TextField{}
	crs.plusBtn = &widget.Clickable{}
	crs.cancelBtn = &widget.Clickable{}
	crs.saveBtn = &widget.Clickable{}
	crs.pasteBtnID = &widget.Clickable{}
	crs.pasteBtnName = &widget.Clickable{}
	crs.clientID = clientID
	plusIcon, _ := widget.NewIcon(icons.ContentAdd)
	crs.barActions = []component.AppBarAction{
		component.SimpleIconAction(crs.plusBtn, plusIcon,
			component.OverflowAction{
				Name: "Create Chat Room",
				Tag:  &crs.plusBtn,
			},
		),
	}
	crs.overflowActions = []component.OverflowAction{
		{
			Name: "Copy My IDHex",
			Tag:  &crs.overflowState,
		},
		{
			Name: "Example 2",
			Tag:  &crs.overflowActions,
		},
	}
	crs.restoreInputs()
	crs.SetBarActions()
	return crs
}

func (crs *ContactView) setBar(gtx C) D {
	crs.nav.AppBar.Title = "Add/Edit contact"
	crs.nav.AppBar.NavigationIcon = crs.backIcon
	d := crs.nav.AppBar.Layout(gtx, &crs.th, "App Bar", "App Bar Overflow")
	return d
}

func (crs *ContactView) SetBarActions() {
	crs.nav.AppBar.SetActions(crs.barActions, crs.overflowActions)
}

func (crs *ContactView) initiateAppBarListener(gtx C) {
	for _, event := range crs.nav.AppBar.Events(gtx) {
		switch event := event.(type) {
		case component.AppBarNavigationClicked:
			log.Println("Back Icon clicked")
			if !crs.isAddingContact {
				crs.nav.PopView()
			}
		case component.AppBarContextMenuDismissed:
			log.Printf("Context menu dismissed: %v", event)
		case component.AppBarOverflowActionClicked:
			log.Printf("Overflow action selected: %v", event)
			if event.Tag == &crs.overflowState {
				Window.WriteClipboard(Nav.ChatService.GetCurrentUser().ID)
				log.Println("copied to clipboard ->", Nav.ChatService.GetCurrentUser().ID)
			}
		}
	}
}

func (crs *ContactView) drawTextField(gtx C) (d D) {
	loader := layout.Rigid(func(gtx C) (d D) {
		return d
	})
	if crs.isAddingContact {
		loader = layout.Rigid(func(gtx C) (d D) {
			return material.Loader(&crs.th).Layout(gtx)
		})
	}
	d = layout.UniformInset(unit.Dp(16.0)).Layout(gtx,
		func(gtx C) D {
			return layout.Flex{
				Axis:      layout.Vertical,
				Spacing:   layout.SpaceEnd,
				Alignment: layout.Middle}.Layout(gtx,
				layout.Rigid(func(gtx C) D {
					gtx.Constraints.Max = image.Point{
						X: gtx.Constraints.Max.X,
						Y: gtx.Constraints.Max.Y / 2,
					}
					return layout.Flex{
						Axis:      layout.Horizontal,
						Spacing:   layout.SpaceBetween,
						Alignment: layout.Middle}.Layout(gtx,
						layout.Flexed(1.0, func(gtx C) D {
							return crs.inputNameField.Layout(gtx,
								&crs.th, "Name")
						}),
						layout.Rigid(func(gtx C) D {
							return layout.Inset{
								Top:    unit.Dp(8.0),
								Right:  unit.Dp(8.0),
								Bottom: unit.Dp(0.0),
								Left:   unit.Dp(16.0),
							}.Layout(gtx, func(gtx C) D {
								return material.Button(&crs.th,
									crs.pasteBtnName, "Paste").Layout(gtx)
							})
						}),
					)
				}),
				layout.Rigid(func(gtx C) D {
					return layout.Spacer{Height: unit.Dp(16.0)}.Layout(gtx)
				}),
				layout.Rigid(func(gtx C) D {
					gtx.Constraints.Max = image.Point{
						X: gtx.Constraints.Max.X,
						Y: gtx.Constraints.Max.Y / 2,
					}
					return layout.Flex{
						Axis:      layout.Horizontal,
						Spacing:   layout.SpaceBetween,
						Alignment: layout.Middle}.Layout(gtx,
						layout.Flexed(1.0, func(gtx C) D {
							return crs.inputIDField.Layout(gtx,
								&crs.th, "Public Address")
						}),
						layout.Rigid(func(gtx C) D {
							return layout.Inset{
								Top:    unit.Dp(8.0),
								Right:  unit.Dp(8.0),
								Bottom: unit.Dp(0.0),
								Left:   unit.Dp(16.0),
							}.Layout(gtx, func(gtx C) D {
								return material.Button(&crs.th,
									crs.pasteBtnID, "Paste").Layout(gtx)
							})
						}),
					)
				}),
				layout.Rigid(func(gtx C) D {
					return layout.Spacer{Height: unit.Dp(32.0)}.Layout(gtx)
				}),
				layout.Rigid(func(gtx C) D {
					return layout.Inset{
						Top:    unit.Dp(16.0),
						Right:  unit.Dp(0.0),
						Bottom: unit.Dp(0.0),
						Left:   unit.Dp(8.0),
					}.Layout(gtx,
						func(gtx C) D {
							return layout.Flex{
								Axis:      layout.Horizontal,
								Spacing:   layout.SpaceBetween,
								Alignment: layout.Middle}.Layout(gtx,
								layout.Rigid(func(gtx C) D {
									return layout.Inset{
										Top:    unit.Dp(0.0),
										Right:  unit.Dp(8.0),
										Bottom: unit.Dp(0.0),
										Left:   unit.Dp(0.0),
									}.Layout(gtx,
										func(gtx C) D {
											return material.Button(&crs.th,
												crs.saveBtn, "Save").Layout(gtx)
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
											return material.Button(&crs.th,
												crs.cancelBtn, "Cancel").Layout(gtx)
										})
								}),
							)
						},
					)

				}),
				loader,
			)
		})

	return d
}

func (crs ContactView) restoreInputs() {
	if crs.clientID != "" {
		if txtService, ok := Nav.ChatService.GetTxtChatServicesMap()[crs.clientID]; ok {
			crs.inputIDField.SetText(txtService.GetClient().IDStr)
			crs.inputNameField.SetText(txtService.GetClient().Name)
		}
	}
}
