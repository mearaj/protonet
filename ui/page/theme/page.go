package theme

import (
	"fmt"
	"gioui.org/io/pointer"
	"gioui.org/layout"
	"gioui.org/op"
	"gioui.org/op/clip"
	"gioui.org/op/paint"
	"gioui.org/text"
	"gioui.org/unit"
	"gioui.org/widget"
	"gioui.org/widget/material"
	"gioui.org/x/colorpicker"
	"gioui.org/x/component"
	"github.com/mearaj/protonet/alog"
	"github.com/mearaj/protonet/assets/fonts"
	. "github.com/mearaj/protonet/ui/fwk"
	"github.com/mearaj/protonet/ui/view"
	"golang.org/x/exp/shiny/materialdesign/colornames"
	"golang.org/x/exp/shiny/materialdesign/icons"
	"image"
	"image/color"
	"time"
)

type page struct {
	layout.List
	Manager
	Theme           *material.Theme
	OrigTheme       *material.Theme
	SavedTheme      *material.Theme
	title           string
	btnNavIcon      widget.Clickable
	btnMenuIcon     widget.Clickable
	btnMenuContent  widget.Clickable
	btnSaveTheme    widget.Clickable
	btnResetTheme   widget.Clickable
	btnDefaultTheme widget.Clickable
	navIcon         *widget.Icon
	menuIcon        *widget.Icon
	colorpicker.MuxState
	colorpicker.State
	menuVisibilityAnim component.VisibilityAnimation
}

func New(manager Manager) Page {
	navIcon, _ := widget.NewIcon(icons.NavigationArrowBack)
	menuIcon, _ := widget.NewIcon(icons.NavigationMoreVert)
	OrigTheme := *manager.Theme()
	Theme := *manager.Theme()
	SavedTheme := *manager.Theme()
	pg := page{
		Manager:    manager,
		Theme:      &Theme,
		OrigTheme:  &OrigTheme,
		SavedTheme: &SavedTheme,
		title:      "Theme",
		navIcon:    navIcon,
		List:       layout.List{Axis: layout.Vertical},
		menuIcon:   menuIcon,
		menuVisibilityAnim: component.VisibilityAnimation{
			Duration: time.Millisecond * 250,
			State:    component.Invisible,
			Started:  time.Time{},
		},
	}
	pg.MuxState = colorpicker.NewMuxState([]colorpicker.MuxOption{
		{
			Label: "Contrast Background",
			Value: &pg.Theme.ContrastBg,
		},
		{
			Label: "Contrast Foreground",
			Value: &pg.Theme.ContrastFg,
		},
		{
			Label: "Background",
			Value: &pg.Theme.Bg,
		},
		{
			Label: "Foreground",
			Value: &pg.Theme.Fg,
		},
	}...)
	pg.State.SetColor(*pg.MuxState.Color())

	return &pg
}

func (p *page) Layout(gtx Gtx) Dim {
	flex := layout.Flex{Axis: layout.Vertical,
		Spacing:   layout.SpaceEnd,
		Alignment: layout.Start,
	}

	d := flex.Layout(gtx,
		layout.Rigid(p.DrawAppBar),
		layout.Rigid(p.drawContentLayout),
	)
	p.drawMenuLayout(gtx)
	for _, e := range gtx.Queue.Events(p) {
		switch e := e.(type) {
		case pointer.Event:
			switch e.Type {
			case pointer.Press:
				if !p.btnMenuContent.Pressed() {
					p.menuVisibilityAnim.Disappear(gtx.Now)
				}
			}
		}
	}
	return d
}

