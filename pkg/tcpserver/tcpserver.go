package tcpserver

import (
	"bufio"
	"fmt"
	"log"
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
		log.Println(err)
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
			log.Println(err)
		}
		scanner.Scan()
		name = strings.Trim(scanner.Text(), " ")
	}
	return name
}

func writer(conn net.Conn, messages chan string) chan struct{} {
	done := make(chan struct{})
	go func() {
		for {
			select {
			case msg := <-messages:
				_, err := fmt.Fprintf(conn, "%s\n", msg)
				if err != nil {
					log.Println(err)
					err := conn.Close()
					if err != nil {
						log.Println(err)
					}
					return
				}
			case <-done:
				return
			}
		}
	}()
	return done
}

func handleConn(conn net.Conn) {
	msgRecv := make(chan string)
	msgSend := make(chan string)

	name := getName(conn)
	done := writer(conn, msgRecv)

	remove := chatroom.Add(name, msgSend, msgRecv)
	defer remove()

	scanner := bufio.NewScanner(conn)
	for scanner.Scan() {
		msgSend <- scanner.Text()
	}
	close(done)
}
