package chatroom

import (
	"bytes"
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
	"github.com/mearaj/protonet/service"
	. "github.com/mearaj/protonet/ui/fwk"
	"github.com/mearaj/protonet/ui/view"
	"golang.org/x/exp/shiny/materialdesign/colornames"
	"golang.org/x/exp/shiny/materialdesign/icons"
	"image"
	"image/color"
	"runtime"
	"strings"
	"time"
)

var defaultListSize = 50
var animation = component.VisibilityAnimation{
	Duration: time.Millisecond * 250,
	State:    component.Invisible,
	Started:  time.Time{},
}

type page struct {
	layout.List
	Manager
	Theme               *material.Theme
	iconSendMessage     *widget.Icon
	inputMsgField       component.TextField
	buttonNavigation    widget.Clickable
	submitButton        widget.Clickable
	btnIconsStack       widget.Clickable
	btnIconExpand       widget.Clickable
	btnIconCollapse     widget.Clickable
	btnVoiceMessage     widget.Clickable
	btnAudioCall        widget.Clickable
	btnVideoCall        widget.Clickable
	iconMenu            *widget.Icon
	iconNav             *widget.Icon
	iconExpand          *widget.Icon
	iconCollapse        *widget.Icon
	iconVoiceMessage    *widget.Icon
	iconAudioCall       *widget.Icon
	iconVideoCall       *widget.Icon
	contact             service.Contact
	menuAnimation       component.VisibilityAnimation
	iconsStackAnimation component.VisibilityAnimation
	view.AvatarView
	totalMessages           []service.Message
	fetchingMessagesCh      chan []service.Message
	fetchingMessagesCountCh chan int64
	isFetchingMessages      bool
	isFetchingMessagesCount bool
	isMarkingMessagesAsRead bool
	// time in milliseconds
	lastDateTimeShown        int64
	userLastTouchedAnimation Animation
	listPosition             layout.Position
	messagesCount            int64
	initialized              bool
}

func New(manager Manager, contact service.Contact) Page {
	navIcon, _ := widget.NewIcon(icons.NavigationArrowBack)
	iconSendMessage, _ := widget.NewIcon(icons.ContentSend)
	iconMenu, _ := widget.NewIcon(icons.NavigationMenu)
	iconExpand, _ := widget.NewIcon(icons.NavigationUnfoldMore)
	iconCollapse, _ := widget.NewIcon(icons.NavigationUnfoldLess)
	iconVoiceMessage, _ := widget.NewIcon(icons.AVMic)
	iconAudioCall, _ := widget.NewIcon(icons.CommunicationPhone)
	iconVideoCall, _ := widget.NewIcon(icons.AVVideoCall)
	submitEnabled := runtime.GOOS != "android" && runtime.GOOS != "ios"
	pg := page{
		Manager:                 manager,
		Theme:                   manager.Theme(),
		iconNav:                 navIcon,
		iconMenu:                iconMenu,
		iconSendMessage:         iconSendMessage,
		contact:                 contact,
		iconExpand:              iconExpand,
		iconCollapse:            iconCollapse,
		iconVoiceMessage:        iconVoiceMessage,
		iconAudioCall:           iconAudioCall,
		iconVideoCall:           iconVideoCall,
		fetchingMessagesCh:      make(chan []service.Message, 10),
		fetchingMessagesCountCh: make(chan int64, 10),
		totalMessages:           make([]service.Message, 0),
		List: layout.List{
			Axis:        layout.Vertical,
			ScrollToEnd: true,
			Position:    layout.Position{},
		},
		inputMsgField: component.TextField{
			Editor: widget.Editor{Submit: submitEnabled},
		},
		menuAnimation:            animation,
		iconsStackAnimation:      animation,
		userLastTouchedAnimation: animation,
	}
	return &pg
}

