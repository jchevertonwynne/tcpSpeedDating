package tcpserver

import (
	"bufio"
	"net"
	"sync"
	"tcpspeeddating/pkg/chatroom"
)

var (
	connections = make(map[chan []byte]struct{})
	mu          = new(sync.Mutex)
)

func Run() error {
	listen, err := net.Listen("tcp", ":8000")
	if err != nil {
		return err
	}

	for {
		conn, err := listen.Accept()
		if err != nil {
			break
		}

		go handleConn(conn)
	}

	return nil
}

func writer(conn net.Conn, messages chan []byte, done chan struct{}) {
	for {
		select {
		case m := <-messages:
			conn.Write(append(m, '\n'))
		case <-done:
			return
		}
	}
}

func reader(conn net.Conn, msgSend chan []byte, done chan struct{}) {
	scanner := bufio.NewScanner(conn)
	for scanner.Scan() {
		msgSend <- scanner.Bytes()
	}
	close(done)
}

func handleConn(conn net.Conn) {
	msgRecv := make(chan []byte)
	msgSend := make(chan []byte)
	done := make(chan struct{})

	mu.Lock()
	connections[msgRecv] = struct{}{}
	mu.Unlock()
	defer func() {
		mu.Lock()
		delete(connections, msgRecv)
		mu.Unlock()
	}()

	go writer(conn, msgRecv, done)
	go reader(conn, msgSend, done)
	chans := chatroom.Chans{In: msgSend, Out: msgRecv}
	chatroom.AddToPool(chans)
	<-done
	chatroom.Remove(chans)
}
