package core

import (
	"container/list"
	"github.com/darkowlzz/openurl"
	"github.com/getlantern/systray"
	"log"
	"net/http"
	"net/rpc"
	"os"
)

type DocsContent struct {
	rpcQuit chan bool
	Srv     http.Server
	Docs    *list.List
}

func NewDocsContent(path string) *DocsContent {
	client, err := rpc.DialHTTP("tcp", "localhost:4056")
	result := &DocsContent{
		rpcQuit: nil,
		Docs:    list.New(),
	}
	if err == nil {
		if err := client.Call("LoadDoc", path, nil); err != nil {
			log.Fatal(err)
		}
	} else {
		result.rpcQuit = make(chan bool)
		server := rpc.NewServer()
		if err := server.Register(result); err != nil {
			log.Fatal(err)
		}
		mux := http.NewServeMux()
		mux.HandleFunc("/", server.ServeHTTP)
		result.Srv = http.Server{
			Addr:              "localhost:4056",
			Handler:           mux,
			TLSConfig:         nil,
			ReadTimeout:       0,
			ReadHeaderTimeout: 0,
			WriteTimeout:      0,
			IdleTimeout:       0,
			MaxHeaderBytes:    0,
			TLSNextProto:      nil,
			ConnState:         nil,
			ErrorLog:          nil,
		}
		runServer := func() {
			log.Println(result.Srv.ListenAndServe())
			result.rpcQuit <- true
		}
		go runServer()
		_ = result.LoadDoc(path, nil)
	}
	return result
}

func (content *DocsContent) Start() {
	<-content.rpcQuit
	for e := content.Docs.Front(); e != nil; e = e.Next() {
		e.Value.(*DocServer).Close()
		_ = os.RemoveAll(e.Value.(*DocServer).path)
	}
}

func (content *DocsContent) LoadDoc(folder string, result *bool) error {
	s, err := newDocServer(folder)
	if err != nil {
		return err
	}
	e := content.Docs.PushBack(s)
	item := systray.AddMenuItem(s.meta.Title, "")
	displayBtn := item.AddSubMenuItem("Display", "")
	go func() {
		for {
			<-displayBtn.ClickedCh
			_ = openurl.Open("http://" + s.addr.String() + "/slide/" + s.meta.Uuid + "/index.html")
		}
	}()
	speakerBtn := item.AddSubMenuItem("Speaker", "")
	go func() {
		for {
			<-speakerBtn.ClickedCh
			_ = openurl.Open("http://" + s.addr.String() + "/slide/" + s.meta.Uuid + "/index.html?mode=speaker")
		}
	}()
	rawBtn := item.AddSubMenuItem("Raw", "")
	go func() {
		for {
			<-rawBtn.ClickedCh
			_ = openurl.Open("http://" + s.addr.String() + "/slide/" + s.meta.Uuid + "/index.md")
		}
	}()
	closeBtn := item.AddSubMenuItem("Close", "")
	go func() {
		<-closeBtn.ClickedCh
		item.Hide()
		content.Docs.Remove(e)
		e.Value.(*DocServer).Close()
		if content.Docs.Len() == 0 {
			systray.Quit()
		}
	}()
	return nil
}