func (p *page) Layout(gtx Gtx) Dim {
	if !p.initialized {
		p.fetchMessages(0, defaultListSize)
		p.fetchMessagesCount()
		p.initialized = true
	}
	p.fetchMessagesOnScroll()
	p.markPreviousMessagesAsRead()

	if msgs, ok := view.Listen[[]service.Message](p.fetchingMessagesCh); ok {
		for i, j := 0, len(msgs)-1; i < j; i, j = i+1, j-1 {
			msgs[i], msgs[j] = msgs[j], msgs[i]
		}
		p.totalMessages = msgs
		p.isFetchingMessages = false
	}
	if msgsCount, ok := view.Listen[int64](p.fetchingMessagesCountCh); ok {
		if msgsCount != p.messagesCount {
			p.messagesCount = msgsCount
			if !p.isFetchingMessages {
				p.fetchMessages(0, len(p.totalMessages))
			}
		}
		p.isFetchingMessagesCount = false
	}

	now := time.Now().UnixMilli()
	if now-p.lastDateTimeShown > 3000 {
		p.userLastTouchedAnimation.Disappear(gtx.Now)
	}
	if p.listPosition.First != p.Position.First {
		p.userLastTouchedAnimation.Appear(gtx.Now)
		p.lastDateTimeShown = time.Now().UnixMilli()
	}

	if p.Theme == nil {
		p.Theme = p.Manager.Theme()
	}
	flex := layout.Flex{Axis: layout.Vertical, Spacing: layout.SpaceBetween}
	d := flex.Layout(gtx,
		layout.Rigid(p.DrawAppBar),
		layout.Flexed(1, p.drawChatRoomList),
		layout.Rigid(p.drawSendMsgField),
	)
	p.drawIconsStack(gtx)
	p.drawMenuLayout(gtx)
	for _, e := range gtx.Queue.Events(p) {
		switch e := e.(type) {
		case pointer.Event:
			switch e.Type {
			case pointer.Press:
				if !p.btnIconsStack.Pressed() {
					p.iconsStackAnimation.Disappear(gtx.Now)
				}
			}
			if !p.btnIconsStack.Pressed() {
				p.userLastTouchedAnimation.Appear(gtx.Now)
				p.lastDateTimeShown = time.Now().UnixMilli()
			}
		}
	}
	return d
}

