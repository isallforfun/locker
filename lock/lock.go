package lock

import (
	"crypto/rand"
	"fmt"
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

type RequestContext struct {
	Path   string
	HasTTL bool
	TTL    int
	Lock   bool
	Wait   bool
	Conn   net.Conn
}

const (
	ResultUnprocessableEntity = 422
	ResultSuccess             = 200
	ResultNotFound            = 404
	ResultConflict            = 409
)

func (l *LockHandle) HandleRefresh(ctx *RequestContext) int {
	path := ctx.Path
	hasTtl := ctx.HasTTL
	if !hasTtl {
		return ResultUnprocessableEntity
	}
	ttl := ctx.TTL

	l.mutex.Lock()
	_, has := l.locks[path]
	if has {
		l.refresh[path] <- ttl
	}
	l.mutex.Unlock()

	if has {
		return ResultSuccess
	}
	return ResultNotFound
}

func (l *LockHandle) HandleDelete(ctx *RequestContext) int {
	path := ctx.Path
	has := l.removeLock(path)
	if has {
		return ResultSuccess
	}
	return ResultNotFound
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

func (l *LockHandle) HandleGet(ctx *RequestContext) int {
	path := ctx.Path
	hasTtl := ctx.HasTTL
	ttl := ctx.TTL
	lock := ctx.Lock
	wait := ctx.Wait
	conn := ctx.Conn
	has, done := l.getLock(path, hasTtl, ttl, lock, conn, wait)
	if done {
		return ResultSuccess
	}
	if has {
		return ResultConflict
	}
	return ResultSuccess
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
