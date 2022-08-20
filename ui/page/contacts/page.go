package contacts

import (
	"fmt"
	"gioui.org/io/pointer"
	"gioui.org/layout"
	"gioui.org/op"
	"gioui.org/op/clip"
	"gioui.org/text"
	"gioui.org/unit"
	"gioui.org/widget"
	"gioui.org/widget/material"
	"gioui.org/x/component"
	"github.com/mearaj/protonet/service"
	. "github.com/mearaj/protonet/ui/fwk"
	"github.com/mearaj/protonet/ui/view"
	"golang.org/x/exp/shiny/materialdesign/colornames"
	"golang.org/x/exp/shiny/materialdesign/icons"
	"image/color"
	"time"
)

var defaultListSize = 50

type page struct {
	layout.List
	Manager
	Theme              *material.Theme
	title              string
	iconNewChat        *widget.Icon
	btnAddContact      widget.Clickable
	btnYes             widget.Clickable
	btnNo              widget.Clickable
	buttonNavigation   widget.Clickable
	btnBackdrop        widget.Clickable
	btnMenuIcon        widget.Clickable
	btnCloseSelection  widget.Clickable
	btnDeleteSelection widget.Clickable
	btnMenuContent     widget.Clickable
	btnSelectAll       widget.Clickable
	btnDeleteAll       widget.Clickable
	btnSelectionMode   widget.Clickable
	menuIcon           *widget.Icon
	closeIcon          *widget.Icon
	menuVisibilityAnim component.VisibilityAnimation
	navigationIcon     *widget.Icon
	contactsView       []*pageItem
	NoContact          View
	NoAccount          View
	ContactForm        View
	*view.ModalContent
	SelectionMode           bool
	fetchingContactsCh      chan []service.Contact
	isFetchingContacts      bool
	isFetchingContactsCount bool
	listPosition            layout.Position
	fetchingContactsCountCh chan int64
	contactsCount           int64
	initialized             bool
}

func New(manager Manager) Page {
	navIcon, _ := widget.NewIcon(icons.NavigationArrowBack)
	closeIcon, _ := widget.NewIcon(icons.ContentClear)
	iconNewChat, _ := widget.NewIcon(icons.ContentCreate)
	iconMenu, _ := widget.NewIcon(icons.NavigationMoreVert)
	errorTh := *manager.Theme()
	errorTh.ContrastBg = color.NRGBA(colornames.Red500)
	theme := *manager.Theme()
	p := page{
		Manager:                 manager,
		Theme:                   &theme,
		title:                   "Contacts",
		navigationIcon:          navIcon,
		closeIcon:               closeIcon,
		iconNewChat:             iconNewChat,
		List:                    layout.List{Axis: layout.Vertical},
		contactsView:            []*pageItem{},
		menuIcon:                iconMenu,
		fetchingContactsCh:      make(chan []service.Contact, 10),
		fetchingContactsCountCh: make(chan int64, 10),
		menuVisibilityAnim: component.VisibilityAnimation{
			Duration: time.Millisecond * 250,
			State:    component.Invisible,
			Started:  time.Time{},
		},
	}
	p.ContactForm = view.NewContactForm(manager, service.Contact{}, p.onAddContactSuccess)
	p.ModalContent = view.NewModalContent(func() { p.Modal().Dismiss(nil) })
	p.NoAccount = view.NewNoAccount(manager)
	p.NoContact = view.NewNoContact(manager, p.onAddContactSuccess, "Add Contact")
	return &p
}