func (p *page) DrawAppBar(gtx Gtx) Dim {
	if p.btnNavIcon.Clicked() {
		p.PopUp()
	}
	gtx.Constraints.Min.X = gtx.Constraints.Max.X
	gtx.Constraints.Max.Y = gtx.Dp(56)
	th := p.OrigTheme
	return view.DrawAppBarLayout(gtx, th, func(gtx Gtx) Dim {
		return layout.Flex{Alignment: layout.Middle, Spacing: layout.SpaceBetween}.Layout(gtx,
			layout.Rigid(func(gtx Gtx) Dim {
				return layout.Flex{Alignment: layout.Middle}.Layout(gtx,
					layout.Rigid(func(gtx Gtx) Dim {
						navigationIcon := p.navIcon
						button := material.IconButton(th, &p.btnNavIcon, navigationIcon, "Nav Icon Button")
						button.Size = unit.Dp(40)
						button.Background = th.Palette.ContrastBg
						button.Color = th.Palette.ContrastFg
						button.Inset = layout.UniformInset(unit.Dp(8))
						return button.Layout(gtx)
					}),
					layout.Rigid(func(gtx Gtx) Dim {
						return layout.Inset{Left: unit.Dp(16)}.Layout(gtx, func(gtx Gtx) Dim {
							titleText := p.title
							title := material.Body1(th, titleText)
							title.Color = th.Palette.ContrastFg
							title.TextSize = unit.Sp(18)
							return title.Layout(gtx)
						})
					}),
				)
			}),
			layout.Rigid(func(gtx layout.Context) layout.Dimensions {
				if p.btnMenuIcon.Clicked() {
					p.menuVisibilityAnim.Appear(gtx.Now)
				}
				button := material.IconButton(th, &p.btnMenuIcon, p.menuIcon, "Context Menu")
				button.Size = unit.Dp(40)
				button.Background = th.Palette.ContrastBg
				button.Color = th.Palette.ContrastFg
				button.Inset = layout.UniformInset(unit.Dp(8))
				d := button.Layout(gtx)
				return d
			}),
		)
	})

}

