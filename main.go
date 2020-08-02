package main

import (
	"fmt"
	"github.com/getlantern/systray"
	"log"
	"mdSlideLauncher/core"
	"mdSlideLauncher/icon"
	"os"
	"path"
	"time"
)

var content *core.DocsContent
var ready chan bool

func main() {
	folder := path.Join(os.TempDir(), "md-slide", fmt.Sprint(time.Now().UnixNano()))
	ready = make(chan bool)
	if err := core.Unzip(folder); err != nil {
		log.Fatal(err)
	}
	go systray.Run(onReady, onExit)
	<-ready
	content = core.NewDocsContent(folder)
	content.Start()
}

func onReady() {
	systray.SetIcon(icon.Data)
	systray.SetTitle("md-slide")
	systray.SetTooltip("md-slide")
	mQuit := systray.AddMenuItem("Quit", "close all slides")
	go func() {
		<-mQuit.ClickedCh
		systray.Quit()
	}()
	ready <- true
}

func onExit() {
	_ = content.Srv.Close()
}
