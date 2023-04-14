// Package fwk stands for framework
package fwk

import (
	"gioui.org/app"
	"gioui.org/io/system"
	"gioui.org/layout"
	"gioui.org/widget"
	"gioui.org/widget/material"
	"gioui.org/x/component"
	"gioui.org/x/notify"
	"github.com/mearaj/protonet/internal/pubsub"
	"image/color"
)

type Manager interface {
	NavigateToPage(page Page, AfterNavCallback func())
	NavigateToURL(pageURL URL, AfterNavCallback func())
	PopUp()
	CurrentPage() Page
	GetWindowWidthInDp() int
	GetWindowWidthInPx() int
	GetWindowHeightInDp() int
	GetWindowHeightInPx() int
	IsStageRunning() bool
	Theme() *material.Theme
	Window() *app.Window
	Notifier() notify.Notifier
	Modal() Modal
	PageFromURL(url URL) Page
	SystemInsets() system.Insets
	ShouldDrawSidebar() bool
	Snackbar() Snackbar
}

type Modal interface {
	Show(widget layout.Widget, onBackdropClickCallback func(), animation Animation)
	Dismiss(afterDismiss func())
	View
}
type Snackbar interface {
	Show(txt string, actionButton *widget.Clickable, actionColor color.NRGBA, actionText string)
	View
}

type ViewWidget interface {
	Layout(gtx Gtx, widget layout.Widget) Dim
}

type View interface {
	Layout(gtx Gtx) Dim
}

type Page interface {
	View
	URL() URL
}

// PagePostPopUp is a page which is active after previous page is popped up
type PagePostPopUp interface {
	Page
	OnPopUpPreviousPage()
}

type DatabaseListener interface {
	OnDatabaseChange(event pubsub.Event)
}

type ViewWithDBListener interface {
	View
	DatabaseListener
}

type URL string

const (
	SettingsPageURL      URL = "/settings"
	AccountsPageURL          = SettingsPageURL + "/accounts"
	ContactsPageURL          = SettingsPageURL + "/contacts"
	ThemePageURL             = SettingsPageURL + "/theme"
	NotificationsPageURL     = SettingsPageURL + "/notifications"
	HelpPageURL              = SettingsPageURL + "/help"
	AboutPageURL             = SettingsPageURL + "/about"
	ChatPageURL          URL = "/chat"
	ChatRoomPageURL      URL = "/chat-room"
	WalletPageURL        URL = "/wallet"
)

type (
	Gtx       = layout.Context
	Dim       = layout.Dimensions
	Animation = component.VisibilityAnimation
)
