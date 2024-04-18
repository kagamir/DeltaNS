package server

import (
	"crypto/tls"
	"net"

	"github.com/kagamir/DeltaNS/common"
	"github.com/miekg/dns"
	"github.com/sirupsen/logrus"
)

func queryRemote(msg *dns.Msg, dotUpstream string) (*dns.Msg, error) {
	// 初始化DNS over TLS连接
	conn, err := dns.DialWithTLS("tcp", dotUpstream, &tls.Config{})
	if err != nil {
		return nil, err
	}
	defer conn.Close()

	// 发送DNS查询
	err = conn.WriteMsg(msg)
	if err != nil {
		return nil, err
	}

	// 接收DNS响应
	r, err := conn.ReadMsg()
	if err != nil {
		return nil, err
	}

	return r, nil
}

func handle(data []byte, key []byte, dotUpstream string, serverConn *net.UDPConn, proxyAddr *net.UDPAddr) error {
	plaintext, err := common.Decrypt(data, key)
	if err != nil {
		logrus.Errorln("Decrypt Err:", err)
		return err
	}

	query := new(dns.Msg)
	// 尝试将字节数组解析为DNS消息。
	err = query.Unpack(plaintext)
	if err != nil {
		logrus.Errorln("解析DNS数据出错:", err)
		return err
	}

	logrus.Debugln("DNS消息:\n", query)

	respMsg, err := queryRemote(query, dotUpstream)
	if err != nil {
		// 返回查询错误
		logrus.Errorln("查询错误", err)
		errResp := new(dns.Msg)
		errResp.SetRcode(query, dns.RcodeServerFailure)
		respMsg = errResp
	}

	logrus.Debugln("DNS返回:\n", respMsg)

	respData, err := respMsg.Pack()
	if err != nil {
		logrus.Errorln("打包响应数据失败:", err)
		return err
	}

	ciphertext, err := common.Encrypt(respData, key)
	if err != nil {
		logrus.Errorln("加密DNS数据出错:", err)
		return err
	}

	sendLen, err := serverConn.WriteToUDP(ciphertext, proxyAddr)
	if err != nil {
		logrus.Errorln("发送响应错误:", err)
		return err
	}
	logrus.Debugln("返回", proxyAddr, "LENGTH =", sendLen)

	return nil

}

func Server(serverAddr string, key []byte, dotUpstream string) {
	server, err := net.ResolveUDPAddr("udp", serverAddr)
	if err != nil {
		logrus.Fatal(err)
	}
	serverConn, err := net.ListenUDP("udp", server)
	if err != nil {
		logrus.Fatal("无法监听UDP端口:", err)
		return
	}
	defer serverConn.Close()

	logrus.Infoln("DNS代理服务器启动，监听地址:", server)

	for {
		var buf [1024]byte

		n, proxyAddr, err := serverConn.ReadFromUDP(buf[0:])
		if err != nil {
			logrus.Errorln("读取错误:", err)
			continue
		}
		logrus.Debugln("Get data len", n, "from", proxyAddr)
		go handle(buf[:n], key, dotUpstream, serverConn, proxyAddr)
	}

}
