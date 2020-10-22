package main

import (
	"fmt"
	"net/http"
	_ "net/http/pprof"
	"tcpspeeddating/pkg/chatroom"
	"tcpspeeddating/pkg/tcpserver"
)

func main() {
	go func() {
		err := http.ListenAndServe(":5000", nil)
		if err != nil {
			fmt.Println(err)
		}
	}()
	go chatroom.StartChat()
	tcpserver.Run()
}
