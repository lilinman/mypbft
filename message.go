package main

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"log"
	"math/big"
	"strings"
	"time"
)

// 消息类型前缀长度固定为12

type command string

const (
	cRequest    command = "request"
	cPrePrepare command = "preprepare"
	cPrepare    command = "prepare"
	cCommit     command = "commit"
)

// 消息传递信封
type Envelop struct {
	Cmd     command         //消息类型
	Message json.RawMessage //消息json
	Sig     Signature       //消息签名
}

// 客户端消息
type Message struct {
	Content string
	ID      int
}

// 客户端request消息
type Request struct {
	Message
	TimeStamp  int64
	ClientAddr string
}

// <<Pre-Prepare,v,n,d>,m>
// v:视图编号
// n:请求的编号
// d:消息摘要
// m:消息
type PrePrepare struct {
	RequestMessage Request //请求消息
	View           int     //视图编号
	Digest         string  //消息摘要
	SequenceID     int     //请求编号
}

// PrePare<v,n,d,i>
type Prepare struct {
	View       int    //视图编号
	SequenceID int    //请求编号
	Digest     string //消息摘要
	NodeID     int    //节点编号
}

// commit v,n,d,i
type Commit struct {
	View       int    //视图编号
	SequenceID int    //请求编号
	Digest     string //消息摘要
	NodeID     int    //节点编号
}

func NewRequest(clientAddr, data string) *Request {
	d := getRandom()
	m := Message{
		Content: strings.TrimSpace(data),
		ID:      d,
	}
	r := &Request{
		Message:    m,
		TimeStamp:  time.Now().UnixNano(),
		ClientAddr: clientAddr,
	}
	return r
}

// 返回一个十位数的随机数，作为msgid
func getRandom() int {
	x := big.NewInt(10000000000)
	for {
		result, err := rand.Int(rand.Reader, x)
		if err != nil {
			log.Panic(err)
		}
		if result.Int64() > 1000000000 {
			return int(result.Int64())
		}
	}
}

// 默认前十二位为命令名称,整合消息

// 对客户端消息详情进行摘要
func getDigest(request Request) string {
	b, err := json.Marshal(request)
	if err != nil {
		log.Panic(err)
	}
	hash := sha256.Sum256(b)
	//进行十六进制字符串编码
	return hex.EncodeToString(hash[:])
}

// 对Msg进行摘要
func getMsgDigest(msg interface{}) []byte {
	b, err := json.Marshal(msg)
	if err != nil {
		log.Panic(err)
	}
	hash := sha256.Sum256(b)
	//进行十六进制字符串编码
	return hash[:]
}
