package main

import (
	"gioui.org/app"
	"github.com/mearaj/protonet/ui"
	log "github.com/sirupsen/logrus"
	"os"
)

func main() {
	go func() {
		w := app.NewWindow(app.Title("Protonet"))
		if err := ui.Loop(w); err != nil {
			log.Fatal(err)
		}
		os.Exit(0)
	}()
	app.Main()
}
