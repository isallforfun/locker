package http

import (
	"log"
	"os"
	"unicode/utf8"

	"github.com/fasthttp/router"
	"github.com/isallforfun/locker/lock"
	"github.com/valyala/fasthttp"
)

func buildLockerContext(ctx *fasthttp.RequestCtx) *lock.RequestContext {
	var lockerContext = &lock.RequestContext{
		Path:   string(ctx.Request.URI().Path()),
		HasTTL: ctx.Request.URI().QueryArgs().Has("ttl"),
		TTL:    ctx.Request.URI().QueryArgs().GetUintOrZero("ttl"),
		Lock:   ctx.Request.URI().QueryArgs().Has("lock"),
		Wait:   ctx.Request.URI().QueryArgs().Has("wait"),
		Conn:   ctx.Conn(),
	}
	return lockerContext
}

func Run() {
	locker := lock.NewLockHandle()
	r := router.New()
	r.GET("/lock/*name", func(ctx *fasthttp.RequestCtx) {
		ctx.SetStatusCode(locker.HandleRefresh(buildLockerContext(ctx)))
	})
	r.DELETE("/lock/*name", func(ctx *fasthttp.RequestCtx) {
		ctx.SetStatusCode(locker.HandleDelete(buildLockerContext(ctx)))
	})
	r.PATCH("/lock/*name", func(ctx *fasthttp.RequestCtx) {
		ctx.SetStatusCode(locker.HandleRefresh(buildLockerContext(ctx)))
	})
	r.GET("/health", func(ctx *fasthttp.RequestCtx) {})
	log.Fatal(fasthttp.ListenAndServe(getHTTPBindPort(), r.Handler))
}

func getHTTPBindPort() string {
	var result = os.Getenv("LOCKER_HTTP_BIND")
	if utf8.RuneCountInString(result) == 0 {
		return ":80"
	}
	return result
}
