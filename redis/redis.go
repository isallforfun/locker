package redis

import (
	"bufio"
	"fmt"
	"log"
	"net"
	"os"
	"strconv"
	"strings"
	"unicode/utf8"

	"github.com/isallforfun/locker/lock"
	"github.com/secmask/go-redisproto"
)

func handleConnection(locker *lock.LockHandle, conn net.Conn) {
	defer conn.Close()
	parser := redisproto.NewParser(conn)
	writer := redisproto.NewWriter(bufio.NewWriter(conn))
	var ew error
	for {
		command, err := parser.ReadCommand()

		if err != nil {
			_, ok := err.(*redisproto.ProtocolError)
			if ok {
				ew = writer.WriteError(err.Error())
			} else {
				break
			}
		} else {
			cmd := strings.ToUpper(string(command.Get(0)))
			var ctx, err = buildContext(conn, command)
			if err != nil {
				ew = writer.WriteError(err.Error())
			} else {
				switch cmd {
				case "GET":
					var result = locker.HandleGet(ctx)
					var data = strconv.Itoa(result)
					ew = writer.WriteBulkString(data)
				case "DELETE":
					var result = locker.HandleDelete(ctx)
					var data = strconv.Itoa(result)
					ew = writer.WriteBulkString(data)
				case "REFRESH":
					var result = locker.HandleRefresh(ctx)
					var data = strconv.Itoa(result)
					ew = writer.WriteBulkString(data)
				default:
					ew = writer.WriteError("Command not support")
				}
			}
		}
		if command.IsLast() {
			writer.Flush()
		}
		if ew != nil {
			//log.Println("Connection closed", ew)
			break
		}
	}
}

func buildContext(conn net.Conn, command *redisproto.Command) (*lock.RequestContext, error) {
	if command.ArgCount() < 2 {
		return nil, fmt.Errorf("no path")
	}
	var path = string(command.Get(1))
	var hasTTL = command.ArgCount() >= 3
	var ttl = 0
	var err error
	if hasTTL {
		ttl, err = strconv.Atoi(string(command.Get(2)))
		if err != nil {
			return nil, err
		}
	}
	var hasLock = false
	if command.ArgCount() > 4 && string(command.Get(3)) == "1" {
		hasLock = true
	}
	var hasWait = false
	if command.ArgCount() > 5 && string(command.Get(4)) == "1" {
		hasWait = true
	}
	return &lock.RequestContext{
		Path:   path,
		HasTTL: hasTTL,
		TTL:    ttl,
		Lock:   hasLock,
		Wait:   hasWait,
		Conn:   conn,
	}, nil
}

func Run(locker *lock.LockHandle) {
	listener, err := net.Listen("tcp", getRedisBindPort())
	if err != nil {
		panic(err)
	}
	for {
		conn, err := listener.Accept()
		if err != nil {
			log.Println("Error on accept: ", err)
			continue
		}
		go handleConnection(locker, conn)
	}
}

func getRedisBindPort() string {
	var result = os.Getenv("LOCKER_REDIS_BIND")
	if utf8.RuneCountInString(result) == 0 {
		return ":6379"
	}
	return result
}
