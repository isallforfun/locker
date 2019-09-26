package main

import (
	"fmt"
	"log"
	"time"

	"github.com/fasthttp/router"
	"github.com/valyala/fasthttp"
)

func main() {

	locker := NewLockHandle()
	r := router.New()
	r.GET("/lock/*name", locker.HandleGet)
	r.DELETE("/lock/*name", locker.HandleDelete)
	r.GET("/health", func(ctx *fasthttp.RequestCtx) {
		tick := time.Tick(1 * time.Second)
		for {
			<-tick
			b := make([]byte, 1)
			if _, v := ctx.Conn().Read(b); v == nil {
				fmt.Println("Connected")
			} else {
				fmt.Println("Not Connected")
			}
		}
	})
	log.Fatal(fasthttp.ListenAndServe(":8080", r.Handler))
}
