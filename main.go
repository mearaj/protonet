package main

import (
	"gioui.org/app"
	"gioui.org/unit"
	"github.com/mearaj/protonet/view"
	log "github.com/sirupsen/logrus"
	"os"
)

func main() {
	//go func() {
	//	for {
	//		time.Sleep(time.Second * 5)
	//		jni.ShowNotification("ABC", "EFG")
	//		//log.Print("============================================")
	//		//log.Println("Number of goroutines", runtime.NumGoroutine())
	//		//runtime.Gosched()
	//		//log.Print("============================================")
	//	}
	//}()
	go func() {
		w := app.NewWindow(app.Title("ProtoNet"), app.Size(unit.Dp(400), unit.Dp(600)))
		if err := view.Loop(w); err != nil {
			log.Fatalf("exiting with error from main Loop %v\n%#v\n", err, err)
		}
		os.Exit(0)
	}()
	app.Main()
}
