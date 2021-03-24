package view

import (
	"gioui.org/app"
	"gioui.org/font/gofont"
	"gioui.org/io/key"
	"gioui.org/io/system"
	"gioui.org/layout"
	"gioui.org/op"
	"gioui.org/widget"
	"gioui.org/widget/material"
	"gioui.org/x/component"
	log "github.com/sirupsen/logrus"
	"golang.org/x/exp/shiny/materialdesign/colornames"
	"golang.org/x/exp/shiny/materialdesign/icons"
	"image/color"
	"protonet.live/jni"
	"protonet.live/service"
	"sync"
)

type (
	C      = layout.Context
	D      = layout.Dimensions
	E      = system.FrameEvent
	Widget = layout.Widget
)

type View interface {
	Layout(C) D
	SetBarActions()
}

// IsStageRunning false indicates app may be running in background
var IsStageRunning = false

type Navigator struct {
	*service.ChatService
	*component.ModalLayer
	*component.ModalNavDrawer
	*component.AppBar
	material.Theme
	NavDrawer         component.NavDrawer
	CurrentView       View
	CurrentViewMutex  sync.Mutex
	NavItems          []component.NavItem
	History           []View
	LastClipboardText string
}

//var appTheme = material.NewTheme(gofont.Collection())
var Window *app.Window

func (nav *Navigator) GetCurrentView() View {
	nav.CurrentViewMutex.Lock()
	defer nav.CurrentViewMutex.Unlock()
	return nav.CurrentView
}

var Nav = &Navigator{}

func Loop(appWindow *app.Window) (err error) {
	defer Nav.BeforeDestroy()
	var ops op.Ops
	Window = appWindow
	Nav.ChatService = service.NewChatService()
	Nav.ModalLayer = component.NewModal()
	Nav.Theme = *material.NewTheme(gofont.Collection())
	Nav.Theme.ContrastBg = color.NRGBA{R: 0, G: 0, B: 0, A: 255}
	Nav.ModalNavDrawer = component.ModalNavFrom(&Nav.NavDrawer, Nav.ModalLayer)
	Nav.AppBar = component.NewAppBar(Nav.ModalLayer)
	Nav.CreateNavDrawer()
	Nav.SetCurrentView(Nav.NavItems[0].Tag.(View))
	Nav.History = append(Nav.History, Nav.GetCurrentView())
	Nav.CurrentView.SetBarActions()
	for {
		select {
		case e := <-appWindow.Events():
			switch e := e.(type) {
			case system.DestroyEvent:
				return e.Err
			case *system.CommandEvent:
				switch e.Type {
				case system.CommandBack:
					if len(Nav.History) > 1 {
						Nav.PopView()
						e.Cancel = true
					}
				}
			case system.FrameEvent:
				insets := e.Insets
				e.Insets = system.Insets{}
				gtx := layout.NewContext(&ops, e)
				e.Insets = insets
				if Nav.NavDrawer.NavDestinationChanged() {
					Nav.SetCurrentView(Nav.NavDrawer.CurrentNavDestination().(View))
					Nav.History[len(Nav.History)-1] = Nav.GetCurrentView()
					Nav.GetCurrentView().SetBarActions()

				}
				//if !Nav.ChatService.IsServiceReady() {
				//	layout.UniformInset(unit.Dp(20)).Layout(gtx, func(gtx C) D {
				//		Nav.Bg.A = 255
				//		Nav.Bg.R = 255
				//		return material.Loader(&Nav.Theme).Layout(gtx)
				//	})
				//} else {
				//	Nav.DrawPage(gtx, e, Nav.GetCurrentView())
				//}
				Nav.DrawPage(gtx, e, Nav.GetCurrentView())
				//if !Nav.ChatService.IsServiceReady() {
				//	component.Rect{Color: color.NRGBA{A: 150},
				//		Size: image.Point{X: gtx.Constraints.Max.X,
				//			Y: gtx.Constraints.Max.Y},
				//		Radii: 0,
				//	}.Layout(gtx)
				//	gtx.Constraints.Min.Y = gtx.Constraints.Max.Y
				//	gtx.Constraints.Min.X = gtx.Constraints.Max.X
				//	layout.Flex{
				//		Axis:      layout.Vertical,
				//		Spacing:   layout.SpaceSides,
				//		Alignment: layout.Middle,
				//		WeightSum: 0,
				//	}.Layout(gtx,
				//		layout.Flexed(1, func(gtx C) D {
				//			return layout.Flex{
				//				Spacing:   layout.SpaceSides,
				//				Alignment: layout.Middle,
				//			}.Layout(gtx,
				//				layout.Rigid(func(gtx C) D {
				//					gtx.Constraints.Max.Y = 100
				//					gtx.Constraints.Max.X = 100
				//					gtx.Constraints.Min.X = 100
				//					gtx.Constraints.Min.Y = 100
				//					return material.Loader(&Nav.Theme).Layout(gtx)
				//				}),
				//			)
				//		},
				//		),
				//	)
				//}
				e.Frame(gtx.Ops)
			case key.Event:
				if e.Name == "⏎" && e.Modifiers == 0x4 && e.State == 0x0 {
					log.Printf("Shift and Enter key pressed %#v\n", e)
				}
			case system.StageEvent:
				log.Printf("Resultis %v\nResultis %#v\n", e.Stage, e.Stage)
				if e.Stage == system.StagePaused {
					IsStageRunning = false
				} else if e.Stage == system.StageRunning {
					IsStageRunning = true
				}
			}
		case <-Nav.ChatService.GetChangesNotifier():
			appWindow.Invalidate()
		case txtMsg := <-Nav.ChatService.GetShowNotification():
			if !IsStageRunning {
				jni.ShowNotification(txtMsg.CreatorID, txtMsg.Message)
			}
		}
	}
}

