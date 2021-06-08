package core

import (
	"encoding/json"
	"log"
	"net"
	"net/http"
	"os"
	"path"
)

type DocServer struct {
	server *http.Server
	addr   net.Addr
	path   string
	meta   Meta
}

func newDocServer(folder string) (*DocServer, error) {
	meta, err := getMeta(path.Join(folder, "meta.json"))
	if err != nil {
		return nil, err
	}
	l, err := net.Listen("tcp", "localhost:0")
	if err != nil {
		return nil, err
	}
	mux := http.NewServeMux()
	mux.Handle("/", http.StripPrefix("/", http.FileServer(http.Dir(folder))))
	srv := &http.Server{
		Addr:              l.Addr().String(),
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
	run := func() {
		log.Println(srv.Serve(l))
	}
	go run()
	return &DocServer{
		server: srv,
		addr:   l.Addr(),
		path:   folder,
		meta:   *meta,
	}, nil
}

func (server *DocServer) Close() {
	_ = server.server.Close()
}

type Meta struct {
	Title  string `json:"title"`
	Author string `json:"author"`
	Uuid   string `json:"uuid"`
}

func getMeta(file string) (*Meta, error) {
	fs, err := os.Open(file)
	if err != nil {
		return nil, err
	}
	defer fs.Close()
	dec := json.NewDecoder(fs)
	var data Meta
	if err := dec.Decode(&data); err != nil {
		return nil, err
	}
	return &data, nil
}
