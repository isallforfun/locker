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
	locks   map[string]struct{}
	mutex   *sync.Mutex
	unlock  map[string]chan struct{}
	close   map[string]chan struct{}
	refresh map[string]chan int
}

func (l *LockHandle) HandleRefresh(ctx *fasthttp.RequestCtx) {
	path := string(ctx.Request.URI().Path())
	hasTtl := ctx.Request.URI().QueryArgs().Has("ttl")
	if !hasTtl {
		ctx.SetStatusCode(422)
		return
	}
	ttl := ctx.Request.URI().QueryArgs().GetUintOrZero("ttl")

	l.mutex.Lock()
	_, has := l.locks[path]
	if has {
		l.refresh[path] <- ttl
	}
	l.mutex.Unlock()

	if has {
		ctx.SetStatusCode(200)
	} else {
		ctx.SetStatusCode(404)
	}
}

func (l *LockHandle) HandleDelete(ctx *fasthttp.RequestCtx) {
	path := string(ctx.Request.URI().Path())
	has := l.removeLock(path)
	if has {
		ctx.SetStatusCode(200)
	} else {
		ctx.SetStatusCode(404)
	}
}

func (l *LockHandle) removeLock(path string) bool {
	l.mutex.Lock()
	_, has := l.locks[path]
	if has {
		l.deleteLock(path)
	}
	l.mutex.Unlock()
	return has
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
	hasTtl := ctx.Request.URI().QueryArgs().Has("ttl")
	ttl := ctx.Request.URI().QueryArgs().GetUintOrZero("ttl")
	lock := ctx.Request.URI().QueryArgs().Has("lock")
	wait := ctx.Request.URI().QueryArgs().Has("wait")
	conn := ctx.Conn()
	has, done := l.getLock(path, hasTtl, ttl, lock, conn, wait)
	if done {
		return
	}
	if has {
		ctx.SetStatusCode(409)
	} else {
		ctx.SetStatusCode(200)
	}

}

func (l *LockHandle) getLock(path string, hasTtl bool, ttl int, lock bool, conn net.Conn, wait bool) (bool, bool) {
RETRY:
	l.mutex.Lock()
	_, has := l.locks[path]
	if !has {
		l.locks[path] = struct{}{}
		l.unlock[path] = make(chan struct{}, 1)
		l.close[path] = make(chan struct{}, 1)
		l.refresh[path] = make(chan int, 1)
		l.mutex.Unlock()
		if hasTtl {
			go l.createTTL(path, ttl)
		} else if lock {
			l.waitConnectionRelease(conn, path)
		}
	} else {
		l.mutex.Unlock()
		if wait {
			connectionClose := checkConnectionClose(conn)
			select {
			case <-l.unlock[path]:
				goto RETRY
			case <-connectionClose:
				return false, true
			}
		}
	}
	return has, false
}

func (l *LockHandle) waitConnectionRelease(conn net.Conn, path string) {
	connectionClose := checkConnectionClose(conn)
	select {
	case <-l.close[path]:
		conn.Close()
		break
	case <-connectionClose:
		l.mutex.Lock()
		l.deleteLock(path)
		l.mutex.Unlock()
		break
	}
}

func (l *LockHandle) createTTL(path string, ttl int) {
	func(durationInt int) {
	REFRESH:
		duration := time.Duration(durationInt) * time.Millisecond
		select {
		case <-l.close[path]:
			break
		case durationInt = <-l.refresh[path]:
			goto REFRESH
		case <-time.After(duration):
			l.mutex.Lock()
			l.deleteLock(path)
			l.mutex.Unlock()
			break
		}
	}(ttl)
}

func checkConnectionClose(ctx net.Conn) chan struct{} {
	connectionClose := make(chan struct{})
	go func(conn net.Conn) {
		b := make([]byte, 1)
		for {
			if _, v := conn.Read(b); v != nil {
				connectionClose <- struct{}{}
				break
			}
		}
	}(ctx)
	return connectionClose
}

func (l *LockHandle) getNewHash() string {
	b := make([]byte, 8)
	rand.Read(b)
	return fmt.Sprintf("%x", b)
}

func NewLockHandle() *LockHandle {
	return &LockHandle{
		mutex:   &sync.Mutex{},
		locks:   make(map[string]struct{}),
		unlock:  make(map[string]chan struct{}),
		close:   make(map[string]chan struct{}),
		refresh: make(map[string]chan int),
	}
}