func (p *page) Layout(gtx Gtx) Dim {
	if !p.initialized {
		p.fetchContacts(0, defaultListSize)
		p.fetchContactsCount()
		p.initialized = true
	}
	p.fetchContactsOnScroll(gtx)
	p.listenToFetchContacts()
	p.listenToFetchContactsCount()

	for _, item := range p.contactsView {
		if p.SelectionMode {
			item.SelectionMode = p.SelectionMode
		} else {
			if item.SelectionMode {
				p.SelectionMode = item.SelectionMode
				break
			}
		}
	}
	if p.SelectionMode {
		p.Theme.ContrastBg = color.NRGBA{A: 255}
	} else {
		p.Theme.ContrastBg = p.Manager.Theme().ContrastBg
	}

	if p.btnAddContact.Clicked() {
		p.Modal().Show(p.drawAddContactModal, nil, Animation{
			Duration: time.Millisecond * 250,
			State:    component.Invisible,
			Started:  time.Time{},
		})
		p.menuVisibilityAnim.Disappear(gtx.Now)
	}
	flex := layout.Flex{Axis: layout.Vertical,
		Spacing:   layout.SpaceEnd,
		Alignment: layout.Start,
	}
	d := flex.Layout(gtx,
		layout.Rigid(p.DrawAppBar),
		layout.Rigid(p.drawContactsItems),
	)
	p.drawMenuLayout(gtx)
	for _, e := range gtx.Queue.Events(p) {
		switch e := e.(type) {
		case pointer.Event:
			switch e.Type {
			case pointer.Press:
				if !p.btnMenuContent.Pressed() {
					p.menuVisibilityAnim.Disappear(gtx.Now)
				}
				for _, contactView := range p.contactsView {
					if !contactView.btnMenuContent.Pressed() && !contactView.Hovered() {
						contactView.menuVisibilityAnim.Disappear(gtx.Now)
					}
				}
			}
		}
	}
	return d
}

func (p *page) DrawAppBar(gtx Gtx) Dim {
	gtx.Constraints.Max.Y = gtx.Dp(56)
	if p.btnMenuIcon.Clicked() {
		p.menuVisibilityAnim.Appear(gtx.Now)
	}
	if p.SelectionMode {
		return p.DrawSelectionAppBar(gtx)
	}
	return p.DrawNormalAppBar(gtx)
}
func (p *page) DrawNormalAppBar(gtx Gtx) Dim {
	th := p.Theme
	if p.buttonNavigation.Clicked() {
		p.PopUp()
	}
	return view.DrawAppBarLayout(gtx, th, func(gtx Gtx) Dim {
		return layout.Flex{Alignment: layout.Middle, Spacing: layout.SpaceBetween}.Layout(gtx,
			layout.Rigid(func(gtx Gtx) Dim {
				return layout.Flex{Alignment: layout.Middle}.Layout(gtx,
					layout.Rigid(func(gtx Gtx) Dim {
						navigationIcon := p.navigationIcon
						button := material.IconButton(th, &p.buttonNavigation, navigationIcon, "Nav Icon Button")
						button.Size = unit.Dp(40)
						button.Background = th.Palette.ContrastBg
						button.Color = th.Palette.ContrastFg
						button.Inset = layout.UniformInset(unit.Dp(8))
						return button.Layout(gtx)
					}),
					layout.Rigid(func(gtx Gtx) Dim {
						return layout.Inset{Left: unit.Dp(16)}.Layout(gtx, func(gtx Gtx) Dim {
							titleText := p.title
							title := material.Body1(th, titleText)
							title.Color = th.Palette.ContrastFg
							title.TextSize = unit.Sp(18)
							return title.Layout(gtx)
						})
					}),
				)
			}),
			layout.Rigid(func(gtx Gtx) Dim {
				button := material.IconButton(th, &p.btnMenuIcon, p.menuIcon, "Context Menu")
				button.Size = unit.Dp(40)
				button.Background = th.Palette.ContrastBg
				button.Color = th.Palette.ContrastFg
				button.Inset = layout.UniformInset(unit.Dp(8))
				d := button.Layout(gtx)
				return d
			}),
		)
	})
}
func (p *page) DrawSelectionAppBar(gtx Gtx) Dim {
	th := p.Theme
	if p.btnCloseSelection.Clicked() {
		p.clearAllSelection()
		p.menuVisibilityAnim.Disappear(gtx.Now)
	}

	return view.DrawAppBarLayout(gtx, th, func(gtx Gtx) Dim {
		return layout.Flex{Alignment: layout.Middle, Spacing: layout.SpaceBetween}.Layout(gtx,
			layout.Rigid(func(gtx Gtx) Dim {
				return layout.Flex{Alignment: layout.Middle}.Layout(gtx,
					layout.Rigid(func(gtx Gtx) Dim {
						closeIcon := p.closeIcon
						button := material.IconButton(th, &p.btnCloseSelection, closeIcon, "Close Selection Icon Button")
						button.Size = unit.Dp(40)
						button.Background = th.Palette.ContrastBg
						button.Color = th.Palette.ContrastFg
						button.Inset = layout.UniformInset(unit.Dp(8))
						return button.Layout(gtx)
					}),
					layout.Rigid(func(gtx Gtx) Dim {
						return layout.Inset{Left: unit.Dp(16)}.Layout(gtx, func(gtx Gtx) Dim {
							var txt string
							count := p.getSelectionCount()
							if count == 0 {
								txt = "None Selected"
							} else {
								txt = fmt.Sprintf("(%d) Selected", count)
							}

							title := material.Body1(th, txt)
							title.Color = th.Palette.ContrastFg
							title.TextSize = unit.Sp(18)
							return title.Layout(gtx)
						})
					}),
				)
			}),
			layout.Rigid(func(gtx Gtx) Dim {
				button := material.IconButton(th, &p.btnMenuIcon, p.menuIcon, "Context Menu")
				button.Size = unit.Dp(40)
				button.Background = th.Palette.ContrastBg
				button.Color = th.Palette.ContrastFg
				button.Inset = layout.UniformInset(unit.Dp(8))
				d := button.Layout(gtx)
				return d
			}),
		)
	})
}

