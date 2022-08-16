package main

import (
	"github.com/mearaj/protonet/ui"
	"os"

	"gioui.org/app"
	log "github.com/sirupsen/logrus"
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
