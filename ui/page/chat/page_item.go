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
	"github.com/mearaj/protonet/alog"
	"github.com/mearaj/protonet/internal/chat"
	"github.com/mearaj/protonet/internal/pubsub"
	"github.com/mearaj/protonet/internal/wallet"
	. "github.com/mearaj/protonet/ui/fwk"
	"github.com/mearaj/protonet/ui/page/chatroom"
	"github.com/mearaj/protonet/ui/view"
	"image"
)

type pageItem struct {
	Manager
	contact chat.Contact
	widget.Clickable
	*material.Theme
	view.AvatarView
	lastMessage   chat.Message
	messagesCount int64
}

func newChatPageItem(manager Manager, contact chat.Contact) *pageItem {
	img, _, _ := image.Decode(bytes.NewReader(contact.Avatar))
	i := pageItem{
		Manager: manager,
		Theme:   manager.Theme(),
		contact: contact,
		AvatarView: view.AvatarView{
			Theme: manager.Theme(),
			Image: img,
		},
	}
	return &i
}

func (pi *pageItem) Layout(gtx Gtx) Dim {
	pi.fetchLastMessage()
	pi.fetchMessagesUnreadCount()
	return pi.layoutContent(gtx)
}

func (pi *pageItem) layoutContent(gtx Gtx) Dim {
	gtx.Constraints.Min.X = gtx.Constraints.Max.X
	btnStyle := material.ButtonLayoutStyle{Background: pi.Theme.ContrastBg, Button: &pi.Clickable}

	if pi.Clickable.Clicked() {
		chatRoomPage := chatroom.New(pi.Manager, pi.contact)
		if pi.CurrentPage().URL() != chatRoomPage.URL() {
			pi.NavigateToURL(ChatPageURL, func() {
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
							timeVal := pi.lastMessage.CreatedAt
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
							timeVal := pi.lastMessage.CreatedAt
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
	var err error
	var count int64
	defer func() {
		if err != nil {
			alog.Logger().Errorln(err)
		}
	}()
	acc, err := wallet.GlobalWallet.Account()
	if err != nil {
		return
	}
	count, err = wallet.GlobalWallet.UnreadMessagesCount(acc.PublicKey, pi.contact.PublicKey)
	if err != nil {
		return
	}
	if count != pi.messagesCount {
		pi.messagesCount = count
		pi.Window().Invalidate()
	}
}

func (pi *pageItem) fetchLastMessage() {
	var err error
	var lastMessage chat.Message
	defer func() {
		if err != nil {
			alog.Logger().Errorln(err)
		}
	}()
	acc, err := wallet.GlobalWallet.Account()
	if err != nil {
		return
	}
	lastMessage, err = wallet.GlobalWallet.LastMessage(acc.PublicKey, pi.contact.PublicKey)
	if err != nil {
		alog.Logger().Errorln(err)
	}
	if lastMessage.ID != pi.lastMessage.ID {
		pi.lastMessage = lastMessage
		pi.Window().Invalidate()
	}
}

func (p *pageItem) OnDatabaseChange(event pubsub.Event) {
	switch e := event.Data.(type) {
	case pubsub.NewMessageReceivedEventData:
		account, _ := wallet.GlobalWallet.Account()
		accountPubKey := account.PublicKey
		contactPubKey := p.contact.PublicKey
		if (e.Sender == contactPubKey || e.Recipient == contactPubKey) &&
			(e.Sender == accountPubKey || e.Recipient == accountPubKey) {
			p.fetchMessagesUnreadCount()
			p.fetchLastMessage()
		}
	}
}

func (pi *pageItem) URL() URL {
	return ChatPageURL + "/" + URL(pi.contact.PublicKey)
}
