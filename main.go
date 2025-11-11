package main

import (
	"log"
	"net/http"
	_ "net/http/pprof"

	"github.com/mmulet/term.everything/termeverything"
)

//go:generate go generate ./wayland

func init() {
	go func() {
		log.Println("pprof at http://127.0.0.1:6060/debug/pprof/")
		_ = http.ListenAndServe("127.0.0.1:6060", nil)
	}()
}

func main() {
	termeverything.MainLoop()
}
