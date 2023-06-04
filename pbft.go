package main

import (
	"bytes"
	"encoding/json"
	"encoding/pem"
	"fmt"
	"log"
	"os"
	"strconv"
	"sync"
)

type Node struct {
	nodeID     int
	addr       string
	EccPrivkey []byte
	EccPubKey  []byte
}

type pbftNode struct {
	node         Node               //节点
	ViewNum      int                //视图编号
	LeaderID     int                //主节点编号
	sequenceID   int                //目前收到的请求序列号
	lock         sync.Mutex         //锁
	recvdmessage map[int]string     //收到的map[请求序号]消息摘要
	messagePool  map[string]Request //收到的请求消息 string为摘要
	//存放收到的prepare数量(至少需要收到并确认2f个)
	//map[摘要]map[节点序号]是否收到prepare
	prePareConfirmCount map[string]map[int]bool
	//存放收到的commit数量（至少需要收到并确认2f+1个），根据摘要来对应
	commitConfirmCount map[string]map[int]bool
	//该笔消息是否已进行Commit广播
	isCommitBordcast map[string]bool
	//该笔消息是否已对客户端进行Reply
	isReply map[string]bool
}

func NewPBFT(nodeID int, addr string) *pbftNode {

	p := new(pbftNode)
	fmt.Printf("创建新节点，地址：%s\n", addr)
	n := Node{
		nodeID:     nodeID,
		addr:       addr,
		EccPrivkey: p.getPrivKey(nodeID),
		EccPubKey:  p.getPubKey(nodeID),
	}
	p.node = n
	p.ViewNum = 0
	p.sequenceID = 0
	p.recvdmessage = make(map[int]string)
	p.messagePool = make(map[string]Request)
	p.prePareConfirmCount = make(map[string]map[int]bool)
	p.commitConfirmCount = make(map[string]map[int]bool)
	p.isCommitBordcast = make(map[string]bool)
	p.isReply = make(map[string]bool)
	return p
}

// 向除自己外的其他节点进行广播 cmd 要广播的消息类型
func (p *pbftNode) broadcast(cmd command, content interface{}) {
	//节点使用私钥对消息签名
	msgDigest := getMsgDigest(content)
	sign := EccSignature(msgDigest, p.node.EccPrivkey)

	bc, err := json.Marshal(content)
	if err != nil {
		log.Panic(err)
	}
	//组装信封
	env := Envelop{cmd, bc, sign}
	msg, _ := json.Marshal(env)
	//向除自己外的节点发送消息
	for i := range nodeTable {
		if i == p.node.nodeID {
			continue
		}
		go tcpDial(msg, nodeTable[i])
	}
}

func (p *pbftNode) handleMessage(data []byte) {

	//去空字节
	index := bytes.IndexByte(data, 0)
	//解析信封
	var env Envelop
	err := json.Unmarshal(data[:index], &env)
	if err != nil {
		log.Panic(err)
	}

	//判断类型处理
	switch command(env.Cmd) {
	case cRequest:
		p.handleClientRequest(env.Message, env.Sig)
	case cPrePrepare:
		p.handlePrePrepare(env.Message, env.Sig)
	case cPrepare:
		p.handlePrepare(env.Message, env.Sig)
	case cCommit:
		p.handleCommit(env.Message, env.Sig)
	}
}

// 传入节点编号， 获取对应的公钥
func (p *pbftNode) getPubKey(nodeID int) []byte {
	id := strconv.Itoa(nodeID)
	file, err := os.Open("Keys/" + id + "/" + id + "__PUB.pem")
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

// 传入节点编号， 获取对应的私钥
func (p *pbftNode) getPrivKey(nodeID int) []byte {
	id := strconv.Itoa(nodeID)
	file, err := os.Open("Keys/" + id + "/" + id + "__PIV.pem")
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

// 序号累加
func (p *pbftNode) sequenceIDAdd() {
	p.lock.Lock()
	p.sequenceID++
	p.lock.Unlock()
}

// 为多重映射开辟赋值
func (p *pbftNode) setPrePareConfirmMap(val string, val2 int, b bool) {
	if _, ok := p.prePareConfirmCount[val]; !ok {
		p.prePareConfirmCount[val] = make(map[int]bool)
	}
	p.prePareConfirmCount[val][val2] = b
}

// 为多重映射开辟赋值
func (p *pbftNode) setCommitConfirmMap(val string, val2 int, b bool) {
	if _, ok := p.commitConfirmCount[val]; !ok {
		p.commitConfirmCount[val] = make(map[int]bool)
	}
	p.commitConfirmCount[val][val2] = b
}
