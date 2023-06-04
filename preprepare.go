package main

import (
	"encoding/json"
	"fmt"
	"log"
)

// 处理客户端发来的请求
func (p *pbftNode) handleClientRequest(content []byte, sig Signature) {
	fmt.Println("主节点已接收到客户端发来的request ...")
	//解析request结构体
	r := new(Request)
	err := json.Unmarshal(content, r)
	if err != nil {
		log.Panic(err)
	}
	//获取request消息摘要
	digest := getDigest(*r)
	//验证客户端签名
	pubkey := p.getPubKey(8888)
	if !EccVerify(getMsgDigest(*r), sig, pubkey) {
		fmt.Println("客户端签名验证不通过，拒绝请求")
	}
	//请求序列号+1
	p.sequenceIDAdd()
	//存入消息池
	p.recvdmessage[p.sequenceID] = digest
	p.messagePool[digest] = *r

	//拼接成PrePrepare消息
	pp := PrePrepare{*r, p.ViewNum, digest, p.sequenceID}

	//进行PrePrepare广播
	fmt.Println("正在向其他节点进行进行PrePrepare广播")
	p.broadcast(cPrePrepare, pp)
	fmt.Println("PrePrepare广播完成")

}

/// 处理预准备消息

func (p *pbftNode) handlePrePrepare(content []byte, sig Signature) {
	fmt.Println("接收到主节点发来的PrePrepare消息")
	//json解析出preprepare结构体
	pp := new(PrePrepare)
	err := json.Unmarshal(content, pp)
	if err != nil {
		log.Panic(err)
	}

	//获取request消息摘要
	digest := getDigest(pp.RequestMessage)
	_, ok := p.recvdmessage[pp.SequenceID]
	//获取主节点公钥验证签名
	leaderkey := p.getPubKey(p.LeaderID)
	digestByte := getMsgDigest(pp)
	//核对消息
	if !EccVerify(digestByte, sig, leaderkey) {
		//leaderpivkey := p.getPrivKey(0)
		fmt.Println("主节点签名验证失败 拒绝接收")
	} else if digest != pp.Digest {
		fmt.Println("信息摘要错误 拒绝进行prepare广播")
	} else if ok {
		if p.recvdmessage[pp.SequenceID] != pp.Digest {
			fmt.Println("收到过序列号n相同但摘要不同的preprepare消息 拒绝接收")
		} else {
			fmt.Println("收到过相同消息 拒绝接收")
		}
	} else {
		//序号赋值
		p.sequenceID = pp.SequenceID
		//信息存入临时消息池
		p.recvdmessage[p.sequenceID] = digest
		p.messagePool[pp.Digest] = pp.RequestMessage
		//生成prepare消息
		pre := Prepare{p.ViewNum, pp.SequenceID, pp.Digest, p.node.nodeID}
		fmt.Println("正在进行Prepare广播 ...")
		p.broadcast(cPrepare, pre)
		fmt.Println("Prepare广播完成")
	}

}
