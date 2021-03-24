package view

import (
	"gioui.org/layout"
	"gioui.org/unit"
	"gioui.org/widget"
	"gioui.org/widget/material"
	"gioui.org/x/component"
	"golang.org/x/exp/shiny/materialdesign/icons"
	"image/color"
	log "github.com/sirupsen/logrus"
	"protonet.live/database"
	"protonet.live/service"
	"runtime"
	"strings"
	"time"
)

type ChatRoom struct {
	nav                    *Navigator
	list                   layout.List
	th                     material.Theme
	overflowActions        []component.OverflowAction
	overflowState          widget.Clickable
	backIcon               *widget.Icon
	Name                   string
	inpTxtField            component.TextField
	submitIcon             *widget.Icon
	submitButton           widget.Clickable
	cs                     *service.TxtChatService
	hs                     *service.ChatService
	msgs                   []*database.TxtMsg
	markAllAsReadTimestamp int64
}

func NewChatRoom(nav *Navigator, cs *service.TxtChatService) (cr *ChatRoom) {
	cr = &ChatRoom{}
	cr.nav = nav
	cr.cs = cs
	cr.backIcon, _ = widget.NewIcon(icons.NavigationArrowBack)
	cr.th = nav.Theme
	cr.list = layout.List{ScrollToEnd: true}
	cr.list.Position.BeforeEnd = false
	cr.list.Axis = layout.Vertical
	cr.inpTxtField = component.TextField{Editor: widget.Editor{
		Alignment:  0,
		SingleLine: false,
		Submit:     true,
		Mask:       0,
	}}
	cr.submitIcon, _ = widget.NewIcon(icons.ContentSend)
	cr.msgs = cr.cs.TextMessagesToArray()
	cr.overflowActions = []component.OverflowAction{
		{
			Name: "Example 1",
			Tag:  &cr.overflowState,
		},
		{
			Name: "Example 2",
			Tag:  &cr.overflowActions,
		},
	}
	cr.SetBarActions()
	go func() {
		for {
			time.Sleep(time.Second / 10)
			switch nav.GetCurrentView() {
			case cr:
				if time.Now().Unix()-cr.markAllAsReadTimestamp > 1 {
					cr.markAllAsReadTimestamp = time.Now().Unix()
					if IsStageRunning {
						cr.cs.MarkAllClientMsgsAsRead()
					}
				}
			}
		}
	}()

	return cr
}

func (cr *ChatRoom) Layout(gtx C) (d D) {
	cr.handleAppBarEvents(gtx)

	fl := layout.Flex{
		Axis:      layout.Vertical,
		Spacing:   layout.SpaceBetween,
		WeightSum: 1.0,
	}
	d = fl.Layout(gtx,
		layout.Rigid(cr.setBar),
		layout.Flexed(1.0, cr.drawChatRoomList),
		layout.Rigid(cr.drawReplyLayout),
	)
	if cr.submitButton.Clicked() {
		cr.sendMessage()
	}
	for _, event := range cr.inpTxtField.Editor.Events() {
		switch event.(type) {
		case widget.SubmitEvent:
			if runtime.GOOS != "android" {
				cr.sendMessage()
			}
		}
	}
	return
}

func (cr *ChatRoom) setBar(gtx C) D {
	cr.nav.AppBar.Title = cr.cs.GetClient().Name
	cr.nav.AppBar.NavigationIcon = cr.backIcon
	d := cr.nav.AppBar.Layout(gtx, &cr.th)
	return d
}

func (cr *ChatRoom) SetBarActions() {
	cr.nav.AppBar.SetActions(nil, cr.overflowActions)
}

func (cr *ChatRoom) handleAppBarEvents(gtx C) {
	for _, event := range cr.nav.AppBar.Events(gtx) {
		switch event := event.(type) {
		case component.AppBarNavigationClicked:
			log.Println("Back Icon clicked")
			cr.nav.PopView()
		case component.AppBarContextMenuDismissed:
			log.Printf("Context menu dismissed: %v", event)
		case component.AppBarOverflowActionClicked:
			log.Printf("Overflow action selected: %v", event)
		}
	}
}

func (cr *ChatRoom) drawChatRoomList(gtx C) D {
	//if len(cr.msgs) != len(cr.cs.TextMessagesToArray()) {
	cr.msgs = cr.cs.TextMessagesToArray()
	//}
	gtx.Constraints.Min = gtx.Constraints.Max
	return layout.Flex{}.Layout(gtx,
		layout.Rigid(func(gtx C) D {
			return layout.Inset{
				Top:    unit.Dp(8),
				Right:  unit.Dp(8),
				Bottom: unit.Dp(8),
				Left:   unit.Dp(8),
			}.Layout(gtx, func(gtx C) D {
				return cr.list.Layout(gtx, len(cr.msgs), cr.drawChatRoomListItem)
			})
		}))

}

func (cr *ChatRoom) drawChatRoomListItem(gtx C, index int) D {
	msg := cr.msgs[index]
	return NewChatItem(msg, cr.cs, cr.th).Layout(gtx)
}

func (cr *ChatRoom) drawReplyLayout(gtx C) D {
	fl := layout.Flex{
		Axis:      layout.Horizontal,
		Spacing:   layout.SpaceBetween,
		Alignment: layout.End,
		WeightSum: 1,
	}
	gtx.Constraints.Max.Y = 200

	return layout.Inset{
		Top:    unit.Dp(8),
		Right:  unit.Dp(8),
		Bottom: unit.Dp(4),
		Left:   unit.Dp(8),
	}.Layout(gtx, func(gtx C) D {
		return fl.Layout(gtx,
			layout.Flexed(1.0, func(gtx C) D {
				return cr.inpTxtField.Layout(gtx, &cr.th,
					"Enter message here...")
			}),
			layout.Rigid(func(gtx C) D {
				return layout.Inset{
					Top:    unit.Value{},
					Right:  unit.Value{},
					Bottom: unit.Value{},
					Left:   unit.Dp(8.0),
				}.Layout(
					gtx,
					func(gtx C) D {
						return material.IconButtonStyle{
							Background: cr.th.ContrastBg,
							Color: color.NRGBA{
								R: 255,
								G: 255,
								B: 255,
								A: 255,
							},
							Icon:   cr.submitIcon,
							Size:   unit.Dp(24.0),
							Button: &cr.submitButton,
							Inset:  layout.UniformInset(unit.Dp(9)),
						}.Layout(gtx)
					},
				)
			}),
		)
	})
}

func (cr *ChatRoom) OnClipboardPasteRequests(text string) {
	cr.inpTxtField.SetText(text)
}

func (cr *ChatRoom) sendMessage() {
	msg := strings.TrimSpace(cr.inpTxtField.Text())
	if msg != "" {
		cr.inpTxtField.SetText("")
		txtMsg, err := cr.cs.NewTextMessage(msg)
		if err != nil {
			log.Println("Unable to create new message detail, err is:", err)
			return
		}
		go func() {
			txtMsg.Action.Type = database.Request
			txtMsg.Action.Message = database.MessageAck
			cr.cs.TxtMsgLiveOutChan <- txtMsg
			cr.cs.TxtMsgOutChan <- txtMsg
		}()
	}
}
