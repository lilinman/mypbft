package main

import (
	"fmt"
	"log"
	"os"
	"strconv"
)

const nodeCount = 5

// 客户端监听地址
var clientAddr = "127.0.0.1:8888"

// 节点监听地址
var nodeTable map[int]string

func main() {
	//为四个节点生成公私钥
	//genEccKeys()
	nodeTable = map[int]string{
		0: "127.0.0.1:8000",
		1: "127.0.0.1:8001",
		2: "127.0.0.1:8002",
		3: "127.0.0.1:8003",
		4: "127.0.0.1:8004",
	}

	//处理命令行参数
	if len(os.Args) != 2 {
		log.Panic("输入参数有误")
	}
	nodeID := os.Args[1]
	if nodeID == "client" {
		c := newClient(clientAddr)
		c.clientSendMessageAndListen() //启动客户端
	} else {
		id, err := strconv.Atoi(nodeID)
		if err != nil {
			log.Panic("输入参数有误")
		} else if addr, ok := nodeTable[id]; ok {
			if id != 4 {
				p := NewPBFT(id, addr)
				fmt.Println("创建完毕 启动节点")
				go p.tcpListen() //启动节点
			} else {
				maliciousTest(id, addr)
			}

		} else {
			log.Fatal("无此节点编号!")
		}
	}
	select {}

}

func maliciousTest(id int, addr string) {
	p := NewPBFT(id, addr)
	fmt.Println("创建完毕 启动恶意节点")
	go p.tcpListen() //启动节点
	//向所有节点发送假prepare消息
	//生成prepare消息
	digest := []byte("12345")
	pre1 := Prepare{p.ViewNum, 0, string(digest), p.node.nodeID}

	fmt.Printf("正在广播虚假Prepare消息\n")
	p.broadcast(cPrepare, pre1)
	fmt.Println("Prepare广播完成")

	//发送假prePrepare消息
	r := NewRequest(clientAddr, string(digest))
	pp := PrePrepare{*r, p.ViewNum, string(digest), 0}

	p.broadcast(cPrePrepare, pp)
	fmt.Printf("正在广播虚假PrePrepare消息\n")
}
