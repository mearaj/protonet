package accounts

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

type page struct {
	layout.List
	Manager
	Theme                   *material.Theme
	title                   string
	iconNewChat             *widget.Icon
	btnBackdrop             widget.Clickable
	buttonNavigation        widget.Clickable
	btnMenuIcon             widget.Clickable
	btnMenuContent          widget.Clickable
	btnAddAccount           widget.Clickable
	btnDeleteAccounts       widget.Clickable
	btnCloseSelection       widget.Clickable
	btnYes                  widget.Clickable
	btnNo                   widget.Clickable
	btnSelectAll            widget.Clickable
	btnDeleteAll            widget.Clickable
	btnSelectionMode        widget.Clickable
	menuIcon                *widget.Icon
	closeIcon               *widget.Icon
	menuVisibilityAnim      component.VisibilityAnimation
	navigationIcon          *widget.Icon
	accountsView            []*pageItem
	NoAccount               View
	AccountForm             View
	ModalContent            *view.ModalContent
	SelectionMode           bool
	isFetchingAccounts      bool
	isFetchingAccountsCount bool
	initialized             bool
	subscription            service.Subscriber
	fetchingAccountsCh      chan []service.Account
	fetchingAccountsCountCh chan int64
	accountsCount           int64
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
		title:                   "Accounts",
		navigationIcon:          navIcon,
		iconNewChat:             iconNewChat,
		List:                    layout.List{Axis: layout.Vertical},
		accountsView:            []*pageItem{},
		menuIcon:                iconMenu,
		closeIcon:               closeIcon,
		fetchingAccountsCh:      make(chan []service.Account, 10),
		fetchingAccountsCountCh: make(chan int64, 10),
		menuVisibilityAnim: component.VisibilityAnimation{
			Duration: time.Millisecond * 250,
			State:    component.Invisible,
			Started:  time.Time{},
		},
	}
	p.AccountForm = view.NewAccountFormView(manager, p.onSuccess)
	p.ModalContent = view.NewModalContent(func() {
		p.Modal().Dismiss(nil)
		p.AccountForm = view.NewAccountFormView(manager, p.onSuccess)
	})
	p.NoAccount = view.NewNoAccount(manager)
	p.subscription = manager.Service().Subscribe(service.AccountsChangedEventTopic)
	return &p
}

