package chat

import (
	"bytes"
	"fmt"
	"gioui.org/layout"
	"gioui.org/op"
	"gioui.org/text"
	"gioui.org/unit"
	"gioui.org/widget"
	"gioui.org/widget/material"
	"gioui.org/x/component"
	"github.com/mearaj/protonet/service"
	. "github.com/mearaj/protonet/ui/fwk"
	"github.com/mearaj/protonet/ui/page/chatroom"
	"github.com/mearaj/protonet/ui/view"
	"image"
	"time"
)

type pageItem struct {
	Manager
	contact service.Contact
	widget.Clickable
	*material.Theme
	view.AvatarView
	fetchingLastMessageCh         chan service.Message
	fetchingUnReadMessagesCountCh chan int64
	fetchingLastMessage           bool
	fetchingUnReadMessagesCount   bool
	lastMessage                   service.Message
	messagesCount                 int64
}

func NewChatPageItem(manager Manager, contact service.Contact) *pageItem {
	img, _, _ := image.Decode(bytes.NewReader(contact.Avatar))
	i := pageItem{
		Manager: manager,
		Theme:   manager.Theme(),
		contact: contact,
		AvatarView: view.AvatarView{
			Theme: manager.Theme(),
			Image: img,
		},
		fetchingLastMessageCh:         make(chan service.Message, 10),
		fetchingUnReadMessagesCountCh: make(chan int64, 10),
	}
	return &i
}

func (pi *pageItem) Layout(gtx Gtx) Dim {
	pi.fetchLastMessage()
	pi.fetchMessagesUnreadCount()
	pi.listenToLastMessage()
	pi.listenToUnreadMessagesCount()
	return pi.layoutContent(gtx)
}

func (pi *pageItem) layoutContent(gtx Gtx) Dim {
	gtx.Constraints.Min.X = gtx.Constraints.Max.X
	btnStyle := material.ButtonLayoutStyle{Background: pi.Theme.ContrastBg, Button: &pi.Clickable}

	if pi.Clickable.Clicked() {
		chatRoomPage := chatroom.New(pi.Manager, pi.contact)
		if pi.CurrentPage().URL() != chatRoomPage.URL() {
			pi.NavigateToUrl(ChatPageURL, func() {
				pi.NavigateToPage(chatRoomPage, nil)
			})
		}
	}
	if pi.Clickable.Hovered() ||
		pi.CurrentPage().URL() == pi.URL() {
		btnStyle.Background.A = 50
	} else {
		btnStyle.Background.A = 10
	}

	d := btnStyle.Layout(gtx, func(gtx Gtx) Dim {
		inset := layout.UniformInset(unit.Dp(16))
		gtx.Constraints.Min.X = gtx.Constraints.Max.X
		d := inset.Layout(gtx, func(gtx Gtx) Dim {
			flex := layout.Flex{Spacing: layout.SpaceBetween, Alignment: layout.Middle}
			d := flex.Layout(gtx,
				layout.Rigid(func(gtx layout.Context) layout.Dimensions {
					flex := layout.Flex{Spacing: layout.SpaceEnd, Alignment: layout.Middle}
					return flex.Layout(gtx,
						layout.Rigid(pi.AvatarView.Layout),
						layout.Rigid(func(gtx Gtx) Dim {
							gtx.Constraints.Max.X = gtx.Constraints.Max.X - gtx.Dp(80)
							flex := layout.Flex{Spacing: layout.SpaceSides, Alignment: layout.Start, Axis: layout.Vertical}
							inset := layout.UniformInset(unit.Dp(16))
							inset.Right = unit.Dp(8)
							d := inset.Layout(gtx, func(gtx Gtx) Dim {
								d := flex.Layout(gtx,
									layout.Rigid(func(gtx Gtx) Dim {
										textSize := unit.Sp(14)
										label := material.Label(pi.Theme, textSize, pi.contact.PublicKey)
										label.Font.Weight = text.Bold
										return component.TruncatingLabelStyle(label).Layout(gtx)
									}),
									layout.Rigid(layout.Spacer{Height: unit.Dp(4)}.Layout),
									layout.Rigid(func(gtx Gtx) Dim {
										if pi.lastMessage.Text == "" {
											return Dim{}
										}
										label := material.Label(pi.Theme, pi.Theme.TextSize*0.9, pi.lastMessage.Text)
										return component.TruncatingLabelStyle(label).Layout(gtx)
									}),
								)
								return d
							})
							return d
						}),
					)
				}),
				layout.Rigid(func(gtx layout.Context) layout.Dimensions {
					flex := layout.Flex{Spacing: layout.SpaceBetween, Alignment: layout.Middle, Axis: layout.Vertical}
					return flex.Layout(gtx,
						layout.Rigid(func(gtx layout.Context) layout.Dimensions {
							if pi.lastMessage.Text == "" {
								return Dim{}
							}
							timeVal, _ := time.Parse(time.RFC3339, pi.lastMessage.Created)
							timeVal = timeVal.Local()
							txt := view.LastSeenTime(timeVal)
							label := material.Label(pi.Theme, pi.Theme.TextSize*0.75, txt)
							label.Color = pi.Theme.ContrastBg
							label.Font.Style = text.Italic
							label.Font.Weight = text.SemiBold
							return label.Layout(gtx)
						}),
						layout.Rigid(layout.Spacer{Height: unit.Dp(4)}.Layout),
						layout.Rigid(func(gtx layout.Context) layout.Dimensions {
							if pi.lastMessage.Text == "" {
								return Dim{}
							}
							timeVal, _ := time.Parse(time.RFC3339, pi.lastMessage.Created)
							timeVal = timeVal.Local()
							txt := timeVal.Format("3:04 PM")
							label := material.Label(pi.Theme, pi.Theme.TextSize*0.75, txt)
							label.Color = pi.Theme.ContrastBg
							label.Font.Style = text.Italic
							label.Font.Weight = text.SemiBold
							return label.Layout(gtx)
						}),
						layout.Rigid(layout.Spacer{Height: unit.Dp(8)}.Layout),
						layout.Rigid(func(gtx layout.Context) layout.Dimensions {
							chatroomURL := ChatRoomPageURL + "/" + URL(pi.contact.PublicKey)
							isActive := pi.CurrentPage().URL() == chatroomURL
							if pi.messagesCount <= 0 || isActive {
								return Dim{}
							}
							maxDim := gtx.Dp(32)
							maxSize := image.Point{X: maxDim, Y: maxDim}
							gtx.Constraints.Max = maxSize
							gtx.Constraints.Min = maxSize
							mac := op.Record(gtx.Ops)
							inset := layout.UniformInset(unit.Dp(4))
							flex := layout.Flex{Alignment: layout.Middle, Spacing: layout.SpaceSides}
							d := inset.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
								return flex.Layout(gtx, layout.Rigid(func(gtx layout.Context) layout.Dimensions {
									gtx.Constraints.Min = gtx.Constraints.Max
									return layout.Center.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
										count := fmt.Sprintf("%d", pi.messagesCount)
										label := material.Label(pi.Theme, pi.Theme.TextSize*0.8, count)
										label.Alignment = text.Middle
										label.Color = pi.Theme.ContrastFg
										label.Font.Weight = text.Bold
										return component.TruncatingLabelStyle(label).Layout(gtx)
									})
								}))
							})
							stop := mac.Stop()
							component.Rect{
								Color: pi.Theme.ContrastBg,
								Size:  image.Point{X: maxDim, Y: maxDim},
								Radii: maxDim / 2,
							}.Layout(gtx)
							stop.Add(gtx.Ops)
							return d
						}),
					)
				}),
			)
			return d
		})
		return d
	})
	return d
}

