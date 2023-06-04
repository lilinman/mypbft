package main

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/sha256"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"log"
	"math/big"
	"os"
	"strconv"
)

type Signature struct {
	Rtext []byte
	Stext []byte
}

// 数字签名
func EccSignature(message []byte, privkey []byte) Signature {

	//x509还原
	prk, err := x509.ParseECPrivateKey(privkey)
	if err != nil {
		log.Fatal("密钥解析错误")
	}
	//消息哈希
	h := sha256.New()
	h.Write(message)
	hashText := h.Sum(nil)
	r, s, err := ecdsa.Sign(rand.Reader, prk, hashText[:])
	if err != nil {
		log.Panic("签名错误")
	}

	//将r和s序列化
	rt, err := r.MarshalText()
	if err != nil {
		log.Panic("转化rt错误")
	}
	st, err := s.MarshalText()
	if err != nil {
		log.Panic("转化st错误")
	}

	return Signature{rt, st}

}

// 签名验证
func EccVerify(message []byte, sig Signature, pubKey []byte) bool {

	pubInterface, err := x509.ParsePKIXPublicKey(pubKey)
	if err != nil {
		log.Fatal("公钥解析错误")
	}
	// 进行类型断言——得到公钥结构体
	pk := pubInterface.(*ecdsa.PublicKey)

	//将r，t转化成int
	var r, s big.Int
	r.UnmarshalText(sig.Rtext)
	s.UnmarshalText(sig.Stext)

	//取消息hash
	h := sha256.New()
	h.Write(message)
	hashText := h.Sum(nil)

	// 签名验证
	bl := ecdsa.Verify(pk, hashText[:], &r, &s)
	//fmt.Println(bl)
	return bl
}

// 生成密钥对{
func genEccKeys() {
	fmt.Println("生成公私钥 ...")
	err := os.Mkdir("Keys", 0644)
	if err != nil {
		log.Panic()
	}
	for i := 0; i <= 4; i++ {
		err := os.Mkdir("./Keys/"+strconv.Itoa(i), 0644)
		if err != nil {
			log.Panic()
		}
		priv, pub := getKeyPair()
		privFileName := "Keys/" + strconv.Itoa(i) + "/" + strconv.Itoa(i) + "__PIV.pem"
		file, err := os.Create(privFileName)
		if err != nil {
			log.Panic(err)
		}
		defer file.Close()
		file.Write(priv)

		pubFileName := "Keys/" + strconv.Itoa(i) + "/" + strconv.Itoa(i) + "__PUB.pem"
		file2, err := os.Create(pubFileName)
		if err != nil {
			log.Panic(err)
		}
		defer file2.Close()
		file2.Write(pub)
	}
	fmt.Println("已为节点们生成RSA公私钥")

	//生成客户端公私钥
	err = os.Mkdir("./Keys/"+"8888", 0644)
	if err != nil {
		log.Panic()
	}
	priv, pub := getKeyPair()
	privFileName := "Keys/" + "8888" + "/" + "8888" + "__PIV.pem"
	file, err := os.Create(privFileName)
	if err != nil {
		log.Panic(err)
	}
	defer file.Close()
	file.Write(priv)

	pubFileName := "Keys/" + "8888" + "/" + "8888" + "__PUB.pem"
	file2, err := os.Create(pubFileName)
	if err != nil {
		log.Panic(err)
	}
	defer file2.Close()
	file2.Write(pub)
	fmt.Println("已为客户端生成RSA公私钥")
}

// 生成密钥对
func getKeyPair() (prvkey, pubkey []byte) {
	privateKey, err := ecdsa.GenerateKey(elliptic.P224(), rand.Reader)
	if err != nil {
		panic(err)
	}
	derText, err := x509.MarshalECPrivateKey(privateKey)
	if err != nil {
		panic(err)
	}
	block := pem.Block{
		Type:  "ecdsa private key",
		Bytes: derText,
	}

	//pem编码
	prvkey = pem.EncodeToMemory(&block)

	publicKey := privateKey.PublicKey
	derText, err = x509.MarshalPKIXPublicKey(&publicKey)
	if err != nil {
		panic(err)
	}
	//得到的切片字符串放入pemBlock
	block = pem.Block{
		Type:  "ecdsa public key",
		Bytes: derText,
	}
	pubkey = pem.EncodeToMemory(&block)

	return
}
