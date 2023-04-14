package ui

import (
	"gioui.org/app"
	"gioui.org/io/system"
	"gioui.org/layout"
	"gioui.org/op/clip"
	"gioui.org/op/paint"
	"gioui.org/unit"
	"gioui.org/widget/material"
	"gioui.org/x/component"
	"gioui.org/x/notify"
	"github.com/mearaj/protonet/alog"
	"github.com/mearaj/protonet/assets/fonts"
	. "github.com/mearaj/protonet/ui/fwk"
	"github.com/mearaj/protonet/ui/page/about"
	"github.com/mearaj/protonet/ui/page/accounts"
	"github.com/mearaj/protonet/ui/page/chat"
	"github.com/mearaj/protonet/ui/page/contacts"
	"github.com/mearaj/protonet/ui/page/help"
	"github.com/mearaj/protonet/ui/page/notifications"
	"github.com/mearaj/protonet/ui/page/settings"
	"github.com/mearaj/protonet/ui/page/theme"
	"github.com/mearaj/protonet/ui/page/wallet"
	"github.com/mearaj/protonet/ui/view"
	"image"
	"sync"
)

// AppManager Always call NewAppManager function to create AppManager instance
type AppManager struct {
	window *app.Window
	view.Greetings
	settingsSideBar Page
	theme           *material.Theme
	Constraints     layout.Constraints
	Metric          unit.Metric
	notifier        notify.Notifier
	system.Insets
	// isStageRunning, true value indicates app is running in foreground,
	// false indicates running in background
	isStageRunning   bool
	modal            Modal
	pagesStack       []Page
	pageAnimation    view.Slider
	snackbar         Snackbar
	initialized      bool
	initializedMutex sync.RWMutex
}

func (m *AppManager) Theme() *material.Theme {
	return m.theme
}

func (m *AppManager) SystemInsets() system.Insets {
	return m.Insets
}

func (m *AppManager) Window() *app.Window {
	return m.window
}

func (m *AppManager) Notifier() notify.Notifier {
	return m.notifier
}

func (m *AppManager) Snackbar() Snackbar {
	return m.snackbar
}

func (m *AppManager) Initialized() bool {
	m.initializedMutex.RLock()
	initialized := m.initialized
	m.initializedMutex.RUnlock()
	return initialized
}
func (m *AppManager) setInitialized(initialized bool) {
	m.initializedMutex.Lock()
	m.initialized = initialized
	m.initializedMutex.Unlock()
}

func (m *AppManager) init() {
	m.theme = fonts.NewTheme()
	settingsPage := settings.New(m)
	m.settingsSideBar = settingsPage
	var err error
	m.notifier, err = notify.NewNotifier()
	if err != nil {
		alog.Logger().Errorln(err)
	}
	m.SetModal(view.NewModalStack())
	m.snackbar = view.NewSnackBar(m)
	m.setInitialized(true)
}

func (m *AppManager) Layout(gtx Gtx) Dim {
	d := layout.Flex{Axis: layout.Vertical}.Layout(gtx, layout.Rigid(func(gtx Gtx) Dim {
		return component.Rect{
			Color: m.Theme().ContrastBg,
			Size:  image.Point{X: gtx.Constraints.Max.X, Y: gtx.Dp(m.Insets.Top)},
			Radii: 0,
		}.Layout(gtx)
	}), layout.Flexed(1, func(gtx Gtx) Dim {
		size := image.Point{X: gtx.Constraints.Max.X, Y: gtx.Constraints.Max.Y - gtx.Dp(m.Insets.Bottom)}
		bounds := image.Rectangle{Max: size}
		paint.FillShape(gtx.Ops, m.Theme().Bg, clip.UniformRRect(bounds, 0).Op(gtx.Ops))
		d := m.drawPage(gtx)
		m.Snackbar().Layout(gtx)
		m.Modal().Layout(gtx)
		return d
	}), layout.Rigid(func(gtx layout.Context) layout.Dimensions {
		return component.Rect{
			Color: m.Theme().ContrastBg,
			Size:  image.Point{X: gtx.Constraints.Max.X, Y: gtx.Dp(m.Insets.Bottom)},
			Radii: 0,
		}.Layout(gtx)
	}))
	return d
}

