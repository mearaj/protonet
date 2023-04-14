package wallet

import (
	"gioui.org/layout"
	"gioui.org/unit"
	"gioui.org/widget"
	"gioui.org/widget/material"
	"gioui.org/x/component"
	"github.com/mearaj/protonet/internal/pubsub"
	"github.com/mearaj/protonet/ui/fwk"
	"github.com/mearaj/protonet/ui/view"
	"golang.org/x/exp/shiny/materialdesign/colornames"
	"golang.org/x/exp/shiny/materialdesign/icons"
	"image/color"
)

type (
	Manager = fwk.Manager
	Gtx     = fwk.Gtx
	Dim     = fwk.Dim
	Page    = fwk.Page
	URL     = fwk.URL
)

type State int

const (
	StateIdle = State(iota)
	StateConnecting
	StateDisconnecting
)

type page struct {
	allChainsTab tabAllChains
	Manager
	Theme            *material.Theme
	title            string
	initialized      bool
	navigationIcon   *widget.Icon
	buttonNavigation widget.Clickable
	view.Tabs
	width  int
	search view.Search
}

func New(manager Manager) Page {
	errorTh := *manager.Theme()
	navIcon, _ := widget.NewIcon(icons.NavigationArrowBack)
	errorTh.ContrastBg = color.NRGBA(colornames.Red500)
	theme := *manager.Theme()
	p := page{
		Manager:        manager,
		Theme:          &theme,
		title:          "Wallet",
		navigationIcon: navIcon,
		allChainsTab:   tabAllChains{title: "All"},
	}
	p.allChainsTab.page = &p
	return &p
}

func (p *page) Layout(gtx Gtx) Dim {
	if !p.initialized {
		if p.Theme == nil {
			p.Theme = p.Manager.Theme()
		}
		p.allChainsTab.Axis = layout.Vertical
		p.initTabs()
		p.initialized = true
	}
	p.width = gtx.Constraints.Max.X
	flex := layout.Flex{Axis: layout.Vertical, Spacing: layout.SpaceEnd, Alignment: layout.Start}
	d := flex.Layout(gtx,
		layout.Rigid(p.DrawAppBar),
		layout.Rigid(p.Tabs.Layout),
	)
	return d
}

func (p *page) DrawAppBar(gtx Gtx) Dim {
	if p.buttonNavigation.Clicked() {
		p.PopUp()
	}
	gtx.Constraints.Max.Y = gtx.Dp(56)
	th := p.Theme

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
						if p.search.Animation.State == component.Visible {
							return Dim{}
						}
						d := layout.Inset{Left: unit.Dp(16)}.Layout(gtx, func(gtx Gtx) Dim {
							titleText := p.title
							title := material.Body1(th, titleText)
							title.Color = th.Palette.ContrastFg
							title.TextSize = unit.Sp(18)
							return title.Layout(gtx)
						})

						return d
					}),
				)
			}),
			layout.Rigid(p.search.Layout),
		)
	})
}

func (p *page) initTabs() {
	tabs := [2]view.Tab{}
	p.Tabs.Header = p.drawTabHead
	p.Tabs.Body = p.drawTabBody
	p.Tabs.Tabs = tabs[:]
}

func (p *page) drawTabHead(gtx Gtx, index int) Dim {
	switch index {
	case 0:
		return p.allChainsTab.drawTabHead(gtx)
	case 1:
		return p.allChainsTab.drawTabHead(gtx)
	}
	return Dim{}
}

func (p *page) drawTabBody(gtx Gtx, index int) Dim {
	switch index {
	case 0:
		return p.allChainsTab.drawTabBody(gtx)
	case 1:
		return p.allChainsTab.drawTabBody(gtx)

	}
	return Dim{}
}

func (p *page) OnDatabaseChange(event pubsub.Event) {
	switch event.Data.(type) {

	}
}
func (p *page) URL() URL {
	return fwk.WalletPageURL
}