func (p *page) DrawAppBar(gtx Gtx) Dim {
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
						navigationIcon := p.iconNav
						button := material.IconButton(th, &p.buttonNavigation, navigationIcon, "Nav Icon Button")
						button.Size = unit.Dp(40)
						button.Background = th.Palette.ContrastBg
						button.Color = th.Palette.ContrastFg
						button.Inset = layout.UniformInset(unit.Dp(8))
						return button.Layout(gtx)
					}),
					layout.Rigid(func(gtx Gtx) Dim {
						gtx.Constraints.Max.X = gtx.Constraints.Max.X - gtx.Dp(56)
						return layout.Inset{Left: unit.Dp(8)}.Layout(gtx, func(gtx Gtx) Dim {
							titleText := p.contact.PublicKey
							title := material.Label(th, unit.Sp(18), titleText)
							title.Color = th.Palette.ContrastFg
							return component.TruncatingLabelStyle(title).Layout(gtx)
						})
					}),
				)
			}),
			layout.Rigid(func(gtx Gtx) Dim {
				if p.AvatarView.Size == (image.Point{}) {
					p.AvatarView.Size = image.Point{X: gtx.Dp(36), Y: gtx.Dp(36)}
				}
				if p.AvatarView.Image == nil {
					img, _, _ := image.Decode(bytes.NewReader(p.contact.Avatar))
					p.AvatarView.Image = img
				}
				if p.AvatarView.Theme == nil {
					p.AvatarView.Theme = p.Theme
				}
				return p.AvatarView.Layout(gtx)
			}),
		)
	})
}
func (p *page) drawChatRoomList(gtx Gtx) Dim {
	gtx.Constraints.Min = gtx.Constraints.Max
	//if strings.TrimSpace(p.Service().Account().PrivateKey) == "" {
	//	return p.AuthAccount.Layout(gtx)
	//}
	areaStack := clip.Rect(image.Rectangle{Max: gtx.Constraints.Max}).Push(gtx.Ops)
	d := layout.Flex{}.Layout(gtx,
		layout.Rigid(func(gtx Gtx) Dim {
			inset := layout.Inset{Right: unit.Dp(8), Left: unit.Dp(8)}
			return inset.Layout(gtx, func(gtx Gtx) Dim {
				return p.List.Layout(gtx, len(p.totalMessages), p.drawChatRoomListItem)
			})
		}))
	layout.Stack{}.Layout(gtx, layout.Stacked(func(gtx layout.Context) layout.Dimensions {
		yConstraints := gtx.Dp(48)
		gtx.Constraints.Min.Y, gtx.Constraints.Max.Y = yConstraints, yConstraints
		gtx.Constraints.Min.X = gtx.Constraints.Max.X
		if p.isFetchingMessages {
			loader := view.Loader{
				Size: image.Point{Y: yConstraints / 2, X: yConstraints / 2},
			}
			loader.Layout(gtx)
		} else if !p.isFetchingMessages {
			if len(p.totalMessages) > 0 && p.List.Position.First < len(p.totalMessages) {
				progress := p.userLastTouchedAnimation.Revealed(gtx)
				timeVal, _ := time.Parse(time.RFC3339, p.totalMessages[p.List.Position.First].Created)
				timeVal = timeVal.Local()
				txtMsg := timeVal.Format("Mon Jan 2 2006")
				label := material.Label(p.Theme, p.Theme.TextSize*0.9, txtMsg)
				label.Color = p.Theme.ContrastFg
				label.Font.Style = text.Italic
				label.Font.Weight = text.SemiBold
				label.Color.A = uint8(float32(label.Color.A) * progress)
				layout.Center.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
					inset := layout.Inset{Top: unit.Dp(8), Bottom: unit.Dp(8), Right: unit.Dp(12), Left: unit.Dp(12)}
					bgColor := p.Theme.ContrastBg
					bgColor.A = uint8(float32(label.Color.A) * progress)
					mac := op.Record(gtx.Ops)
					d := inset.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
						return label.Layout(gtx)
					})
					stop := mac.Stop()
					component.Rect{
						Color: bgColor,
						Size:  d.Size,
						Radii: gtx.Dp(16),
					}.Layout(gtx)
					stop.Add(gtx.Ops)
					return d
				})
			}
		}
		return Dim{Size: image.Pt(gtx.Constraints.Max.X, yConstraints)}
	}))
	areaStack.Pop()
	return d
}
func (p *page) drawChatRoomListItem(gtx Gtx, index int) Dim {
	item := PageItem{
		Message: p.totalMessages[index],
		Theme:   p.Theme,
	}
	return item.Layout(gtx)
}

func (p *page) inputMsgFieldSubmitted() (submit bool) {
	for _, event := range p.inputMsgField.Events() {
		if _, submit = event.(widget.SubmitEvent); submit {
			break
		}
	}
	return submit
}

func (p *page) drawSendMsgField(gtx Gtx) Dim {
	if p.submitButton.Clicked() || p.inputMsgFieldSubmitted() {
		msg := strings.TrimSpace(p.inputMsgField.Text())
		canSend := msg != ""
		if canSend {
			p.inputMsgField.Clear()
			created := time.Now().UTC().Format(time.RFC3339)
			go func(addr string, msg string, created string) {
				err := <-p.Service().SendMessage(addr, msg, created)
				if err != nil {
					alog.Logger().Errorln(err)
				} else {
					alog.Logger().Infoln("successfully sent msg...")
				}
			}(p.contact.PublicKey, msg, created)
		}
	}
	fl := layout.Flex{
		Axis:      layout.Horizontal,
		Spacing:   layout.SpaceBetween,
		Alignment: layout.End,
		WeightSum: 1,
	}

	inset := layout.UniformInset(unit.Dp(8))
	return inset.Layout(gtx, func(gtx Gtx) Dim {
		gtx.Constraints.Max.Y = gtx.Dp(120)
		return fl.Layout(gtx,
			layout.Flexed(1.0, func(gtx Gtx) Dim {
				return p.inputMsgField.Layout(gtx, p.Theme, "Message")
			}),
			layout.Rigid(func(gtx Gtx) Dim {
				inset := layout.Inset{Left: unit.Dp(8.0)}
				return inset.Layout(
					gtx,
					func(gtx Gtx) Dim {
						return material.IconButtonStyle{
							Background: p.Theme.ContrastBg,
							Color:      p.Theme.ContrastFg,
							Icon:       p.iconSendMessage,
							Size:       unit.Dp(24.0),
							Button:     &p.submitButton,
							Inset:      layout.UniformInset(unit.Dp(9)),
						}.Layout(gtx)
					},
				)
			}),
			layout.Rigid(func(gtx Gtx) Dim {
				inset := layout.Inset{Left: unit.Dp(8.0)}
				return inset.Layout(
					gtx,
					func(gtx Gtx) Dim {
						btn := &p.btnIconExpand
						icon := p.iconExpand
						if p.btnIconCollapse.Clicked() {
							p.iconsStackAnimation.Disappear(gtx.Now)
						}
						if p.btnIconExpand.Clicked() {
							p.iconsStackAnimation.Appear(gtx.Now)
						}
						if p.iconsStackAnimation.Revealed(gtx) != 0 {
							btn = &p.btnIconCollapse
							icon = p.iconCollapse
						}

						return material.IconButtonStyle{
							Background: p.Theme.ContrastBg,
							Color:      p.Theme.ContrastFg,
							Icon:       icon,
							Size:       unit.Dp(24.0),
							Button:     btn,
							Inset:      layout.UniformInset(unit.Dp(9)),
						}.Layout(gtx)
					},
				)
			}),
		)
	})
}