func (m *AppManager) CurrentPage() Page {
	stackSize := len(m.pagesStack)
	if stackSize > 0 {
		return m.pagesStack[stackSize-1]
	}
	m.pagesStack = []Page{settings.New(m)}
	return m.settingsSideBar
}
func (m *AppManager) GetWindowWidthInDp() int {
	width := int(float32(m.Constraints.Max.X) / m.Metric.PxPerDp)
	return width
}

func (m *AppManager) GetWindowWidthInPx() int {
	return m.Constraints.Max.X
}

func (m *AppManager) GetWindowHeightInDp() int {
	width := int(float32(m.Constraints.Max.Y) / m.Metric.PxPerDp)
	return width
}

func (m *AppManager) GetWindowHeightInPx() int {
	return m.Constraints.Max.Y
}

func (m *AppManager) IsStageRunning() bool {
	return m.isStageRunning
}
func (m *AppManager) ShouldDrawSidebar() bool {
	minWidth := 800 // 800 is value in Dp
	winWidth := m.GetWindowWidthInDp()
	return winWidth >= minWidth
}

func (m *AppManager) Modal() Modal {
	return m.modal
}
func (m *AppManager) SetModal(modal Modal) {
	m.modal = modal
}

func (m *AppManager) PageFromURL(url URL) Page {
	switch url {
	case SettingsPageURL:
		return m.settingsSideBar
	case ChatPageURL:
		fallthrough
	default:
		return m.settingsSideBar
	}
}

func (m *AppManager) drawPage(gtx Gtx) Dim {
	maxDim := gtx.Constraints.Max

	d := layout.Flex{}.Layout(gtx,
		layout.Rigid(func(gtx layout.Context) (d Dim) {
			if !m.ShouldDrawSidebar() {
				return d
			}
			gtx.Constraints.Max.X = int(float32(maxDim.X) * 0.40)
			gtx.Constraints.Min = gtx.Constraints.Max
			d = m.settingsSideBar.Layout(gtx)
			return d
		}),
		layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			if m.ShouldDrawSidebar() {
				gtx.Constraints.Max.X = int(float32(maxDim.X) * 0.60)
			}
			maxDim := gtx.Constraints.Max
			gtx.Constraints.Min = maxDim
			areaStack := clip.Rect(image.Rectangle{Max: gtx.Constraints.Max}).Push(gtx.Ops)
			d := m.pageAnimation.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
				if m.CurrentPage().URL() == SettingsPageURL && m.ShouldDrawSidebar() {
					m.Greetings.Layout(gtx)
					return Dim{Size: maxDim}
				}
				return m.CurrentPage().Layout(gtx)
			})
			areaStack.Pop()
			return d
		}),
	)
	return d
}

func (m *AppManager) NavigateToPage(page Page, afterNavCallback func()) {
	pageURL := page.URL()
	if pageURL == m.CurrentPage().URL() {
		m.PopUp()
	}
	switch pageURL {
	case SettingsPageURL:
		m.pagesStack = []Page{m.settingsSideBar}
	default:
		m.pagesStack = append(m.pagesStack, page)
	}
	m.pageAnimation.PushLeft()
}

func (m *AppManager) NavigateToURL(pageURL URL, afterNavCallback func()) {
	if pageURL == m.CurrentPage().URL() {
		m.PopUp()
	}
	var page Page
	switch pageURL {
	case SettingsPageURL:
		m.pagesStack = []Page{m.settingsSideBar}
	case ChatPageURL:
		page = chat.New(m)
	case WalletPageURL:
		page = wallet.New(m)
	case AccountsPageURL:
		page = accounts.New(m)
	case ContactsPageURL:
		page = contacts.New(m)
	case ThemePageURL:
		page = theme.New(m)
	case NotificationsPageURL:
		page = notifications.New(m)
	case HelpPageURL:
		page = help.New(m)
	case AboutPageURL:
		page = about.New(m)
	}
	if page != nil {
		m.pagesStack = append(m.pagesStack, page)
	}
	m.pageAnimation.PushLeft()
	if afterNavCallback != nil {
		afterNavCallback()
	}
}

func (m *AppManager) PopUp() {
	stackLength := len(m.pagesStack)
	if stackLength > 1 {
		m.pagesStack = m.pagesStack[0 : stackLength-1]
		if pageAfterPopPup, ok := m.pagesStack[len(m.pagesStack)-1].(PagePostPopUp); ok {
			pageAfterPopPup.OnPopUpPreviousPage()
		}
	}
	m.pageAnimation.PushRight()
}
