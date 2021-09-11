package user

import (
	"math/rand"
	"net"
	"time"
)

type User struct {
	Id             int
	Addr           string
	EnterAt        time.Time
	MessageChannel chan string
}

func GenUserId() int {
	return rand.Intn(100)
}

func New(conn net.Conn) *User {
	return &User{
		Id:             GenUserId(),
		Addr:           conn.RemoteAddr().String(),
		EnterAt:        time.Now(),
		MessageChannel: make(chan string, 8),
	}
}
