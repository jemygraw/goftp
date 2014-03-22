package main

import (
	"fmt"
	"goftp"
	"os"
	"strconv"
)

//ftp服务器默认监听端口号
const (
	FTP_SERVER_DEFAULT_LISTENING_PORT = 21
)

func help() {
	fmt.Println("usage: ftp [host-name] [port]")
}

func main() {
	var ftpServerHost string
	var ftpServerPort int
	//获取命令行参数切片(不包括命令名称)
	var progArgs = os.Args[1:]
	var progArgCount = len(progArgs)
	//检查命令行参数
	/*
	  ftp命令有以下几种调用方式
	  1. ftp
	     直接进入ftp交互式命令界面
	  2. ftp hostname
	     尝试以hostname所指定的主机名，默认ftp服务器的端口21来连接ftp服务器；
	     连接成功或失败后进入ftp交互式命令界面
	  3. ftp hostname port
	     尝试以hostname所指定的主机名，port所指定的ftp服务器监听端口来连接
	     ftp服务器，连接成功或失败后进入ftp交互式命令界面
	*/
	switch progArgCount {
	case 0:
		ftpServerHost = ""
		ftpServerPort = FTP_SERVER_DEFAULT_LISTENING_PORT
	case 1:
		ftpServerHost = progArgs[0]
		ftpServerPort = FTP_SERVER_DEFAULT_LISTENING_PORT
	case 2:
		ftpServerHost = progArgs[0]
		port, err := strconv.Atoi(progArgs[1])
		if err != nil {
			ftpServerPort = port
		} else {
			ftpServerPort = FTP_SERVER_DEFAULT_LISTENING_PORT
		}
	default:
		help()
	}

	var ftpClient = goftp.GoFtpClient{
		Host: ftpServerHost,
		Port: ftpServerPort,
	}
	if ftpClient.Host != "" {
		ftpClient.TryConnect()
	} else {
		ftpClient.EnterPromptMode()
	}
}
