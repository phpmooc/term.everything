//go:build profile
// +build profile

package termeverything

import (
	_ "net/http/pprof"

	"log"
	"net/http"
)

func init() {
	go func() {
		log.Println("pprof at http://127.0.0.1:6060/debug/pprof/")
		_ = http.ListenAndServe("127.0.0.1:6060", nil)
	}()
}
