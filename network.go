package main

import (
	"fmt"
	"log"
	"net"
)

// 发消息
func tcpDial(context []byte, addr string) {
	conn, err := net.Dial("tcp", addr)
	if err != nil {
		log.Println("connect error", err)
		return
	}
	_, err = conn.Write(context)
	if err != nil {
		log.Fatal(err)
	}
	conn.Close()
}

// 节点使用的tcp监听
func (p *pbftNode) tcpListen() {
	listen, err := net.Listen("tcp", p.node.addr)
	if err != nil {
		log.Panic(err)
	}
	fmt.Printf("节点开启监听，地址：%s\n", p.node.addr)
	defer listen.Close()

	for {
		conn, err := listen.Accept()

		if err != nil {
			log.Panic(err)
		}
		buf := make([]byte, 1024)
		conn.Read(buf)

		p.handleMessage(buf)
	}

}