func (p *page) drawContentLayout(gtx Gtx) Dim {
	th := fonts.NewTheme()
	paint.FillShape(gtx.Ops, th.Bg,
		clip.UniformRRect(image.Rect(0, 0, gtx.Constraints.Max.X, gtx.Constraints.Max.Y), 0).Op(gtx.Ops))
	inset := layout.UniformInset(unit.Dp(16))
	gtx.Constraints.Min = gtx.Constraints.Max
	if p.MuxState.Changed() {
		p.State.SetColor(*p.MuxState.Color())
		alog.Logger().Debugln("mux state changed")
	}
	if p.State.Changed() {
		k := p.MuxState.Value
		clr := p.MuxState.Options[k]
		clr.R = p.State.Color().R
		clr.G = p.State.Color().G
		clr.B = p.State.Color().B
		clr.A = p.State.Color().A
		p.State.Editor.SetText(fmt.Sprintf("%02x%02x%02x%02x", clr.R, clr.G, clr.B, clr.A))
	}
	return inset.Layout(gtx,
		func(gtx layout.Context) layout.Dimensions {
			p.List.Axis = layout.Vertical
			return p.List.Layout(gtx, 1, func(gtx layout.Context, index int) layout.Dimensions {
				flex := layout.Flex{Axis: layout.Vertical}
				return flex.Layout(gtx,
					layout.Rigid(func(gtx layout.Context) layout.Dimensions {
						return layout.Flex{Alignment: layout.Middle}.Layout(gtx,
							layout.Rigid(func(gtx layout.Context) layout.Dimensions {
								gtx.Constraints.Min.X = gtx.Dp(50)
								return material.Body1(th, "Red").Layout(gtx)
							}),
							layout.Flexed(1, func(gtx layout.Context) layout.Dimensions {
								return material.Slider(th, &p.State.R, 0, 1).Layout(gtx)
							}),
							layout.Rigid(layout.Spacer{Width: unit.Dp(4)}.Layout),
							p.drawColorIntBox(p.State.Color().R),
						)
					}),
					layout.Rigid(func(gtx layout.Context) layout.Dimensions {
						return layout.Flex{Alignment: layout.Middle}.Layout(gtx,
							layout.Rigid(func(gtx layout.Context) layout.Dimensions {
								gtx.Constraints.Min.X = gtx.Dp(50)
								return material.Body1(th, "Green").Layout(gtx)
							}),
							layout.Flexed(1, func(gtx layout.Context) layout.Dimensions {
								return material.Slider(th, &p.State.G, 0, 1).Layout(gtx)
							}),
							layout.Rigid(layout.Spacer{Width: unit.Dp(4)}.Layout),
							p.drawColorIntBox(p.State.Color().G),
						)
					}),
					layout.Rigid(func(gtx layout.Context) layout.Dimensions {
						return layout.Flex{Alignment: layout.Middle}.Layout(gtx,
							layout.Rigid(func(gtx layout.Context) layout.Dimensions {
								gtx.Constraints.Min.X = gtx.Dp(50)
								return material.Body1(th, "Blue").Layout(gtx)
							}),
							layout.Flexed(1, func(gtx layout.Context) layout.Dimensions {
								return material.Slider(th, &p.State.B, 0, 1).Layout(gtx)
							}),
							layout.Rigid(layout.Spacer{Width: unit.Dp(4)}.Layout),
							p.drawColorIntBox(p.State.Color().B),
						)
					}),
					layout.Rigid(func(gtx layout.Context) layout.Dimensions {
						return layout.Flex{Alignment: layout.Middle}.Layout(gtx,
							layout.Rigid(func(gtx layout.Context) layout.Dimensions {
								gtx.Constraints.Min.X = gtx.Dp(50)
								return material.Body1(th, "Alpha").Layout(gtx)
							}),
							layout.Flexed(1, func(gtx layout.Context) layout.Dimensions {
								return material.Slider(th, &p.State.A, 0, 1).Layout(gtx)
							}),
							layout.Rigid(layout.Spacer{Width: unit.Dp(4)}.Layout),
							p.drawColorIntBox(p.State.Color().A),
						)
					}),
					layout.Rigid(layout.Spacer{Height: unit.Dp(8)}.Layout),
					layout.Rigid(func(gtx layout.Context) layout.Dimensions {
						return layout.Flex{Alignment: layout.Middle}.Layout(gtx,
							layout.Rigid(func(gtx layout.Context) layout.Dimensions {
								gtx.Constraints.Min.X = gtx.Dp(50)
								return material.Body1(th, "Hex").Layout(gtx)
							}),
							layout.Rigid(func(gtx layout.Context) layout.Dimensions {
								return layout.UniformInset(unit.Dp(2)).Layout(gtx, func(gtx Gtx) Dim {
									return layout.Flex{Alignment: layout.Baseline}.Layout(gtx,
										layout.Rigid(func(gtx Gtx) Dim {
											return material.Body1(p.OrigTheme, "#"+p.Editor.Text()).Layout(gtx)
										}),
									)
								})
							}),
						)
					}),
					layout.Rigid(func(gtx layout.Context) layout.Dimensions {
						gtx.Constraints.Min.X = gtx.Constraints.Max.X
						return layout.Flex{Alignment: layout.Start, Spacing: layout.SpaceEnd}.Layout(gtx,
							layout.Rigid(func(gtx layout.Context) layout.Dimensions {
								mux := colorpicker.Mux(th, &p.MuxState, "")
								return mux.Layout(gtx)
							}),
						)
					}),
					layout.Rigid(layout.Spacer{Height: unit.Dp(32)}.Layout),
					layout.Rigid(func(gtx layout.Context) layout.Dimensions {
						gtx.Constraints.Max.Y = gtx.Dp(64)
						gtx.Constraints.Min = gtx.Constraints.Max
						background := p.MuxState.Options["Contrast Background"]
						foreground := p.MuxState.Options["Contrast Foreground"]
						paint.FillShape(gtx.Ops, *background, clip.Rect(image.Rectangle{Max: gtx.Constraints.Max}).Op())
						layout.Center.Layout(gtx,
							func(gtx layout.Context) layout.Dimensions {
								body := material.Body1(p.OrigTheme, "Contrast Foreground")
								body.Color = *foreground
								return body.Layout(gtx)
							},
						)
						return Dim{Size: gtx.Constraints.Max}
					}),
					layout.Rigid(layout.Spacer{Height: unit.Dp(24)}.Layout),
					layout.Rigid(func(gtx layout.Context) layout.Dimensions {
						gtx.Constraints.Max.Y = gtx.Dp(64)
						gtx.Constraints.Min = gtx.Constraints.Max
						background := p.MuxState.Options["Background"]
						foreground := p.MuxState.Options["Foreground"]
						paint.FillShape(gtx.Ops, *background, clip.Rect(image.Rectangle{Max: gtx.Constraints.Max}).Op())
						layout.Center.Layout(gtx,
							func(gtx layout.Context) layout.Dimensions {
								body := material.Body1(p.OrigTheme, "Foreground")
								body.Color = *foreground
								return body.Layout(gtx)
							},
						)
						return Dim{Size: gtx.Constraints.Max}
					}),
					layout.Rigid(layout.Spacer{Height: unit.Dp(32)}.Layout),
				)
			})
		},
	)
}