func (p *page) drawContactsItems(gtx Gtx) Dim {
	if <-p.Service().AccountsCount() == 0 {
		return p.NoAccount.Layout(gtx)
	}

	if len(p.contactsView) == 0 {
		return p.NoContact.Layout(gtx)
	}

	return p.List.Layout(gtx, len(p.contactsView), func(gtx Gtx, index int) (d Dim) {
		return p.contactsView[index].Layout(gtx)
	})
}

func (p *page) drawMenuLayout(gtx Gtx) Dim {
	if p.btnBackdrop.Clicked() {
		p.menuVisibilityAnim.Disappear(gtx.Now)
	}

	layout.Stack{Alignment: layout.NE}.Layout(gtx,
		layout.Stacked(func(gtx Gtx) Dim {
			return p.btnBackdrop.Layout(gtx, func(gtx Gtx) Dim {
				progress := p.menuVisibilityAnim.Revealed(gtx)
				gtx.Constraints.Max.X = int(float32(gtx.Constraints.Max.X) * progress)
				gtx.Constraints.Max.Y = int(float32(gtx.Constraints.Max.Y) * progress)
				return component.Rect{Size: gtx.Constraints.Max, Color: color.NRGBA{A: 200}}.Layout(gtx)
			})
		}),
		layout.Stacked(func(gtx Gtx) Dim {
			progress := p.menuVisibilityAnim.Revealed(gtx)
			macro := op.Record(gtx.Ops)
			d := p.btnMenuContent.Layout(gtx, p.drawMenuItems)
			call := macro.Stop()
			d.Size.X = int(float32(d.Size.X) * progress)
			d.Size.Y = int(float32(d.Size.Y) * progress)
			component.Rect{Size: d.Size, Color: color.NRGBA(colornames.White)}.Layout(gtx)
			clipOp := clip.Rect{Max: d.Size}.Push(gtx.Ops)
			call.Add(gtx.Ops)
			clipOp.Pop()
			return d
		}),
	)
	return Dim{}
}

func (p *page) drawMenuItems(gtx Gtx) Dim {
	gtx.Constraints.Max.X = int(float32(gtx.Constraints.Max.X) / 1.5)
	gtx.Constraints.Min.X = gtx.Constraints.Max.X
	if p.SelectionMode {
		return p.drawSelectionMenuItems(gtx)
	}
	return p.drawNormalMenuItems(gtx)
}

