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
	"github.com/mearaj/audio"
	"github.com/mearaj/protonet/alog"
	chat2 "github.com/mearaj/protonet/internal/chat"
	"github.com/mearaj/protonet/internal/pubsub"
	"github.com/mearaj/protonet/internal/wallet"
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
	Theme                    *material.Theme
	iconSendMessage          *widget.Icon
	inputMsgField            component.TextField
	buttonNavigation         widget.Clickable
	submitButton             widget.Clickable
	btnIconsStack            widget.Clickable
	btnIconExpand            widget.Clickable
	btnIconCollapse          widget.Clickable
	btnVoiceMessage          widget.Clickable
	btnAudioCall             widget.Clickable
	btnVideoCall             widget.Clickable
	iconMenu                 *widget.Icon
	iconNav                  *widget.Icon
	iconExpand               *widget.Icon
	iconCollapse             *widget.Icon
	iconVoiceMessage         *widget.Icon
	iconAudioCall            *widget.Icon
	iconVideoCall            *widget.Icon
	contact                  chat2.Contact
	menuAnimation            component.VisibilityAnimation
	iconsStackAnimation      component.VisibilityAnimation
	AvatarView               view.AvatarView
	pageItems                []*PageItem
	fetchingMessagesCh       chan []chat2.Message
	isFetchingMessages       bool
	isFetchingMessagesCount  bool
	lastDateTimeShown        int64
	userLastTouchedAnimation Animation
	listPosition             layout.Position
	messagesCount            int64
	initialized              bool
	recorder                 *audio.RawRecorder
}

