package chat

import (
	"gioui.org/io/pointer"
	"gioui.org/layout"
	"gioui.org/op"
	"gioui.org/op/clip"
	"gioui.org/text"
	"gioui.org/unit"
	"gioui.org/widget"
	"gioui.org/widget/material"
	"gioui.org/x/component"
	"github.com/mearaj/protonet/alog"
	"github.com/mearaj/protonet/internal/chat"
	"github.com/mearaj/protonet/internal/pubsub"
	"github.com/mearaj/protonet/internal/wallet"
	. "github.com/mearaj/protonet/ui/fwk"
	"github.com/mearaj/protonet/ui/page/chatroom"
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
	*material.Theme
	btnAddContact           widget.Clickable
	btnAddAccount           widget.Clickable
	btnNavIcon              widget.Clickable
	btnBackdrop             widget.Clickable
	btnMenuContent          widget.Clickable
	btnMenuIcon             widget.Clickable
	btnAccountDetails       widget.Clickable
	navIcon                 *widget.Icon
	menuIcon                *widget.Icon
	menuVisibilityAnim      component.VisibilityAnimation
	NoAccount               *view.NoAccountView
	PasswordForm            ViewWithDBListener
	NoContact               *view.NoContactView
	AccountForm             View
	ContactForm             View
	AccountDetails          *view.AccountDetails
	chatPageItems           []*pageItem
	isFetchingContacts      bool
	isFetchingContactsCount bool
	listPosition            layout.Position
	contactsCount           int64
	ModalContent            *view.ModalContent
	initialized             bool
}

