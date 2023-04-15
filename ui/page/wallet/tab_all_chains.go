package wallet

import (
	"fmt"
	"gioui.org/layout"
	"gioui.org/unit"
	"gioui.org/widget"
	"gioui.org/widget/material"
	"gioui.org/x/component"
	"github.com/mearaj/protonet/alog"
	"github.com/mearaj/protonet/assets/fonts"
	"github.com/mearaj/protonet/internal/evm"
	"github.com/mearaj/protonet/internal/wallet"
	"github.com/mearaj/protonet/ui/fwk"
	"github.com/mearaj/protonet/ui/view"
	"golang.org/x/image/colornames"
	"image"
	"image/color"
	"strings"
)

type tabAllChains struct {
	initialized bool
	layout.List
	*page
	title              string
	chainItems         []*tabAllChainsConnItem
	chainItemsFiltered []*tabAllChainsConnItem
	filterText         string
}

func (p *tabAllChains) init() {
	p.chainItems = make([]*tabAllChainsConnItem, len(evm.ChainsSlice()))
	for i, ch := range wallet.GlobalWallet.Connections() {
		p.chainItems[i] = &tabAllChainsConnItem{
			ConnChain: ch,
			Theme:     p.Theme,
		}
	}
	filteredItems := make([]*tabAllChainsConnItem, len(p.chainItems))
	copy(filteredItems, p.chainItems)
	p.chainItemsFiltered = append(p.chainItemsFiltered[:0], filteredItems...)
	p.initialized = true
}

func (p *tabAllChains) drawTabHead(gtx fwk.Gtx) fwk.Dim {
	if !p.initialized {
		p.init()
	}
	gtx.Constraints.Max.X = p.width
	inset := layout.UniformInset(12)
	maxWidth := p.width / len(p.Tabs.Tabs)
	gtx.Constraints.Max.X, gtx.Constraints.Min.X = maxWidth, maxWidth
	return inset.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
		return layout.Center.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
			d := material.H6(p.Theme, p.title).Layout(gtx)
			return d
		})
	})
}

func (p *tabAllChains) drawTabBody(gtx fwk.Gtx) fwk.Dim {
	if !p.initialized {
		p.init()
	}
	filterText := strings.TrimSpace(strings.ToLower(p.filterText))
	searchText := strings.TrimSpace(strings.ToLower(p.search.Text()))
	if filterText != searchText {
		p.filterText = searchText
		filteredItems := make([]*tabAllChainsConnItem, 0)
		for _, ch := range p.chainItems {
			if len(ch.ConnChain.RPCClients) == 0 {
				continue
			}
			text1 := strings.ToLower(ch.ConnChain.Chain.Name)
			text2 := strings.ToLower(ch.ConnChain.Chain.ShortName)
			text3 := strings.ToLower(ch.ConnChain.Chain.Chain)
			shouldContain := strings.Contains(text1, searchText) || strings.Contains(text2, searchText) ||
				strings.Contains(text3, searchText)
			if shouldContain {
				filteredItems = append(filteredItems, ch)
			}
		}
		p.chainItemsFiltered = append(p.chainItemsFiltered[:0], filteredItems...)
	}
	inset := layout.UniformInset(16)
	return inset.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
		return p.allChainsTab.List.Layout(gtx, len(p.chainItemsFiltered), func(gtx layout.Context, index int) layout.Dimensions {
			if len(p.chainItemsFiltered[index].ConnChain.RPCClients) == 0 {
				return layout.Dimensions{}
			}
			flex := layout.Flex{Axis: layout.Vertical}
			return flex.Layout(gtx,
				layout.Rigid(func(gtx layout.Context) layout.Dimensions {
					inset := layout.Inset{Top: 12, Bottom: 12}
					return inset.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
						return p.chainItemsFiltered[index].Layout(gtx)
					})
				}),
				layout.Rigid(component.Divider(p.Theme).Layout),
			)
		})
	})
}

type tabAllChainsConnItem struct {
	ConnChain      *evm.RPCClients
	connStateItems []*tabAllChainsConnStateItem
	layout.List
	widget.Clickable
	InsetHeader layout.Inset
	*material.Theme
	initialized bool
}