func New(manager Manager, contact chat2.Contact) Page {
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
		Manager:            manager,
		Theme:              manager.Theme(),
		iconNav:            navIcon,
		iconMenu:           iconMenu,
		iconSendMessage:    iconSendMessage,
		contact:            contact,
		iconExpand:         iconExpand,
		iconCollapse:       iconCollapse,
		iconVoiceMessage:   iconVoiceMessage,
		iconAudioCall:      iconAudioCall,
		iconVideoCall:      iconVideoCall,
		fetchingMessagesCh: make(chan []chat2.Message, 10),
		pageItems:          make([]*PageItem, 0),
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
		if p.Theme == nil {
			p.Theme = p.Manager.Theme()
		}
		p.fetchMessages(0, defaultListSize)
		p.fetchMessagesCount()
		p.initialized = true
	}
	p.markPreviousMessagesAsRead()
	p.fetchMessagesOnScroll()

	now := time.Now().UnixMilli()
	if now-p.lastDateTimeShown > 3000 {
		p.userLastTouchedAnimation.Disappear(gtx.Now)
	}
	if p.listPosition.First != p.Position.First {
		p.userLastTouchedAnimation.Appear(gtx.Now)
		p.lastDateTimeShown = time.Now().UnixMilli()
	}

	flex := layout.Flex{Axis: layout.Vertical, Spacing: layout.SpaceBetween}
	d := flex.Layout(gtx,
		layout.Rigid(p.DrawAppBar),
		layout.Flexed(1, p.drawChatRoomList),
		layout.Rigid(p.drawSendMsgField),
	)
	p.drawIconsStack(gtx)
	p.drawMenuLayout(gtx)
	p.handleEvents(gtx)
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
	//if strings.TrimSpace(p.Wallet().Account().PrivateKey) == "" {
	//	return p.AuthAccount.Layout(gtx)
	//}
	areaStack := clip.Rect(image.Rectangle{Max: gtx.Constraints.Max}).Push(gtx.Ops)
	d := layout.Flex{}.Layout(gtx,
		layout.Rigid(func(gtx Gtx) Dim {
			inset := layout.Inset{Right: unit.Dp(8), Left: unit.Dp(8)}
			return inset.Layout(gtx, func(gtx Gtx) Dim {
				return p.List.Layout(gtx, len(p.pageItems), p.drawChatRoomListItem)
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
			if len(p.pageItems) > 0 && p.List.Position.First < len(p.pageItems) {
				progress := p.userLastTouchedAnimation.Revealed(gtx)
				timeVal := p.pageItems[p.List.Position.First].CreatedAt
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
	return p.pageItems[(len(p.pageItems))-1-index].Layout(gtx)
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
			msg := chat2.Message{
				Recipient: p.contact.PublicKey,
				CreatedAt: time.Now().UTC(),
				Text:      msg,
			}
			acc, _ := wallet.GlobalWallet.Account()
			chat2.GlobalChat.SendNewMessage(&acc, &msg)
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
					p.handleRecordingClick(gtx)
					bg := p.Theme.ContrastBg
					isRecording := p.recorder != nil &&
						p.recorder.State() == audio.RawRecorderStateRecording
					if isRecording {
						bg = color.NRGBA(colornames.Red500)
					}
					return material.IconButtonStyle{
						Background: bg,
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
	shouldFetch := p.Position.First == 0 && !p.isFetchingMessages && int64(len(p.pageItems)) < p.messagesCount
	if shouldFetch {
		currentSize := len(p.pageItems) + defaultListSize
		p.fetchMessages(0, currentSize)
	}
}

func (p *page) markPreviousMessagesAsRead() {
	acc, err := wallet.GlobalWallet.Account()
	if err == nil {
		p.messagesCount, err = wallet.GlobalWallet.MarkPrevMessagesAsRead(acc.PublicKey, p.contact.PublicKey)
	}
	if err != nil {
		alog.Logger().Errorln(err)
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

func (p *page) OnDatabaseChange(event pubsub.Event) {
	shouldFetch := false
	acc, _ := wallet.GlobalWallet.Account()
	switch e := event.Data.(type) {
	case pubsub.MessageStateChangedEventData:
		msg := e.Message
		for _, i := range p.pageItems {
			if i.Message.ID == msg.ID {
				i.Message = msg
				p.Window().Invalidate()
				break
			}
		}
	case pubsub.SendNewMessageEventData, pubsub.NewMessageReceivedEventData:
		shouldFetch = true
	case pubsub.MessagesStateChangedEventData:
		if e.AccountPublicKey == acc.PublicKey &&
			e.ContactPublicKey == p.contact.PublicKey {
			shouldFetch = true
		}
	case pubsub.CurrentAccountChangedEventData:
		shouldFetch = true
	}
	if shouldFetch {
		p.isFetchingMessages = false
		p.fetchMessagesCount()
		if len(p.pageItems) == 0 {
			p.fetchMessages(0, defaultListSize)
		} else {
			p.fetchMessages(0, len(p.pageItems)+defaultListSize)
		}
	}
}

func (p *page) syncMessagesPageItems(messages []chat2.Message) {
	acc, _ := wallet.GlobalWallet.Account()
	if len(p.pageItems) < len(messages) {
		for i := len(p.pageItems); i < len(messages); i++ {
			msgItem := &PageItem{
				Message:          messages[i],
				Theme:            p.Theme,
				accountPublicKey: acc.PublicKey,
			}
			p.pageItems = append(p.pageItems, msgItem)
		}
	} else if len(p.pageItems) > len(messages) {
		p.pageItems = p.pageItems[:len(messages)]
	}
	for i := range messages {
		p.pageItems[i].Message = messages[i]
		if len(p.pageItems[i].Message.Audio) != 0 {
			var err error
			// Todo: Need to recheck
			if p.pageItems[i].player != nil {
				p.pageItems[i].player, err = audio.NewRawPlayer(p.pageItems[i].Message.Audio, 0, 0)
				if err != nil {
					alog.Logger().Errorln(err)
				}
			}
		}
	}
}

func (p *page) fetchMessages(offset, limit int) {
	if !p.isFetchingMessages {
		p.isFetchingMessages = true
		acc, _ := wallet.GlobalWallet.Account()
		messages, _ := wallet.GlobalWallet.Messages(acc.PublicKey, p.contact.PublicKey, offset, limit)
		p.syncMessagesPageItems(messages)
		p.isFetchingMessages = false
		p.Window().Invalidate()
	}
}

func (p *page) fetchMessagesCount() {
	acc, _ := wallet.GlobalWallet.Account()
	if !p.isFetchingMessagesCount {
		p.isFetchingMessagesCount = true
		msgsCount, _ := wallet.GlobalWallet.MessagesCount(acc.PublicKey, p.contact.PublicKey)
		if msgsCount != p.messagesCount {
			p.messagesCount = msgsCount
			if !p.isFetchingMessages {
				if len(p.pageItems) != 0 {
					p.fetchMessages(0, len(p.pageItems)+defaultListSize)
				} else {
					p.fetchMessages(0, defaultListSize)
				}
			}
			p.Window().Invalidate()
		}
		p.isFetchingMessagesCount = false
	}
}

func (p *page) handleEvents(gtx Gtx) {
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
}

func (p *page) handleRecordingClick(gtx Gtx) {
	if p.btnVoiceMessage.Clicked() {
		go func() {
			if p.recorder == nil {
				var err error
				p.recorder, err = audio.NewRawRecorder(0, 0)
				if err != nil {
					alog.Logger().Errorln(err)
				} else {
					go p.recorder.Record()
				}
			} else {
				state := p.recorder.State()
				if state == audio.RawRecorderStateRecording {
					_ = p.recorder.Stop()
					if len(p.recorder.Bytes()) > 0 {
						msg := chat2.Message{
							Recipient: p.contact.PublicKey,
							CreatedAt: time.Now().UTC(),
							Audio:     p.recorder.Bytes(),
						}
						acc, _ := wallet.GlobalWallet.Account()
						chat2.GlobalChat.SendNewMessage(&acc, &msg)
					}
					p.recorder = nil
				} else {
					go p.recorder.Record()
				}
			}
		}()
	}
}

func (p *page) URL() URL {
	return ChatRoomPageURL + "/" + URL(p.contact.PublicKey)
}
