package main

import (
	"encoding/json"
	"fmt"
	"log"
	"strconv"
)

// 处理提交确认消息
func (p *pbftNode) handleCommit(content []byte, sig Signature) {
	//使用json解析出Commit结构体
	c := new(Commit)
	err := json.Unmarshal(content, c)
	if err != nil {
		log.Panic(err)
	}

	fmt.Printf("本节点已接收到%d节点发来的Commit ... \n", c.NodeID)

	//获取消息源节点的公钥，用于数字签名验证
	MessageNodePubKey := p.getPubKey(c.NodeID)
	digestByte := getMsgDigest(c)

	if _, ok := p.prePareConfirmCount[c.Digest]; !ok {
		fmt.Println("当前prepare池无此摘要，拒绝将信息持久化到本地消息池")
	} else if p.sequenceID != c.SequenceID {
		fmt.Println("消息序号对不上，拒绝将信息持久化到本地消息池")
	} else if !EccVerify(digestByte, sig, MessageNodePubKey) {
		fmt.Println("节点签名验证失败！,拒绝将信息持久化到本地消息池")
	} else {
		p.setCommitConfirmMap(c.Digest, c.NodeID, true)
		count := 0
		for range p.commitConfirmCount[c.Digest] {
			count++
		}
		//如果节点至少收到了2f+1个commit消息（包括自己）,并且节点没有回复过,并且已进行过commit广播，则提交信息至本地消息池，并reply成功标志至客户端！
		p.lock.Lock()
		if count >= nodeCount/3*2 && !p.isReply[c.Digest] && p.isCommitBordcast[c.Digest] {
			fmt.Println("本节点已收到至少2f + 1 个节点(包括本地节点)发来的Commit信息 ...")
			//将消息信息，提交到本地消息池中！
			//localMessagePool = append(localMessagePool, p.messagePool[c.Digest].Message)
			info := strconv.Itoa(p.node.nodeID) + "号节点已将msgid:" + strconv.Itoa(p.messagePool[c.Digest].ID) + "commit本地,消息内容为" + p.messagePool[c.Digest].Content
			fmt.Println(info)
			fmt.Println("正在reply客户端 ...")
			tcpDial([]byte(info), p.messagePool[c.Digest].ClientAddr)
			p.isReply[c.Digest] = true
			fmt.Println("reply完毕")
		}
		p.lock.Unlock()
	}
}