func (p *page) drawNormalMenuItems(gtx Gtx) Dim {
	if p.btnSelectAll.Clicked() {
		p.selectAll()
		p.menuVisibilityAnim.Disappear(gtx.Now)
	}
	if p.btnSelectionMode.Clicked() {
		p.SelectionMode = true
		p.menuVisibilityAnim.Disappear(gtx.Now)
	}
	if p.btnDeleteAll.Clicked() {
		p.selectAll()
		p.Modal().Show(p.drawDeleteContactsModal, nil, Animation{
			Duration: time.Millisecond * 250,
			State:    component.Invisible,
			Started:  time.Time{},
		})
		p.menuVisibilityAnim.Disappear(gtx.Now)
	}

	return layout.Flex{Axis: layout.Vertical, Alignment: layout.Start}.Layout(gtx,
		p.drawMenuItem("Add Contact", &p.btnAddContact),
		p.drawMenuItem("Selection Mode", &p.btnSelectionMode),
		p.drawMenuItem("Select All Contacts", &p.btnSelectAll),
		p.drawMenuItem("Delete All Contacts", &p.btnDeleteAll),
	)
}
func (p *page) drawSelectionMenuItems(gtx Gtx) Dim {
	if p.btnDeleteSelection.Clicked() {
		p.Modal().Show(p.drawDeleteContactsModal, nil, Animation{
			Duration: time.Millisecond * 250,
			State:    component.Invisible,
			Started:  time.Time{},
		})
		p.menuVisibilityAnim.Disappear(gtx.Now)
	}
	return layout.Flex{Axis: layout.Vertical, Alignment: layout.Start}.Layout(gtx,
		p.drawMenuItem("Delete Selected Contacts", &p.btnDeleteSelection),
		p.drawMenuItem("Clear Selection", &p.btnCloseSelection),
	)
}

func (p *page) drawMenuItem(txt string, btn *widget.Clickable) layout.FlexChild {
	inset := layout.UniformInset(unit.Dp(12))
	return layout.Rigid(func(gtx layout.Context) layout.Dimensions {
		a := p.Service().Account()
		if a.PublicKey == "" {
			return Dim{}
		}
		btnStyle := material.ButtonLayoutStyle{Button: btn}
		btnStyle.Background = color.NRGBA(colornames.White)
		return btnStyle.Layout(gtx,
			func(gtx Gtx) Dim {
				gtx.Constraints.Min.X = gtx.Constraints.Max.X
				inset := inset
				return inset.Layout(gtx, func(gtx Gtx) Dim {
					return layout.Flex{Spacing: layout.SpaceEnd}.Layout(gtx,
						layout.Rigid(func(gtx Gtx) Dim {
							bd := material.Body1(p.Theme, txt)
							bd.Color = color.NRGBA(colornames.Black)
							bd.Alignment = text.Start
							return bd.Layout(gtx)
						}),
					)
				})
			},
		)
	})
}

func (p *page) drawAddContactModal(gtx Gtx) Dim {
	gtx.Constraints.Max.X = int(float32(gtx.Constraints.Max.X) * 0.85)
	gtx.Constraints.Max.Y = int(float32(gtx.Constraints.Max.Y) * 0.85)
	return p.ModalContent.DrawContent(gtx, p.Theme, p.ContactForm.Layout)
}

func (p *page) drawDeleteContactsModal(gtx Gtx) Dim {
	gtx.Constraints.Max.X = int(float32(gtx.Constraints.Max.X) * 0.85)
	gtx.Constraints.Max.Y = int(float32(gtx.Constraints.Max.Y) * 0.85)
	if p.btnYes.Clicked() {
		contacts := make([]service.Contact, 0)
		contactsViewSize := len(p.contactsView)
		for _, eachView := range p.contactsView {
			if eachView.Selected {
				contacts = append(contacts, eachView.Contact)
			}
		}
		<-p.Service().DeleteContacts(contacts)
		var txtTmp string
		if len(contacts) == contactsViewSize {
			txtTmp = "all accounts."
		} else {
			txtTmp = fmt.Sprintf("%d contacts.", len(contacts))
		}
		if len(contacts) == 1 {
			txtTmp = "1 contact."
		}
		txt := fmt.Sprintf("Successfully deleted %s", txtTmp)
		p.Modal().Dismiss(func() {
			p.clearAllSelection()
			p.Snackbar().Show(txt, nil, color.NRGBA{}, "")
		})
	}
	if p.btnNo.Clicked() {
		p.Modal().Dismiss(func() {
			p.clearAllSelection()
		})
	}
	count := p.getSelectionCount()
	accountsSize := len(p.contactsView)
	var txt string
	if count == accountsSize {
		txt = "all contacts"
	} else {
		txt = fmt.Sprintf("%d selected contacts", count)
	}
	if count == 1 {
		txt = "the selected contact"
	}
	promptContent := view.NewPromptContent(p.Theme,
		"Contacts Deletion!",
		fmt.Sprintf("Are you sure you want to delete %s?", txt),
		&p.btnYes, &p.btnNo)
	return p.ModalContent.DrawContent(gtx, p.Theme, promptContent.Layout)
}

