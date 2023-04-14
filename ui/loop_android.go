package ui

import (
	"gioui.org/app"
	"github.com/mearaj/audio"
)

func androidViewEvent(e interface{}) {
	if e, ok := e.(app.ViewEvent); ok {
		audio.SetView(e.View)
	}
}