func (p *page) drawMenuLayout(gtx Gtx) Dim {

	layout.Stack{Alignment: layout.NE}.Layout(gtx,
		layout.Stacked(func(gtx Gtx) Dim {
			progress := p.menuAnimation.Revealed(gtx)
			gtx.Constraints.Max.X = int(float32(gtx.Constraints.Max.X) * progress)
			gtx.Constraints.Max.Y = int(float32(gtx.Constraints.Max.Y) * progress)
			return component.Rect{Size: gtx.Constraints.Max, Color: color.NRGBA{A: 200}}.Layout(gtx)
		}),
		layout.Stacked(func(gtx Gtx) Dim {
			progress := p.menuAnimation.Revealed(gtx)
			macro := op.Record(gtx.Ops)
			gtx.Constraints.Max.X = int(float32(gtx.Dp(300)) * progress)
			gtx.Constraints.Min.X = gtx.Constraints.Max.X
			d := p.drawMenuItems(gtx)
			call := macro.Stop()
			d.Size.Y = int(float32(d.Size.Y) * progress)
			component.Rect{Size: d.Size, Color: color.NRGBA(colornames.White)}.Layout(gtx)
			call.Add(gtx.Ops)
			return d
		}),
	)
	return Dim{}
}

func (p *page) drawIconsStack(gtx Gtx) Dim {
	layout.Stack{Alignment: layout.SE}.Layout(gtx,
		layout.Stacked(func(gtx Gtx) Dim {
			offset := image.Pt(-gtx.Dp(8), -gtx.Dp(64))
			op.Offset(offset).Add(gtx.Ops)
			progress := p.iconsStackAnimation.Revealed(gtx)
			macro := op.Record(gtx.Ops)
			d := p.btnIconsStack.Layout(gtx, p.drawIconsStackItems)
			call := macro.Stop()
			d.Size.Y = int(float32(d.Size.Y) * progress)
			component.Rect{Size: d.Size, Color: color.NRGBA{}}.Layout(gtx)
			clipOp := clip.Rect{Max: d.Size}.Push(gtx.Ops)
			call.Add(gtx.Ops)
			clipOp.Pop()
			return d
		}),
	)
	return Dim{}
}

