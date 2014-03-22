package goftp

import (
	"errors"
	"fmt"
	"net"
	"strconv"
	"strings"
)

//定义ftp客户端命令
const (
	FCC_QUIT       string = "quit"
	FCC_VERSION    string = "version"
	FCC_DISCONNECT string = "disconnect"
	FCC_CLOSE      string = "close"
)

type GoFtpClient struct {
	Host string
	Port int

	running      bool
	ftpClientCmd GoFtpClientCmd
}

//使用初始命令行参数来连接ftp服务器
func (this *GoFtpClient) TryConnect() {
	this.running = true

	ips, lookupErr := net.LookupIP(this.Host)
	if lookupErr != nil {
		fmt.Println("goftp: Can't lookup ", this.Host)
	} else {
		var port = strconv.Itoa(this.Port)
		for _, ip := range ips {
			conn, connErr := net.Dial("tcp", net.JoinHostPort(ip.String(), port))
			if connErr != nil {
				fmt.Println("Trying ", ip, "...")
				fmt.Println("goftp:", connErr.Error())
			} else {
				fmt.Println("Connected to", ip, ".")
				this.ftpClientCmd = GoFtpClientCmd{
					FtpConn: conn,
				}
				this.ftpClientCmd.welcome()
				break
			}
		}
		this.EnterPromptMode()
	}
}

func (this *GoFtpClient) EnterPromptMode() {
	this.running = true
	var line string
	for this.running {
		fmt.Print("ftp>")
		fmt.Scanln(&line)
		if line != "" {
			this.parseCommand(line)
			err := this.executeCommand()
			if err != nil {
				fmt.Println(err.Error())
			}
			//重置line值
			line = ""
		}
	}
}

//解析交互命令
func (this *GoFtpClient) parseCommand(cmdStr string) {
	var parts = strings.Fields(cmdStr)
	if len(parts) > 0 {
		this.ftpClientCmd.Name = parts[0]
		this.ftpClientCmd.Params = parts[1:]
	}
}

//执行交互命令
func (this *GoFtpClient) executeCommand() (err error) {
	var cmdName = strings.ToLower(this.ftpClientCmd.Name)
	//var cmdParams = this.ftpClientCmd.Params
	switch cmdName {
	case FCC_QUIT:
		this.quit()
	case FCC_VERSION:
		this.version()
	case FCC_CLOSE, FCC_DISCONNECT:
		this.disconnect()
	default:
		err = errors.New("?Invalid command.")
	}

	//重置命令
	this.ftpClientCmd.Name = ""
	this.ftpClientCmd.Params = nil
	return
}

func (this *GoFtpClient) disconnect() {
	this.ftpClientCmd.disconnect()
}

func (this *GoFtpClient) quit() {
	this.running = false
	this.disconnect()
}

func (this *GoFtpClient) version() {
	fmt.Println("GoFtpClient v1.0\r\n多科学堂出品\r\nhttps://github.com/jemygraw/goftp")
}