func (p *page) Layout(gtx Gtx) Dim {
	if !p.initialized {
		if p.Theme == nil {
			p.Theme = p.Manager.Theme()
		}
		p.fetchAccounts()
		p.fetchAccountsCount()
		p.initialized = true
	}

	p.listenToFetchAccounts()
	p.listenToFetchAccountsCount()
	p.handleSelectionMode()
	p.handleAddAccountClick(gtx)

	flex := layout.Flex{Axis: layout.Vertical, Spacing: layout.SpaceEnd, Alignment: layout.Start}

	d := flex.Layout(gtx,
		layout.Rigid(p.DrawAppBar),
		layout.Rigid(p.drawIdentitiesItems),
	)
	p.drawMenuLayout(gtx)
	p.handleEvents(gtx)
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
	gtx.Constraints.Max.Y = gtx.Dp(56)
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
	gtx.Constraints.Max.Y = gtx.Dp(56)
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
						button := material.IconButton(th, &p.btnCloseSelection, closeIcon, "Nav Icon Button")
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

func (p *page) drawIdentitiesItems(gtx Gtx) Dim {

	if len(p.accountsView) == 0 {
		return p.NoAccount.Layout(gtx)
	}
	return p.List.Layout(gtx, len(p.accountsView), func(gtx Gtx, index int) (d Dim) {
		inset := layout.Inset{Top: unit.Dp(0), Bottom: unit.Dp(0)}
		return inset.Layout(gtx, func(gtx Gtx) Dim {
			return p.accountsView[index].Layout(gtx)
		})
	})
}

func (p *page) drawMenuLayout(gtx Gtx) Dim {
	if p.btnBackdrop.Clicked() {
		if !p.btnMenuContent.Pressed() {
			p.menuVisibilityAnim.Disappear(gtx.Now)
		}
		for _, idView := range p.accountsView {
			if !idView.btnMenuContent.Pressed() && !idView.Hovered() {
				idView.menuVisibilityAnim.Disappear(gtx.Now)
			}
		}
	}
	layout.Stack{Alignment: layout.NE}.Layout(gtx,
		layout.Stacked(func(gtx Gtx) Dim {
			return p.btnBackdrop.Layout(gtx,
				func(gtx layout.Context) layout.Dimensions {
					progress := p.menuVisibilityAnim.Revealed(gtx)
					gtx.Constraints.Max.X = int(float32(gtx.Constraints.Max.X) * progress)
					gtx.Constraints.Max.Y = int(float32(gtx.Constraints.Max.Y) * progress)
					return component.Rect{Size: gtx.Constraints.Max, Color: color.NRGBA{A: 200}}.Layout(gtx)
				},
			)
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
		p.Modal().Show(p.drawDeleteAccountsModal, nil, Animation{
			Duration: time.Millisecond * 250,
			State:    component.Invisible,
			Started:  time.Time{},
		})
		p.menuVisibilityAnim.Disappear(gtx.Now)
	}

	return layout.Flex{Axis: layout.Vertical, Alignment: layout.Start}.Layout(gtx,
		p.drawMenuItem("Add Account", &p.btnAddAccount),
		p.drawMenuItem("Selection Mode", &p.btnSelectionMode),
		p.drawMenuItem("Select All Accounts", &p.btnSelectAll),
		p.drawMenuItem("Delete All Accounts", &p.btnDeleteAll),
	)
}
func (p *page) drawSelectionMenuItems(gtx Gtx) Dim {
	if p.btnDeleteAccounts.Clicked() {
		p.Modal().Show(p.drawDeleteAccountsModal, nil, Animation{
			Duration: time.Millisecond * 250,
			State:    component.Invisible,
			Started:  time.Time{},
		})
		p.menuVisibilityAnim.Disappear(gtx.Now)
	}

	return layout.Flex{Axis: layout.Vertical, Alignment: layout.Start}.Layout(gtx,
		p.drawMenuItem("Delete Selected Accounts", &p.btnDeleteAccounts),
		p.drawMenuItem("Clear Selection", &p.btnCloseSelection),
	)
}
func (p *page) drawMenuItem(txt string, btn *widget.Clickable) layout.FlexChild {
	inset := layout.UniformInset(unit.Dp(12))
	return layout.Rigid(func(gtx layout.Context) layout.Dimensions {
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

func (p *page) drawAddAccountModal(gtx Gtx) Dim {
	gtx.Constraints.Max.X = int(float32(gtx.Constraints.Max.X) * 0.85)
	gtx.Constraints.Max.Y = int(float32(gtx.Constraints.Max.Y) * 0.85)
	return p.ModalContent.DrawContent(gtx, p.Theme, p.AccountForm.Layout)
}

func (p *page) drawDeleteAccountsModal(gtx Gtx) Dim {
	gtx.Constraints.Max.X = int(float32(gtx.Constraints.Max.X) * 0.85)
	gtx.Constraints.Max.Y = int(float32(gtx.Constraints.Max.Y) * 0.85)
	if p.btnYes.Clicked() {
		accounts := make([]service.Account, 0)
		accountsViewSize := len(p.accountsView)
		for _, eachView := range p.accountsView {
			if eachView.Selected {
				accounts = append(accounts, eachView.Account)
			}
		}
		<-p.Service().DeleteAccounts(accounts)
		p.Modal().Dismiss(func() {
			p.Window().Invalidate()
			p.clearAllSelection()
			var txtTmp string
			if len(accounts) == accountsViewSize {
				txtTmp = "all accounts."
			} else {
				txtTmp = fmt.Sprintf("%d accounts.", len(accounts))
			}
			if len(accounts) == 1 {
				txtTmp = "1 account."
			}
			txt := fmt.Sprintf("Successfully deleted %s", txtTmp)
			p.Snackbar().Show(txt, nil, color.NRGBA{}, "")
		})
	}
	if p.btnNo.Clicked() {
		p.Modal().Dismiss(func() {
			p.clearAllSelection()
		})
	}
	count := p.getSelectionCount()
	accountsSize := len(p.accountsView)
	var txt string
	if count == accountsSize {
		txt = "all accounts"
	} else {
		txt = fmt.Sprintf("%d selected accounts", count)
	}
	if count == 1 {
		txt = "the selected account"
	}
	promptContent := view.NewPromptContent(p.Theme,
		"Account Deletion!",
		fmt.Sprintf("Are you sure you want to delete %s?", txt),
		&p.btnYes, &p.btnNo)
	return p.ModalContent.DrawContent(gtx, p.Theme, promptContent.Layout)
}

func (p *page) onSuccess() {
	p.Modal().Dismiss(func() {
		p.AccountForm = view.NewAccountFormView(p.Manager, p.onSuccess)
		a := p.Service().Account()
		txt := fmt.Sprintf("Successfully created %s", a.PublicKey)
		p.Window().Invalidate()
		p.Snackbar().Show(txt, nil, color.NRGBA{}, "")
	})
}
func (p *page) getSelectionCount() (count int) {
	for _, item := range p.accountsView {
		if item.Selected {
			count++
		}
	}
	return count
}
func (p *page) clearAllSelection() {
	p.SelectionMode = false
	for _, item := range p.accountsView {
		item.Selected = false
		item.SelectionMode = false
	}
}
func (p *page) selectAll() {
	p.SelectionMode = true
	for _, item := range p.accountsView {
		item.Selected = true
		item.SelectionMode = true
	}
}

func (p *page) fetchAccounts() {
	if !p.isFetchingAccounts {
		p.isFetchingAccounts = true
		go func() {
			p.fetchingAccountsCh <- <-p.Service().Accounts()
			p.Window().Invalidate()
		}()
	}
}
func (p *page) fetchAccountsCount() {
	if !p.isFetchingAccountsCount {
		p.isFetchingAccountsCount = true
		go func() {
			p.fetchingAccountsCountCh <- <-p.Service().AccountsCount()
			p.Window().Invalidate()
		}()
	}
}

func (p *page) listenToFetchAccounts() {
	shouldBreak := false
	for {
		select {
		case accounts := <-p.fetchingAccountsCh:
			accountViews := make([]*pageItem, len(accounts))
			for i, eachContact := range accounts {
				accountViews[i] = &pageItem{
					Theme:        p.Theme,
					Manager:      p.Manager,
					Account:      eachContact,
					ModalContent: p.ModalContent,
				}
			}
			//pos := p.Position.First
			p.accountsView = accountViews
			//p.Position.First = pos + len(accounts)
			p.isFetchingAccounts = false
		default:
			shouldBreak = true
		}
		if shouldBreak {
			break
		}
	}

}

func (p *page) listenToFetchAccountsCount() {
	shouldBreak := false
	for {
		select {
		case accountsCount := <-p.fetchingAccountsCountCh:
			if accountsCount != p.accountsCount {
				p.accountsCount = accountsCount
				if !p.isFetchingAccounts {
					p.fetchAccounts()
				}
			}
			p.isFetchingAccountsCount = false
		default:
			shouldBreak = true
		}
		if shouldBreak {
			break
		}
	}
}

func (p *page) handleSelectionMode() {
	for _, item := range p.accountsView {
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
}

func (p *page) handleAddAccountClick(gtx Gtx) {
	if p.btnAddAccount.Clicked() {
		p.Modal().Show(p.drawAddAccountModal, func() {
			p.AccountForm = view.NewAccountFormView(p.Manager, p.onSuccess)
			p.Modal().Dismiss(nil)
		}, Animation{
			Duration: time.Millisecond * 250,
			State:    component.Invisible,
			Started:  time.Time{},
		})
		p.menuVisibilityAnim.Disappear(gtx.Now)
	}
}

func (p *page) handleEvents(gtx Gtx) {
	for _, e := range gtx.Queue.Events(p) {
		switch e := e.(type) {
		case pointer.Event:
			switch e.Type {
			case pointer.Press:
				if !p.btnMenuContent.Pressed() {
					p.menuVisibilityAnim.Disappear(gtx.Now)
				}
				for _, idView := range p.accountsView {
					if !idView.btnMenuContent.Pressed() && !idView.Hovered() {
						idView.menuVisibilityAnim.Disappear(gtx.Now)
					}
				}
			}
		}
	}
}

func (p *page) OnDatabaseChange(event service.Event) {
	switch e := event.Data.(type) {
	case service.AccountsChangedEventData:
		_ = e
		p.fetchAccounts()
		p.fetchAccountsCount()
	}
}
func (p *page) URL() URL {
	return AccountsPageURL
}
