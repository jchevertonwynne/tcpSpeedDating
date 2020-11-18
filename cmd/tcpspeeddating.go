package main

import (
	"log"
	"net/http"
	_ "net/http/pprof"
	"tcpspeeddating/pkg/chatroom"
	"tcpspeeddating/pkg/tcpserver"
)

func main() {
	go func() {
		err := http.ListenAndServe(":5000", nil)
		if err != nil {
			log.Println(err)
		}
	}()
	go chatroom.StartChat()
	err := tcpserver.Run()
	if err != nil {
		log.Println(err)
	}
}