func (p *page) onAddContactSuccess(addr string) {
	p.Modal().Dismiss(func() {
		p.ContactForm = view.NewContactForm(p.Manager, service.Contact{}, p.onAddContactSuccess)
		txt := fmt.Sprintf("Successfully added contact %s", addr)
		p.Snackbar().Show(txt, nil, color.NRGBA{}, "")
	})
}

func (p *page) getSelectionCount() (count int) {
	for _, item := range p.contactsView {
		if item.Selected {
			count++
		}
	}
	return count
}

func (p *page) clearAllSelection() {
	p.SelectionMode = false
	for _, item := range p.contactsView {
		item.Selected = false
		item.SelectionMode = false
	}
}
func (p *page) selectAll() {
	p.SelectionMode = true
	for _, item := range p.contactsView {
		item.Selected = true
		item.SelectionMode = true
	}
}

func (p *page) fetchContactsOnScroll(_ Gtx) {
	p.listPosition = p.Position
	shouldFetch := p.Position.First == 0 && !p.isFetchingContacts && int64(len(p.contactsView)) < p.contactsCount
	if shouldFetch {
		currentSize := len(p.contactsView) + defaultListSize
		p.fetchContacts(0, currentSize)
	}
}

func (p *page) fetchContacts(offset, limit int) {
	if !p.isFetchingContacts {
		p.isFetchingContacts = true
		go func(offset int, limit int) {
			accountPublicKey := p.Service().Account().PublicKey
			p.fetchingContactsCh <- <-p.Service().Contacts(accountPublicKey, offset, limit)
			p.Window().Invalidate()
		}(offset, limit)
	}
}
func (p *page) fetchContactsCount() {
	if !p.isFetchingContactsCount {
		p.isFetchingContactsCount = true
		go func() {
			p.fetchingContactsCountCh <- <-p.Service().ContactsCount(p.Service().Account().PublicKey)
			p.Window().Invalidate()
		}()
	}
}
func (p *page) listenToFetchContacts() {
	shouldBreak := false
	for {
		select {
		case contacts := <-p.fetchingContactsCh:
			// reversing
			contactsView := make([]*pageItem, len(contacts))
			for i, eachContact := range contacts {
				contactsView[i] = &pageItem{
					Theme:   p.Theme,
					Manager: p.Manager,
					Contact: eachContact,
				}
			}
			//pos := p.Position.First
			p.contactsView = contactsView
			//p.Position.First = pos + len(contacts)
			p.isFetchingContacts = false
		default:
			shouldBreak = true
		}
		if shouldBreak {
			break
		}
	}
}
func (p *page) listenToFetchContactsCount() {
	shouldBreak := false
	for {
		select {
		case contactsCount := <-p.fetchingContactsCountCh:
			if contactsCount != p.contactsCount {
				p.contactsCount = contactsCount
				if !p.isFetchingContacts {
					p.fetchContacts(0, len(p.contactsView))
				}
			}
			p.isFetchingContactsCount = false
		default:
			shouldBreak = true
		}
		if shouldBreak {
			break
		}
	}
	if p.Theme == nil {
		p.Theme = p.Manager.Theme()
	}
}

func (p *page) OnDatabaseChange(event service.Event) {
	switch event.Data.(type) {
	case service.AccountChangedEventData, service.AccountsChangedEventData:
		p.fetchContactsCount()
		if len(p.contactsView) == 0 {
			p.fetchContacts(0, defaultListSize)
		} else {
			p.fetchContacts(0, defaultListSize)
		}
	case service.ContactsChangeEventData:
		p.fetchContactsCount()
		if len(p.contactsView) == 0 {
			p.fetchContacts(0, defaultListSize)
		} else {
			p.fetchContacts(0, defaultListSize)
		}
	}
}

func (p *page) URL() URL {
	return ContactsPageURL
}
