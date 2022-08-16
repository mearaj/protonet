package view

import (
	"gioui.org/layout"
	"gioui.org/widget/material"
	"github.com/mearaj/protonet/assets/fonts"
	. "github.com/mearaj/protonet/ui/fwk"
)

type ErrorView struct {
	*material.Theme
	Manager
	Error string
}

func (i *ErrorView) Layout(gtx Gtx) (d Dim) {
	if i.Theme == nil {
		i.Theme = fonts.NewTheme()
	}
	if i.Error != "" {
		return layout.Flex{Axis: layout.Vertical}.Layout(gtx,
			layout.Rigid(func(gtx Gtx) Dim {
				return material.Body1(i.Theme, i.Error).Layout(gtx)
			}),
		)
	}
	return d
}
