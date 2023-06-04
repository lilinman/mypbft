package main

import (
	"bufio"
	"encoding/json"
	"encoding/pem"
	"fmt"
	"log"
	"net"
	"os"
)

type Client struct {
	ClientAddr string //客户端地址
	ID         int    //客户端编号
	EccPrivkey []byte
	EccPubKey  []byte
}

func newClient(addr string) *Client {
	c := new(Client)
	c.ClientAddr = addr
	c.ID = 8888 //只有一个client 默认8888了、、
	c.EccPrivkey = getClientPrivKey(addr)
	c.EccPubKey = getClientPubKey(addr)
	return c
}

// 客户端监听
func (c *Client) clientTcpListen() {
	listen, err := net.Listen("tcp", c.ClientAddr)
	if err != nil {
		log.Panic(err)
	}
	defer listen.Close()

	//读取reply信息
	for {
		conn, err := listen.Accept()
		if err != nil {
			log.Panic(err)
		}
		buf := make([]byte, 1024)
		n, _ := conn.Read(buf)
		fmt.Println(string(buf[:n]))
	}
}

// 开启客户端本地监听
func (c *Client) clientSendMessageAndListen() {
	go c.clientTcpListen()
	fmt.Println("客户端开启监听，地址", c.ClientAddr)

	fmt.Println(" ---------------------------------------------------------------------------------")
	fmt.Println("请在下方输入要存入节点的信息??")

	//获取用户输入
	stdReader := bufio.NewReader(os.Stdin)

	for {
		data, err := stdReader.ReadString('\n')
		if err != nil {
			fmt.Println("Error reading from stdin")
			panic(err)
		}
		r := NewRequest(clientAddr, data)

		//将request转化为json字符串输出
		br, err := json.Marshal(r)
		if err != nil {
			log.Panic(err)
		}
		fmt.Println("请求消息为", string(br))

		//对消息签名
		msgDigest := getMsgDigest(r)
		sign := EccSignature(msgDigest, c.EccPrivkey)
		//默认NO为主节点，把请求消息发送给N0
		env := Envelop{cRequest, br, sign}
		msg, _ := json.Marshal(env)
		tcpDial(msg, nodeTable[0])
	}
}
func getClientPrivKey(addr string) []byte {
	file, err := os.Open("Keys/" + "8888" + "/" + "8888" + "__PIV.pem")
	if err != nil {
		panic(err)
	}
	info, err := file.Stat()
	if err != nil {
		panic(err)
	}
	buf := make([]byte, info.Size())
	file.Read(buf)
	file.Close()
	block, _ := pem.Decode(buf)
	return block.Bytes
}

func getClientPubKey(addr string) []byte {
	file, err := os.Open("Keys/" + "8888" + "/" + "8888" + "__PUB.pem")
	if err != nil {
		panic(err)
	}
	info, err := file.Stat()
	if err != nil {
		panic(err)
	}
	buf := make([]byte, info.Size())
	file.Read(buf)
	file.Close()
	block, _ := pem.Decode(buf)
	return block.Bytes
}
