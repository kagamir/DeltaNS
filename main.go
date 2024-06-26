package main

import (
	"crypto/sha256"
	"encoding/hex"
	"flag"
	"os"

	"github.com/kagamir/DeltaNS/proxy"
	"github.com/kagamir/DeltaNS/server"
	"github.com/sirupsen/logrus"
)

func getKey(password string) []byte {
	hasher := sha256.New()
	saltData, _ := hex.DecodeString("d669d5b95563a307d544e51b1ae5b1ca")
	passwdBytes := append([]byte(password), saltData...)
	hasher.Write(passwdBytes)
	key := hasher.Sum(nil)
	return key
}

func main() {
	logrus.SetFormatter(&logrus.TextFormatter{
		TimestampFormat: "2006-01-02 15:04:05",
		FullTimestamp:   true,
	})
	logrus.SetOutput(os.Stdout)
	logrus.SetLevel(logrus.DebugLevel)

	// 定义命令行参数
	loglevel := flag.String("loglevel", "info", "日志等级")
	mode := flag.String("m", "", "启动模式，可以是 'proxy' 或 'server'")
	serverAddr := flag.String("s", "localhost:5331", "服务器的IP地址和端口号")
	localAddr := flag.String("l", "localhost:5330", "监听的IP地址和端口号")
	password := flag.String("p", "", "密码")
	dotUpstream := flag.String("u", "dns.google:853", "上游dns服务")
	timeout := flag.Int("w", 500, "等待每次回复的超时时间(毫秒)")

	// 解析命令行参数
	flag.Parse()

	if *password == "" {
		logrus.Fatalln("错误：密码参数 -p 不能为空")
		flag.Usage()
		os.Exit(1)
	}

	switch *loglevel {
	case "info":
		logrus.SetLevel(logrus.InfoLevel)
	case "debug":
		logrus.SetLevel(logrus.DebugLevel)
	case "warn":
		logrus.SetLevel(logrus.WarnLevel)
	default:
		logrus.Fatalln("loglevel in [info, debug, warn]")
		flag.Usage()
		os.Exit(1)
	}

	key := getKey(*password)

	switch *mode {
	case "proxy":
		proxyObj := proxy.Proxy{
			LocalAddr:  *localAddr,
			ServerAddr: *serverAddr,
			Key:        key,
			Timeout:    *timeout,
		}
		proxyObj.Run()
	case "server":
		server.Server(*localAddr, key, *dotUpstream)
	default:
		flag.Usage()
		os.Exit(1)
	}

}
