package view

import (
	"gioui.org/layout"
	"gioui.org/op"
	"gioui.org/op/clip"
	"gioui.org/op/paint"
	"gioui.org/widget"
	"gioui.org/widget/material"
	"gioui.org/x/component"
	"github.com/mearaj/protonet/assets"
	. "github.com/mearaj/protonet/ui/fwk"
	"golang.org/x/exp/shiny/materialdesign/icons"
	"image"
)

type AvatarView struct {
	Size  image.Point
	Image image.Image
	widget.Clickable
	*material.Theme
	Selected      bool
	SelectionMode bool
}

func (v *AvatarView) Layout(gtx Gtx) Dim {
	if v.Size == (image.Point{}) {
		v.Size = image.Point{X: gtx.Dp(48), Y: gtx.Dp(48)}
	}
	gtx.Constraints.Min, gtx.Constraints.Max = v.Size, v.Size
	var imgWidget widget.Image
	if v.Image == nil {
		v.Image = assets.AppIconImage
		imgOps := paint.NewImageOp(v.Image)
		imgWidget = widget.Image{Src: imgOps, Fit: widget.Fill, Position: layout.Center, Scale: 0}
	} else {
		imgOps := paint.NewImageOp(v.Image)
		imgWidget = widget.Image{Src: imgOps, Fit: widget.Fill, Position: layout.Center, Scale: 0}
	}
	stack := layout.Stack{Alignment: layout.SE}
	return stack.Layout(gtx,
		layout.Stacked(func(gtx Gtx) Dim {
			ops := clip.UniformRRect(image.Rectangle{
				Max: image.Point{
					X: gtx.Constraints.Max.X,
					Y: gtx.Constraints.Max.Y,
				},
			}, gtx.Constraints.Max.X/2).Push(gtx.Ops)
			defer ops.Pop()
			return imgWidget.Layout(gtx)
		}),
		layout.Stacked(func(gtx Gtx) Dim {
			if !v.SelectionMode {
				return Dim{}
			}
			gtx.Constraints.Max.X = int(float64(v.Size.X) * 0.40)
			gtx.Constraints.Max.Y = int(float64(v.Size.Y) * 0.40)
			gtx.Constraints.Min = gtx.Constraints.Max
			offsetOp := op.Offset(
				image.Pt(gtx.Constraints.Max.X/4, gtx.Constraints.Max.Y/4),
			).Push(gtx.Ops)
			defer offsetOp.Pop()
			clr := v.Theme.ContrastBg
			d := component.Rect{Size: gtx.Constraints.Max, Color: clr, Radii: gtx.Constraints.Max.X / 2}.Layout(gtx)
			iconClr := v.Theme.ContrastFg
			icon, _ := widget.NewIcon(icons.ActionCheckCircle)
			if !v.Selected {
				return d
			}
			return icon.Layout(gtx, iconClr)
		}),
	)

}