func (p *page) drawColorIntBox(num uint8) layout.FlexChild {
	return layout.Rigid(func(gtx layout.Context) layout.Dimensions {
		gtx.Constraints.Max.X = gtx.Dp(40)
		gtx.Constraints.Max.Y = gtx.Dp(32)
		gtx.Constraints.Min = gtx.Constraints.Max
		paint.FillShape(gtx.Ops, color.NRGBA(colornames.White), clip.Rect(image.Rectangle{Max: gtx.Constraints.Max}).Op())
		layout.Center.Layout(gtx,
			func(gtx layout.Context) layout.Dimensions {
				body := material.Body1(p.OrigTheme, fmt.Sprintf("%d", num))
				body.Color = color.NRGBA(colornames.Black)
				return body.Layout(gtx)
			},
		)
		bounds := image.Rect(0, 0, gtx.Constraints.Max.X, gtx.Constraints.Max.Y)
		rect := clip.UniformRRect(bounds, gtx.Dp(4))
		paint.FillShape(gtx.Ops,
			color.NRGBA(colornames.Black),
			clip.Stroke{Path: rect.Path(gtx.Ops), Width: float32(gtx.Dp(1))}.Op(),
		)
		return Dim{Size: gtx.Constraints.Max}
	})
}

func (p *page) drawMenuLayout(gtx Gtx) Dim {
	return layout.Stack{Alignment: layout.NE}.Layout(gtx,
		layout.Stacked(func(gtx layout.Context) layout.Dimensions {
			progress := p.menuVisibilityAnim.Revealed(gtx)
			gtx.Constraints.Max.X = int(float32(gtx.Constraints.Max.X) * progress)
			gtx.Constraints.Max.Y = int(float32(gtx.Constraints.Max.Y) * progress)
			return component.Rect{Size: gtx.Constraints.Max, Color: color.NRGBA{A: 200}}.Layout(gtx)
		}),
		layout.Stacked(func(gtx layout.Context) layout.Dimensions {
			progress := p.menuVisibilityAnim.Revealed(gtx)
			macro := op.Record(gtx.Ops)
			d := p.btnMenuContent.Layout(gtx, p.drawMenuItems)
			call := macro.Stop()
			d.Size.X = int(float32(d.Size.X) * progress)
			d.Size.Y = int(float32(d.Size.Y) * progress)
			component.Rect{Size: d.Size, Color: color.NRGBA(colornames.White)}.Layout(gtx)
			clipOp := clip.Rect{Max: d.Size}.Push(gtx.Ops)
			call.Add(gtx.Ops)
			clipOp.Pop()
			return d
		}),
	)
}

