package main

import (
	"github.com/isallforfun/locker/http"
	"github.com/isallforfun/locker/lock"
	"github.com/isallforfun/locker/redis"
)

func main() {
	locker := lock.NewLockHandle()
	go http.Run(locker)
	go redis.Run(locker)
	select {}
}