func (nav *Navigator) DrawPage(gtx C, e system.FrameEvent, v View) {
	layout.Stack{}.Layout(gtx,
		// fill the entire screen including status bar and bottom soft Navigator bar
		layout.Expanded(func(gtx C) D {
			return component.Rect{
				Color: nav.Theme.ContrastBg,
				Size:  gtx.Constraints.Max,
			}.Layout(gtx)
		}),
		//
		layout.Stacked(func(gtx C) D {
			return layout.Inset{
				Bottom: e.Insets.Bottom,
				Left:   e.Insets.Left,
				Right:  e.Insets.Right,
				Top:    e.Insets.Top,
			}.Layout(gtx, func(gtx C) D {
				return layout.Stack{}.Layout(gtx,
					layout.Expanded(func(gtx C) D {
						return component.Rect{
							Color: color.NRGBA(colornames.White),
							Size:  gtx.Constraints.Max,
						}.Layout(gtx)
					}),
					layout.Stacked(func(gtx C) D {
						d := v.Layout(gtx)
						d = nav.ModalLayer.Layout(gtx, &nav.Theme)
						return d
					}),
				)
			})
		}),
	)
}

func (nav *Navigator) PushView(view View) {
	log.Println("Nav PushView called..........")
	nav.History = append(nav.History, view)
	nav.SetCurrentView(view)
	nav.GetCurrentView().SetBarActions()
	InvalidateWindows()
}
func (nav *Navigator) PopView() {
	if len(nav.History) > 1 {
		nav.History = Nav.History[0 : len(Nav.History)-1]
		nav.SetCurrentView(Nav.History[len(Nav.History)-1])
		nav.GetCurrentView().SetBarActions()
		//InvalidateWindows()
	}
}

func (nav *Navigator) CreateNavDrawer() {
	var name string
	//if ChatService.IsRunning && ChatService.AccountsService.user != nil {
	//	name =  ChatService.AccountsService.user.Name
	//}
	var homeIcon, _ = widget.NewIcon(icons.ActionHome)
	var accountsIcon, _ = widget.NewIcon(icons.ActionAccountCircle)
	nav.NavDrawer = component.NewNav("ProtoNet", name)
	nav.NavItems = []component.NavItem{
		{
			Name: "Contacts",
			Icon: homeIcon,
			Tag:  NewContactsView(nav),
		},
		{
			Name: "Accounts",
			Icon: accountsIcon,
			Tag:  NewAccountsView(nav),
		},
	}
	for _, n := range nav.NavItems {
		nav.NavDrawer.AddNavItem(n)
	}
}

func (nav *Navigator) SetCurrentView(view View) {
	nav.CurrentViewMutex.Lock()
	defer nav.CurrentViewMutex.Unlock()
	nav.CurrentView = view
}

func (nav *Navigator) BeforeDestroy() {
	log.Println("beforeDestroy is called")
}
func InvalidateWindows() {
	if Window != nil {
		Window.Invalidate()
	}
}