func (p *page) drawMenuItems(gtx Gtx) Dim {
	inset := layout.UniformInset(unit.Dp(12))
	gtx.Constraints.Max.X = int(float32(gtx.Constraints.Max.X) / 1.5)
	gtx.Constraints.Min.X = gtx.Constraints.Max.X
	if p.btnSaveTheme.Clicked() {
		*p.Manager.Theme() = *p.Theme
		*p.OrigTheme = *p.Theme
		p.menuVisibilityAnim.Disappear(gtx.Now)
		op.InvalidateOp{}.Add(gtx.Ops)
	}

	if p.btnResetTheme.Clicked() {
		*p.Theme = *p.SavedTheme
		*p.Manager.Theme() = *p.Theme
		*p.OrigTheme = *p.Theme
		p.State.SetColor(*p.MuxState.Color())
		p.menuVisibilityAnim.Disappear(gtx.Now)
		op.InvalidateOp{}.Add(gtx.Ops)
	}

	if p.btnDefaultTheme.Clicked() {
		*p.Manager.Theme() = *fonts.NewTheme()
		*p.Theme = *p.Manager.Theme()
		*p.OrigTheme = *p.Manager.Theme()
		p.State.SetColor(*p.MuxState.Color())
		p.menuVisibilityAnim.Disappear(gtx.Now)
		op.InvalidateOp{}.Add(gtx.Ops)
	}

	return layout.Flex{Axis: layout.Vertical, Alignment: layout.Start}.Layout(gtx,
		layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			btnStyle := material.ButtonLayoutStyle{Button: &p.btnSaveTheme}
			btnStyle.Background = color.NRGBA(colornames.White)
			return btnStyle.Layout(gtx,
				func(gtx Gtx) Dim {
					gtx.Constraints.Min.X = gtx.Constraints.Max.X
					inset := inset
					return inset.Layout(gtx, func(gtx Gtx) Dim {
						return layout.Flex{Spacing: layout.SpaceEnd}.Layout(gtx,
							layout.Rigid(func(gtx Gtx) Dim {
								bd := material.Body1(p.Theme, "Save")
								bd.Color = color.NRGBA(colornames.Black)
								bd.Alignment = text.Start
								return bd.Layout(gtx)
							}),
						)
					})
				},
			)
		}),
		layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			btnStyle := material.ButtonLayoutStyle{Button: &p.btnResetTheme}
			btnStyle.Background = color.NRGBA(colornames.White)
			return btnStyle.Layout(gtx,
				func(gtx Gtx) Dim {
					gtx.Constraints.Min.X = gtx.Constraints.Max.X
					inset := inset
					return inset.Layout(gtx, func(gtx Gtx) Dim {
						return layout.Flex{Spacing: layout.SpaceEnd}.Layout(gtx,
							layout.Rigid(func(gtx Gtx) Dim {
								bd := material.Body1(p.Theme, "Reset")
								bd.Color = color.NRGBA(colornames.Black)
								bd.Alignment = text.Start
								return bd.Layout(gtx)
							}),
						)
					})
				},
			)
		}),
		layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			btnStyle := material.ButtonLayoutStyle{Button: &p.btnDefaultTheme}
			btnStyle.Background = color.NRGBA(colornames.White)
			return btnStyle.Layout(gtx,
				func(gtx Gtx) Dim {
					gtx.Constraints.Min.X = gtx.Constraints.Max.X
					inset := inset
					return inset.Layout(gtx, func(gtx Gtx) Dim {
						return layout.Flex{Spacing: layout.SpaceEnd}.Layout(gtx,
							layout.Rigid(func(gtx Gtx) Dim {
								bd := material.Body1(p.Theme, "Default")
								bd.Color = color.NRGBA(colornames.Black)
								bd.Alignment = text.Start
								return bd.Layout(gtx)
							}),
						)
					})
				},
			)
		}),
	)
}

func (p *page) URL() URL {
	return ThemePageURL
}