func (pi *pageItem) fetchMessagesUnreadCount() {
	if !pi.fetchingUnReadMessagesCount {
		pi.fetchingUnReadMessagesCount = true
		go func() {
			pi.fetchingUnReadMessagesCountCh <- <-pi.Service().UnreadMessagesCount(pi.contact.PublicKey)
			pi.Window().Invalidate()
		}()
	}
}

func (pi *pageItem) fetchLastMessage() {
	if !pi.fetchingLastMessage {
		pi.fetchingLastMessage = true
		go func() {
			pi.fetchingLastMessageCh <- <-pi.Service().LastMessage(pi.contact.PublicKey)
			pi.Window().Invalidate()
		}()
	}
}

func (pi *pageItem) listenToLastMessage() {
	shouldBreak := false
	for {
		select {
		case pi.lastMessage = <-pi.fetchingLastMessageCh:
			pi.fetchingLastMessage = false
		default:
			shouldBreak = true
		}
		if shouldBreak {
			break
		}
	}
}

func (pi *pageItem) listenToUnreadMessagesCount() {
	shouldBreak := false
	for {
		select {
		case pi.messagesCount = <-pi.fetchingUnReadMessagesCountCh:
			pi.fetchingUnReadMessagesCount = false
		default:
			shouldBreak = true
		}
		if shouldBreak {
			break
		}
	}

}

func (p *pageItem) OnDatabaseChange(event service.Event) {
	switch e := event.Data.(type) {
	case service.MessagesCountChangedEventData:
		if p.contact.PublicKey == e.ContactPublicKey {
			p.fetchMessagesUnreadCount()
		}
	case service.MessagesStateChangedEventData:
		if p.contact.PublicKey == e.ContactPublicKey {
			p.fetchMessagesUnreadCount()
		}
	}
}

func (pi *pageItem) URL() URL {
	return ChatPageURL + "/" + URL(pi.contact.PublicKey)
}
