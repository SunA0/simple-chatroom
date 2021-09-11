package main

import (
	"bufio"
	u "chatroom/cmd/user"
	"fmt"
	"log"
	"net"
	"strconv"
	"time"
)

var (
	enterChannel   = make(chan *u.User)
	leavingChannel = make(chan *u.User)
	msgChannel     = make(chan string, 8)
)

func main() {
	listener, err := net.Listen("tcp", ":8080")

	if err != nil {
		log.Panic(err)
	}

	go broadcaster()

	for {
		conn, err := listener.Accept()
		if err != nil {
			log.Println(err)
			continue
		}

		go handleConn(conn)
	}
}

func broadcaster() {
	users := make(map[*u.User]struct{})
	for {
		select {
		case user := <-enterChannel:
			// 新用户进入
			users[user] = struct{}{}
		case user := <-leavingChannel:
			// 用户离开
			delete(users, user)
			// 避免 goroutine 泄露
			close(user.MessageChannel)
		case msg := <-msgChannel:
			// 给所有在线用户发送消息
			for user := range users {
				user.MessageChannel <- msg
			}
		}
	}

}

func handleConn(conn net.Conn) {
	defer conn.Close()

	// create user
	user := u.New(conn)

	go sendMessage(conn, user.MessageChannel)

	// someone enter in chatroom
	user.MessageChannel <- "welcome: " + user.Addr
	msgChannel <- "user:" + strconv.Itoa(user.Id) + "has enter"
	enterChannel <- user

	// write something
	input := bufio.NewScanner(conn)
	for input.Scan() {
		msgChannel <- strconv.Itoa(user.Id) + ":" + input.Text()
	}

	if err := input.Err(); err != nil {
		log.Println("read error", err)
	}
	// do leave
	leavingChannel <- user
	msgChannel <- "user:`" + strconv.Itoa(user.Id) + "` has left"

	var userActive = make(chan struct{})
	go func() {
		d := 5 * time.Minute
		timer := time.NewTimer(d)
		for {
			select {
			case <-timer.C:
				conn.Close()
			case <-userActive:
				timer.Reset(d)
			}
		}
	}()
}

func sendMessage(conn net.Conn, ch <-chan string) {
	for msg := range ch {
		fmt.Fprintln(conn, time.Now().Format("2006-01-02 15:04:05"), "\n", msg)
	}
}