func New(manager Manager) Page {
	th := manager.Theme()
	iconNav, _ := widget.NewIcon(icons.NavigationArrowBack)
	iconMenu, _ := widget.NewIcon(icons.NavigationMoreVert)
	p := page{
		Manager:       manager,
		Theme:         th,
		chatPageItems: make([]*pageItem, 0),
		List:          layout.List{Axis: layout.Vertical},
		navIcon:       iconNav,
		menuIcon:      iconMenu,
		menuVisibilityAnim: component.VisibilityAnimation{
			Duration: time.Millisecond * 250,
			State:    component.Invisible,
			Started:  time.Time{},
		},
	}
	p.AccountForm = view.NewAccountFormView(manager, p.onAddAccountSuccess)
	p.ContactForm = view.NewContactForm(manager, chat.Contact{}, p.onAddContactSuccess)
	p.ModalContent = view.NewModalContent(func() {
		p.Modal().Dismiss(nil)
		acc, _ := wallet.GlobalWallet.Account()
		p.AccountDetails = view.NewAccountDetails(p.Manager, acc)
	})
	p.NoAccount = view.NewNoAccount(manager)
	p.PasswordForm = view.NewPasswordForm(manager, func() {})
	p.NoContact = view.NewNoContact(manager, p.onAddContactSuccess, "New Chat")
	return &p
}
func (p *page) Layout(gtx Gtx) (d Dim) {
	if !p.initialized {
		if p.Theme == nil {
			p.Theme = p.Manager.Theme()
		}
		p.fetchContacts(0, defaultListSize)
		p.fetchContactsCount()
		p.initialized = true
	}
	p.fetchContactsOnScroll(gtx)

	a, _ := wallet.GlobalWallet.Account()
	if p.btnAddAccount.Clicked() {
		p.AccountForm = view.NewAccountFormView(p.Manager, p.onAddAccountSuccess)
		p.menuVisibilityAnim.Disappear(gtx.Now)
		p.Modal().Show(p.drawAddAccountModal, nil, Animation{
			Duration: time.Millisecond * 250,
			State:    component.Invisible,
			Started:  time.Time{},
		})
	}
	if p.btnAddContact.Clicked() {
		p.ContactForm = view.NewContactForm(p.Manager, chat.Contact{}, p.onAddContactSuccess)
		p.Modal().Show(p.drawAddContactModal, nil, Animation{
			Duration: time.Millisecond * 250,
			State:    component.Invisible,
			Started:  time.Time{},
		})
		//p.menuVisibilityAnim.Disappear(gtx.Now)
	}
	if p.btnAccountDetails.Clicked() {
		if p.AccountDetails == nil || p.AccountDetails.Account.PublicKey != a.PublicKey {
			p.AccountDetails = view.NewAccountDetails(p.Manager, a)
		}
		p.Modal().Show(p.drawAccountDetailsModal, func() {
			p.AccountDetails = view.NewAccountDetails(p.Manager, a)
			p.Modal().Dismiss(nil)
		}, Animation{
			Duration: time.Millisecond * 250,
			State:    component.Invisible,
			Started:  time.Time{},
		})
		p.menuVisibilityAnim.Disappear(gtx.Now)
	}

	if p.chatPageItems == nil {
		p.chatPageItems = make([]*pageItem, 0)
	}
	flex := layout.Flex{Axis: layout.Vertical,
		Spacing:   layout.SpaceEnd,
		Alignment: layout.Start,
	}

	d = flex.Layout(gtx,
		layout.Rigid(p.DrawAppBar),
		layout.Rigid(p.drawChatItems),
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
			}
		}
	}
	return d
}
func (p *page) DrawAppBar(gtx Gtx) Dim {
	if p.btnNavIcon.Clicked() {
		p.PopUp()
	}
	if p.btnMenuIcon.Clicked() {
		p.menuVisibilityAnim.Appear(gtx.Now)
	}

	return view.DrawAppBarLayout(gtx, p.Manager.Theme(), func(gtx Gtx) Dim {
		return layout.Flex{Alignment: layout.Middle, Spacing: layout.SpaceBetween}.Layout(gtx,
			layout.Rigid(func(gtx Gtx) Dim {
				return layout.Flex{Alignment: layout.Middle}.Layout(gtx,
					layout.Rigid(func(gtx Gtx) Dim {
						button := material.IconButton(p.Manager.Theme(), &p.btnNavIcon, p.navIcon, "Navigates to settings page")
						button.Size = unit.Dp(40)
						button.Background = p.Manager.Theme().Palette.ContrastBg
						button.Color = p.Manager.Theme().Palette.ContrastFg
						button.Inset = layout.UniformInset(unit.Dp(8))
						return button.Layout(gtx)
					}),
					layout.Rigid(func(gtx Gtx) Dim {
						gtx.Constraints.Max.X = gtx.Constraints.Max.X - gtx.Dp(56)
						return layout.Inset{Left: unit.Dp(8)}.Layout(gtx, func(gtx Gtx) Dim {
							titleText := "Protonet"
							a, _ := wallet.GlobalWallet.Account()
							accs, _ := wallet.GlobalWallet.Accounts()
							if a.PublicKey != "" && len(accs) != 0 {
								titleText = a.PublicKey
							}
							label := material.Label(p.Manager.Theme(), unit.Sp(18), titleText)
							label.Color = p.Manager.Theme().ContrastFg
							return component.TruncatingLabelStyle(label).Layout(gtx)
						})
					}),
				)
			}),
			layout.Rigid(func(gtx Gtx) Dim {
				button := material.IconButton(p.Manager.Theme(), &p.btnMenuIcon, p.menuIcon, "Context Menu")
				button.Size = unit.Dp(40)
				button.Background = p.Manager.Theme().Palette.ContrastBg
				button.Color = p.Manager.Theme().Palette.ContrastFg
				button.Inset = layout.UniformInset(unit.Dp(8))
				d := button.Layout(gtx)
				return d
			}),
		)
	})
}
func (p *page) drawChatItems(gtx Gtx) Dim {
	isPasswordSet := wallet.GlobalWallet.IsOpen()
	if !isPasswordSet {
		return p.PasswordForm.Layout(gtx)
	}
	accs, _ := wallet.GlobalWallet.Accounts()
	if len(accs) == 0 {
		return p.NoAccount.Layout(gtx)
	}

	if len(p.chatPageItems) == 0 {
		return p.NoContact.Layout(gtx)
	}

	return p.List.Layout(gtx, len(p.chatPageItems), func(gtx Gtx, index int) (d Dim) {
		inset := layout.Inset{Top: unit.Dp(0), Bottom: unit.Dp(0)}
		return inset.Layout(gtx, func(gtx Gtx) Dim {
			return p.chatPageItems[index].Layout(gtx)
		})
	})
}