func (p *page) drawIconsStackItems(gtx Gtx) Dim {
	//inset := layout.UniformInset(unit.Dp(12))
	flex := layout.Flex{Axis: layout.Vertical}
	return flex.Layout(gtx,
		layout.Rigid(func(gtx Gtx) Dim {
			inset := layout.Inset{Left: unit.Dp(8.0)}
			return inset.Layout(
				gtx,
				func(gtx Gtx) Dim {
					return material.IconButtonStyle{
						Background: p.Theme.ContrastBg,
						Color:      p.Theme.ContrastFg,
						Icon:       p.iconVideoCall,
						Size:       unit.Dp(24.0),
						Button:     &p.btnVideoCall,
						Inset:      layout.UniformInset(unit.Dp(9)),
					}.Layout(gtx)
				},
			)
		}),
		layout.Rigid(layout.Spacer{Height: unit.Dp(8)}.Layout),
		layout.Rigid(func(gtx Gtx) Dim {
			inset := layout.Inset{Left: unit.Dp(8.0)}
			return inset.Layout(
				gtx,
				func(gtx Gtx) Dim {
					return material.IconButtonStyle{
						Background: p.Theme.ContrastBg,
						Color:      p.Theme.ContrastFg,
						Icon:       p.iconAudioCall,
						Size:       unit.Dp(24.0),
						Button:     &p.btnAudioCall,
						Inset:      layout.UniformInset(unit.Dp(9)),
					}.Layout(gtx)
				},
			)
		}),
		layout.Rigid(layout.Spacer{Height: unit.Dp(8)}.Layout),
		layout.Rigid(func(gtx Gtx) Dim {
			inset := layout.Inset{Left: unit.Dp(8.0)}
			return inset.Layout(
				gtx,
				func(gtx Gtx) Dim {
					return material.IconButtonStyle{
						Background: p.Theme.ContrastBg,
						Color:      p.Theme.ContrastFg,
						Icon:       p.iconVoiceMessage,
						Size:       unit.Dp(24.0),
						Button:     &p.btnVoiceMessage,
						Inset:      layout.UniformInset(unit.Dp(9)),
					}.Layout(gtx)
				},
			)
		}),
	)
}

func (p *page) fetchMessagesOnScroll() {
	p.listPosition = p.Position
	shouldFetch := p.Position.First == 0 && !p.isFetchingMessages && int64(len(p.totalMessages)) < p.messagesCount
	if shouldFetch {
		currentSize := len(p.totalMessages) + defaultListSize
		p.fetchMessages(0, currentSize)
	}
}

func (p *page) markPreviousMessagesAsRead() {
	if !p.isMarkingMessagesAsRead {
		p.isMarkingMessagesAsRead = true
		go func() {
			<-p.Service().MarkPrevMessagesAsRead(p.contact.PublicKey)
			p.isMarkingMessagesAsRead = false
		}()
	}
}

func (p *page) drawMenuItems(gtx Gtx) Dim {
	//inset := layout.UniformInset(unit.Dp(12))
	flex := layout.Flex{Axis: layout.Vertical, Spacing: layout.SpaceEnd, Alignment: layout.Start, WeightSum: 1}
	return flex.Layout(gtx,
		layout.Rigid(func(gtx Gtx) Dim {
			return Dim{}
		}),
	)
}

func (p *page) OnDatabaseChange(event service.Event) {
	switch e := event.Data.(type) {
	case service.MessagesCountChangedEventData:
		if e.AccountPublicKey == p.Service().Account().PublicKey &&
			e.ContactPublicKey == p.contact.PublicKey {
			p.fetchMessages(0, len(p.totalMessages))
			p.fetchMessagesCount()
		}
	case service.MessagesStateChangedEventData:
		if e.AccountPublicKey == p.Service().Account().PublicKey &&
			e.ContactPublicKey == p.contact.PublicKey {
			p.fetchMessages(0, len(p.totalMessages))
			p.fetchMessagesCount()
		}
	}
}

func (p *page) fetchMessages(offset, limit int) {
	if !p.isFetchingMessages {
		p.isFetchingMessages = true
		go func(contactAddr string, offset int, limit int) {
			p.fetchingMessagesCh <- <-p.Service().Messages(contactAddr, offset, limit)
			p.Window().Invalidate()
		}(p.contact.PublicKey, offset, limit)
	}
}

func (p *page) fetchMessagesCount() {
	if !p.isFetchingMessagesCount {
		p.isFetchingMessagesCount = true
		go func() {
			p.fetchingMessagesCountCh <- <-p.Service().MessagesCount(p.contact.PublicKey)
			p.Window().Invalidate()
		}()
	}
}

func (p *page) URL() URL {
	return ChatRoomPageURL + "/" + URL(p.contact.PublicKey)
}
