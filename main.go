package main

import (
	"github.com/fasthttp/router"
	"github.com/valyala/fasthttp"
	"log"
)

func main() {

	locker := NewLockHandle()
	r := router.New()
	r.GET("/lock/*name", locker.HandleGet)
	r.DELETE("/lock/*name", locker.HandleDelete)
	r.GET("/health", func(ctx *fasthttp.RequestCtx) {})
	log.Fatal(fasthttp.ListenAndServe(":80", r.Handler))
}