func (p *page) drawMenuLayout(gtx Gtx) Dim {
	if p.btnBackdrop.Clicked() {
		p.menuVisibilityAnim.Disappear(gtx.Now)
		p.AccountForm = view.NewAccountFormView(p.Manager, p.onAddAccountSuccess)
		p.ContactForm = view.NewContactForm(p.Manager, chat.Contact{}, p.onAddContactSuccess)
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
	inset := layout.UniformInset(unit.Dp(12))
	gtx.Constraints.Max.X = int(float32(gtx.Constraints.Max.X) / 1.5)
	gtx.Constraints.Min.X = gtx.Constraints.Max.X
	a, _ := wallet.GlobalWallet.Account()
	return layout.Flex{Axis: layout.Vertical, Alignment: layout.Start}.Layout(gtx,
		layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			if a.PublicKey == "" {
				return Dim{}
			}
			btnStyle := material.ButtonLayoutStyle{Button: &p.btnAccountDetails}
			btnStyle.Background = color.NRGBA(colornames.White)
			return btnStyle.Layout(gtx,
				func(gtx Gtx) Dim {
					gtx.Constraints.Min.X = gtx.Constraints.Max.X
					inset := inset
					return inset.Layout(gtx, func(gtx Gtx) Dim {
						return layout.Flex{Spacing: layout.SpaceEnd}.Layout(gtx,
							layout.Rigid(func(gtx Gtx) Dim {
								bd := material.Body1(p.Theme, "Account Details")
								bd.Color = color.NRGBA(colornames.Black)
								bd.Alignment = text.Start
								return bd.Layout(gtx)
							}),
						)
					})
				},
			)
		}),
		layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			btnStyle := material.ButtonLayoutStyle{Button: &p.btnAddAccount}
			btnStyle.Background = color.NRGBA(colornames.White)
			return btnStyle.Layout(gtx,
				func(gtx Gtx) Dim {
					gtx.Constraints.Min.X = gtx.Constraints.Max.X
					inset := inset
					return inset.Layout(gtx, func(gtx Gtx) Dim {
						return layout.Flex{Spacing: layout.SpaceEnd}.Layout(gtx,
							layout.Rigid(func(gtx Gtx) Dim {
								bd := material.Body1(p.Theme, "Add Account")
								bd.Color = color.NRGBA(colornames.Black)
								bd.Alignment = text.Start
								return bd.Layout(gtx)
							}),
						)
					})
				},
			)
		}),
		layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			if a.PublicKey == "" {
				return Dim{}
			}
			btnStyle := material.ButtonLayoutStyle{Button: &p.btnAddContact}
			btnStyle.Background = color.NRGBA(colornames.White)
			return btnStyle.Layout(gtx,
				func(gtx Gtx) Dim {
					gtx.Constraints.Min.X = gtx.Constraints.Max.X
					inset := inset
					return inset.Layout(gtx, func(gtx Gtx) Dim {
						return layout.Flex{Spacing: layout.SpaceEnd}.Layout(gtx,
							layout.Rigid(func(gtx Gtx) Dim {
								bd := material.Body1(p.Theme, "New Chat")
								bd.Color = color.NRGBA(colornames.Black)
								bd.Alignment = text.Start
								return bd.Layout(gtx)
							}),
						)
					})
				},
			)
		}),
	)
}

