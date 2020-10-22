package tcpserver

import (
	"bufio"
	"fmt"
	"net"
	"strings"
	"tcpspeeddating/pkg/chatroom"
	"tcpspeeddating/pkg/textcolour"
)

func Run() error {
	listen, err := net.Listen("tcp", ":8000")
	if err != nil {
		return err
	}

	for {
		conn, err := listen.Accept()
		if err != nil {
			return err
		}

		go handleConn(conn)
	}
}

func getName(conn net.Conn) string {
	_, err := conn.Write([]byte(textcolour.Green("what is your name?\n")))
	if err != nil {
		fmt.Println(err)
	}

	scanner := bufio.NewScanner(conn)
	scanner.Scan()
	name := strings.Trim(scanner.Text(), " ")
	for name == "" || !chatroom.Available(name) {
		var err error
		if !chatroom.Available(name) {
			_, err = conn.Write([]byte(textcolour.Red("this username is taken, please try again\n")))
		} else {
			_, err = conn.Write([]byte(textcolour.Red("please enter a valid name\n")))
		}
		if err != nil {
			fmt.Println(err)
		}
		scanner.Scan()
		name = strings.Trim(scanner.Text(), " ")
	}
	return name
}

func writer(conn net.Conn, messages chan string, done chan struct{}) {
	for {
		select {
		case m := <-messages:
			_, err := conn.Write(append([]byte(m), '\n'))
			if err != nil {
				fmt.Println(err)
			}
		case <-done:
			return
		}
	}
}

func handleConn(conn net.Conn) {
	msgRecv := make(chan string)
	msgSend := make(chan string)
	done := make(chan struct{})

	name := getName(conn)
	go writer(conn, msgRecv, done)

	user := chatroom.User{Name: name, In: msgSend, Out: msgRecv}
	chatroom.AddToPool(user)
	defer chatroom.Remove(user)

	scanner := bufio.NewScanner(conn)
	for scanner.Scan() {
		msgSend <- scanner.Text()
	}
	close(done)
}