func (c *tabAllChainsConnItem) Layout(gtx fwk.Gtx) fwk.Dim {
	if !c.initialized {
		if c.Theme == nil {
			c.Theme = fonts.NewTheme()
		}
		if len(c.ConnChain.RPCClients) > 0 {
			c.connStateItems = make([]*tabAllChainsConnStateItem, len(c.ConnChain.RPCClients))
			for i, connState := range c.ConnChain.RPCClients {
				c.connStateItems[i] = &tabAllChainsConnStateItem{RPCClient: connState, Theme: c.Theme}
			}
		}
		if c.InsetHeader == (layout.Inset{}) {
			c.InsetHeader = layout.Inset{Bottom: 4}
		}
		c.initialized = true
	}
	flex := layout.Flex{Axis: layout.Vertical}
	return flex.Layout(gtx,
		layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			btnStyle := material.ButtonLayout(c.Theme, &c.Clickable)
			return btnStyle.Button.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
				return c.InsetHeader.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
					return material.H5(c.Theme, c.ConnChain.Chain.Name).Layout(gtx)
				})
			})
		}),
		layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			inset := layout.Inset{Bottom: 8}
			return inset.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
				flex := layout.Flex{}
				return flex.Layout(gtx,
					layout.Rigid(func(gtx layout.Context) layout.Dimensions {
						txt := fmt.Sprintf("Currency: %s", c.ConnChain.Chain.NativeCurrency.Name)
						w := material.Body1(c.Theme, txt)
						w.TextSize = unit.Sp(16)
						return w.Layout(gtx)
					}),
				)
			})
		}),
		layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			c.List.Axis = layout.Vertical
			return c.List.Layout(gtx, len(c.ConnChain.RPCClients), func(gtx layout.Context, index int) layout.Dimensions {
				return c.connStateItems[index].Layout(gtx)
			})
		}),
	)
}

type tabAllChainsConnStateItem struct {
	*evm.RPCClient
	btnConnect widget.Clickable
	*material.Theme
	State
	layout.Inset
	initialized bool
	bal         string
	err         error
	balFetched  bool
}

func (c *tabAllChainsConnStateItem) Layout(gtx fwk.Gtx) fwk.Dim {
	if !c.initialized {
		if c.Theme == nil {
			c.Theme = fonts.NewTheme()
		}
		c.initialized = true
	}
	inset := layout.Inset{Bottom: 12}
	isConnected := c.IsConnected()
	return inset.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
		flex := layout.Flex{Axis: layout.Vertical}
		return flex.Layout(gtx,
			layout.Rigid(func(gtx layout.Context) layout.Dimensions {
				inset := layout.Inset{Bottom: 4}
				return inset.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
					w := material.Body1(c.Theme, string(c.RPCClient.RPC))
					w.TextSize = unit.Sp(16)
					return w.Layout(gtx)
				})
			}),
			layout.Rigid(func(gtx layout.Context) layout.Dimensions {
				flex := layout.Flex{}
				return flex.Layout(gtx, layout.Rigid(func(gtx layout.Context) layout.Dimensions {
					flex := layout.Flex{Axis: layout.Vertical, Alignment: layout.Start}
					return flex.Layout(gtx,
						layout.Rigid(func(gtx layout.Context) layout.Dimensions {
							inset := layout.Inset{Bottom: 8}
							if c.State == StateConnecting ||
								c.State == StateDisconnecting {
								loader := view.Loader{Theme: c.Theme, Size: image.Pt(gtx.Dp(28), gtx.Dp(28))}
								gtx.Constraints.Max.Y = 40
								return inset.Layout(gtx, loader.Layout)
							}
							txt := "Connect"
							btnStyle := material.Button(c.Theme, &c.btnConnect, txt)
							if c.btnConnect.Clicked() && c.State == StateIdle {
								c.balFetched = false
								switch isConnected {
								case true:
									c.State = StateDisconnecting
									err := c.RPCClient.Disconnect()
									if err != nil {
										alog.Logger().Errorln(err)
									}
									c.State = StateIdle
								case false:
									c.State = StateConnecting
									err := c.RPCClient.Connect()
									if err != nil {
										alog.Logger().Errorln(err)
									}
									c.State = StateIdle
								}
							}
							btnStyle.Background = color.NRGBA(colornames.Green)
							if isConnected {
								txt = "Disconnect"
								btnStyle.Text = txt
								btnStyle.Background = color.NRGBA(colornames.Red)
							}
							return inset.Layout(gtx, btnStyle.Layout)
						}),
						layout.Rigid(func(gtx layout.Context) layout.Dimensions {
							if !isConnected || !wallet.GlobalWallet.IsOpen() {
								return fwk.Dim{}
							}
							fetched := c.balFetched
							if !fetched {
								go func() {
									acc, _ := wallet.GlobalWallet.Account()
									c.bal, c.err = c.ShowBalance(acc)
									c.balFetched = true
								}()
							}
							inset := layout.Inset{Bottom: 8}
							return inset.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
								flex := layout.Flex{}
								return flex.Layout(gtx,
									layout.Rigid(func(gtx layout.Context) layout.Dimensions {
										if c.err != nil {
											return fwk.Dim{}
										}
										return material.Body1(c.Theme, c.bal).Layout(gtx)
									}))
							})
						}),
					)
				}))
			}),
		)
	})
}
