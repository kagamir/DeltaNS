package proxy

import (
	"net"
	"time"

	"github.com/kagamir/DeltaNS/common"
	"github.com/miekg/dns"
	"github.com/sirupsen/logrus"
)

func handle(data []byte, key []byte, proxyConn *net.UDPConn, clientAddr *net.UDPAddr, server *net.UDPAddr) error {
	query := new(dns.Msg)
	// 尝试将字节数组解析为DNS消息。
	err := query.Unpack(data)
	if err != nil {
		logrus.Warnln("解析DNS数据出错", err)
		return err
	}

	// 查看解析出来的DNS消息
	logrus.Debugln("DNS消息:\n", query)

	// for _, question := range query.Question {
	// 	logrus.Printf("查询名称: %s, 查询类型: %d\n", question.Name, question.Qtype)
	// }

	dnsBytes, err := query.Pack()
	if err != nil {
		logrus.Errorln("解析DNS数据出错:", err)
		return err
	}

	ciphertext, err := common.Encrypt(dnsBytes, key)
	if err != nil {
		logrus.Errorln("加密出错:", err)
		return err
	}

	for i := 1; i <= 3; i++ {
		err = sentToServer(ciphertext, key, proxyConn, clientAddr, server)
		if err == nil {
			break
		}
		logrus.Infoln(err, "Retry", i)
	}
	if err != nil {
		logrus.Errorln(err)
		return err
	}

	return nil
}

func sentToServer(ciphertext []byte, key []byte, proxyConn *net.UDPConn, clientAddr *net.UDPAddr, server *net.UDPAddr) error {
	serverConn, err := getServerConn(server)
	if err != nil {
		logrus.Errorln("服务器错误:", err)
		return err
	}
	defer serverConn.Close()

	_, err = serverConn.Write(ciphertext)
	if err != nil {
		logrus.Errorln("发送错误", err)
		return err
	}

	buffer := make([]byte, 1024)
	n, _, err := serverConn.ReadFromUDP(buffer)
	if err != nil {
		if err, ok := err.(net.Error); ok && err.Timeout() {
			logrus.Errorln("读取超时，服务器没有响应", err)
		} else {
			logrus.Errorln("读取数据失败:", err)
		}
		return err
	}

	plaintext, err := common.Decrypt(buffer[:n], key)
	if err != nil {
		logrus.Errorln("解密出错:", err)
		return err
	}

	_, err = proxyConn.WriteToUDP(plaintext, clientAddr)
	if err != nil {
		logrus.Errorln("发送响应错误:", err)
		return err
	}

	return nil
}

func getServerConn(server *net.UDPAddr) (*net.UDPConn, error) {
	serverConn, err := net.DialUDP("udp", nil, server)
	if err != nil {
		return nil, err
	}

	// 设置读取超时时间
	err = serverConn.SetReadDeadline(time.Now().Add(5 * time.Second))
	if err != nil {
		logrus.Errorln("设置读取超时失败:", err)
		return nil, err
	}

	return serverConn, nil
}

func Proxy(localAddr string, serverAddr string, key []byte) {
	server, err := net.ResolveUDPAddr("udp", serverAddr)
	if err != nil {
		logrus.Fatal(err)
	}

	proxy, err := net.ResolveUDPAddr("udp", localAddr)
	if err != nil {
		logrus.Fatal(err)
	}
	proxyConn, err := net.ListenUDP("udp", proxy)
	if err != nil {
		logrus.Fatal("无法监听UDP端口:", err)
		return
	}
	defer proxyConn.Close()

	logrus.Printf("DNS代理服务器启动，监听地址：%v\n", proxy)

	for {
		var buf [512]byte

		// 读取客户端请求
		n, client, err := proxyConn.ReadFromUDP(buf[0:])
		if err != nil {
			logrus.Errorln("读取错误:", err)
			continue
		}

		go handle(buf[:n], key, proxyConn, client, server)
	}
}