func (p *page) fetchContacts(offset, limit int) {
	if !p.isFetchingContacts {
		p.isFetchingContacts = true
		account, _ := wallet.GlobalWallet.Account()
		accountPublicKey := account.PublicKey
		contacts, _ := wallet.GlobalWallet.Contacts(accountPublicKey, offset, limit)
		p.isFetchingContacts = false
		pageItems := make([]*pageItem, len(contacts))
		for i, eachContact := range contacts {
			pageItems[i] = &pageItem{
				Theme:   p.Theme,
				Manager: p.Manager,
				contact: eachContact,
			}
		}
		p.chatPageItems = pageItems
		//pos := p.Position.First
		//p.Position.First = pos + len(contacts)
	}
}
func (p *page) fetchContactsOnScroll(_ Gtx) {
	p.listPosition = p.Position
	shouldFetch := p.Position.First == 0 && !p.isFetchingContacts && int64(len(p.chatPageItems)) < p.contactsCount
	if shouldFetch {
		currentSize := len(p.chatPageItems) + defaultListSize
		p.fetchContacts(0, currentSize)
	}
}

func (p *page) drawAddAccountModal(gtx Gtx) Dim {
	gtx.Constraints.Max.X = int(float32(gtx.Constraints.Max.X) * 0.85)
	gtx.Constraints.Max.Y = int(float32(gtx.Constraints.Max.Y) * 0.85)
	return p.ModalContent.DrawContent(gtx, p.Theme, p.AccountForm.Layout)
}

func (p *page) drawAddContactModal(gtx Gtx) Dim {
	gtx.Constraints.Max.X = int(float32(gtx.Constraints.Max.X) * 0.85)
	gtx.Constraints.Max.Y = int(float32(gtx.Constraints.Max.Y) * 0.85)
	return p.ModalContent.DrawContent(gtx, p.Theme, p.ContactForm.Layout)
}
func (p *page) drawAccountDetailsModal(gtx Gtx) Dim {
	gtx.Constraints.Max.X = int(float32(gtx.Constraints.Max.X) * 0.85)
	gtx.Constraints.Max.Y = int(float32(gtx.Constraints.Max.Y) * 0.85)
	return p.ModalContent.DrawContent(gtx, p.Theme, p.AccountDetails.Layout)
}

func (p *page) onAddAccountSuccess() {
	p.Modal().Dismiss(func() {
		p.AccountForm = view.NewAccountFormView(p.Manager, p.onAddAccountSuccess)
		p.Manager.Window().Invalidate()
	})
}

func (p *page) OnDatabaseChange(event pubsub.Event) {
	acc, _ := wallet.GlobalWallet.Account()
	switch e := event.Data.(type) {
	case pubsub.CurrentAccountChangedEventData, pubsub.AccountsChangedEventData:
		p.fetchContacts(0, defaultListSize)
	case pubsub.ContactsChangeEventData:
		if e.AccountPublicKey == acc.PublicKey {
			if len(p.chatPageItems) == 0 {
				p.fetchContacts(0, defaultListSize)
			} else {
				p.fetchContacts(0, len(p.chatPageItems))
			}
		}
	}
	for _, chatItem := range p.chatPageItems {
		chatItem.OnDatabaseChange(event)
	}
	p.PasswordForm.OnDatabaseChange(event)
}

func (p *page) fetchContactsCount() {
	if !p.isFetchingContactsCount {
		p.isFetchingContactsCount = true
		var err error
		var count int64
		acc, _ := wallet.GlobalWallet.Account()
		count, err = wallet.GlobalWallet.ContactsCount(acc.PublicKey)
		if err != nil {
			alog.Logger().Errorln(err)
		}
		p.isFetchingContactsCount = false
		if p.contactsCount != count {
			p.contactsCount = count
			p.Window().Invalidate()
		}
	}
}

func (p *page) onAddContactSuccess(addr string) {
	p.Modal().Dismiss(func() {
		p.ContactForm = view.NewContactForm(p.Manager, chat.Contact{}, p.onAddContactSuccess)
		chatRoomPage := chatroom.New(p.Manager, chat.Contact{PublicKey: addr})
		p.NavigateToPage(chatRoomPage, nil)
		p.Window().Invalidate()
	})
}
func (p *page) URL() URL {
	return ChatPageURL
}
