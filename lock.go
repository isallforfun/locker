package main

import (
	"crypto/rand"
	"fmt"
	"github.com/valyala/fasthttp"
	"net"
	"sync"
	"time"
)

type LockHandle struct {
	locks  map[string]struct{}
	mutex  *sync.Mutex
	unlock map[string]chan struct{}
	close  map[string]chan struct{}
}

func (l *LockHandle) HandleDelete(ctx *fasthttp.RequestCtx) {
	path := string(ctx.Request.URI().Path())
	l.mutex.Lock()
	_, has := l.locks[path]
	if has {
		l.deleteLock(path)
	}
	l.mutex.Unlock()
	if has {
		ctx.SetStatusCode(200)
	} else {
		ctx.SetStatusCode(404)
	}
}

func (l *LockHandle) deleteLock(path string) {
	delete(l.locks, path)
	l.close[path] <- struct{}{}
	close(l.close[path])
	delete(l.close, path)
	l.unlock[path] <- struct{}{}
	close(l.unlock[path])
	delete(l.unlock, path)
}

func (l *LockHandle) HandleGet(ctx *fasthttp.RequestCtx) {
	path := string(ctx.Request.URI().Path())
RETRY:
	l.mutex.Lock()
	_, has := l.locks[path]
	if !has {
		l.locks[path] = struct{}{}
		l.unlock[path] = make(chan struct{}, 1)
		l.close[path] = make(chan struct{}, 1)
		l.mutex.Unlock()
		if ctx.Request.URI().QueryArgs().Has("ttl") {
			go l.createTTL(path, ctx)

		} else if ctx.Request.URI().QueryArgs().Has("lock") {
			connectionClose := checkConnectionClose(ctx)
			select {
			case <-l.close[path]:
				ctx.Conn().Close()
				break
			case <-connectionClose:
				l.mutex.Lock()
				l.deleteLock(path)
				l.mutex.Unlock()
				break
			}
		}
	} else {
		l.mutex.Unlock()
		if ctx.Request.URI().QueryArgs().Has("wait") {
			connectionClose := checkConnectionClose(ctx)
			select {
			case <-l.unlock[path]:
				goto RETRY
			case <-connectionClose:
				return
			}
		}
	}
	if has {
		ctx.SetStatusCode(409)
	} else {
		ctx.SetStatusCode(200)
	}

}

func (l *LockHandle) createTTL(path string, ctx *fasthttp.RequestCtx) {
	func(durationInt int) {
		duration := time.Duration(durationInt) * time.Millisecond
		select {
		case <-l.close[path]:
			break
		case <-time.After(duration):
			l.mutex.Lock()
			l.deleteLock(path)
			l.mutex.Unlock()
			break
		}
	}(ctx.Request.URI().QueryArgs().GetUintOrZero("ttl"))
}

func checkConnectionClose(ctx *fasthttp.RequestCtx) chan struct{} {
	connectionClose := make(chan struct{})
	go func(conn net.Conn) {
		b := make([]byte, 1)
		for {
			if _, v := conn.Read(b); v != nil {
				connectionClose <- struct{}{}
				break
			}
		}
	}(ctx.Conn())
	return connectionClose
}

func (l *LockHandle) getNewHash() string {
	b := make([]byte, 8)
	rand.Read(b)
	return fmt.Sprintf("%x", b)
}

func NewLockHandle() *LockHandle {
	return &LockHandle{
		mutex:  &sync.Mutex{},
		locks:  make(map[string]struct{}),
		unlock: make(map[string]chan struct{}),
		close:  make(map[string]chan struct{}),
	}
}
