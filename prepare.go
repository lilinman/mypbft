package main

import (
	"encoding/json"
	"fmt"
	"log"
)

// 处理prepare消息
func (p *pbftNode) handlePrepare(content []byte, sig Signature) {
	//使用json解析出Prepare结构体
	pre := new(Prepare)
	err := json.Unmarshal(content, pre)
	if err != nil {
		log.Panic(err)
	}
	fmt.Printf("本节点已接收到%d号节点发来的Prepare ... \n", pre.NodeID)

	//验证是否收到此消息preprepare
	_, ok := p.recvdmessage[pre.SequenceID]
	//获取消息源节点的公钥，用于数字签名验证
	MessageNodePubKey := p.getPubKey(pre.NodeID)
	//获取消息摘要
	digestByte := getMsgDigest(pre)
	if !ok || p.recvdmessage[pre.SequenceID] != pre.Digest {
		fmt.Println("没有收到过此序号此摘要消息的preprepare消息 拒绝")
	} else if !EccVerify(digestByte, sig, MessageNodePubKey) {
		fmt.Println("节点签名验证失败 拒绝执行commit广播")
	} else {
		p.setPrePareConfirmMap(pre.Digest, pre.NodeID, true)
		count := 0
		//计算消息pre数目
		for range p.prePareConfirmCount[pre.Digest] {
			count++
		}
		//主节点不会发送Prepare，所以不包含自己
		specifiedCount := 0
		if p.node.nodeID == 0 {
			specifiedCount = nodeCount / 3 * 2
		} else {
			specifiedCount = (nodeCount / 3 * 2) - 1
		}

		//如果节点至少收到了2f+1个prepare的消息（包括自己）,并且没有进行过commit广播，则进行commit广播
		p.lock.Lock()
		if count >= specifiedCount && !p.isCommitBordcast[pre.Digest] {
			fmt.Println("本节点已收到至少2f+1个节点(包括本地节点)发来的Prepare信息")
			//组装commit消息
			c := Commit{pre.View, pre.SequenceID, pre.Digest, p.node.nodeID}
			//进行提交信息的广播
			fmt.Println("正在进行commit广播")
			p.broadcast(cCommit, c)
			p.isCommitBordcast[pre.Digest] = true
			fmt.Println("commit广播完成")
		}
		p.lock.Unlock()
	}
}
