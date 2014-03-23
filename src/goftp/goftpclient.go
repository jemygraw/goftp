package goftp

import (
	"bufio"
	"errors"
	"fmt"
	"net"
	"os"
	"strconv"
	"strings"
	"time"
)

//定义ftp客户端支持的用户命令，FCC表示Ftp Client Command
const (
	FCC_QUESTION_MARK string = "?"
	FCC_HELP          string = "help"
	FCC_BYE           string = "bye"
	FCC_QUIT          string = "quit"
	FCC_VERSION       string = "version"
	FCC_DISCONNECT    string = "disconnect"
	FCC_CLOSE         string = "close"
	FCC_PWD           string = "pwd"
)

const (
	DIAL_FTP_SERVER_TIMEOUT_SECONDS int = 30 //连接ftp服务器的超时时间
)

//表示ftp客户端的结构体
type GoFtpClient struct {
	Host string //ftp服务器主机名
	Port int    //ftp服务器监听端口号

	running      bool           //表示ftp客户端是否处于运行中的flag
	ftpClientCmd GoFtpClientCmd //组合的ftp客户端命令结构体

	GoFtpClientHelp //组合的ftp帮助结构体
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
			conn, connErr := net.DialTimeout("tcp", net.JoinHostPort(ip.String(), port),
				time.Duration(DIAL_FTP_SERVER_TIMEOUT_SECONDS)*time.Second)
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

//进入命令交互模式
func (this *GoFtpClient) EnterPromptMode() {
	this.running = true
	for this.running {
		fmt.Print("ftp>")
		cmdReader := bufio.NewReader(os.Stdin)
		cmdStr, err := cmdReader.ReadString('\n')
		if err == nil && cmdStr != "" {
			this.parseCommand(cmdStr)
			err := this.executeCommand()
			if err != nil {
				fmt.Println(err.Error())
			}
			//重置cmdStr值
			cmdStr = ""
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
	var cmdParams = this.ftpClientCmd.Params
	switch cmdName {
	case FCC_QUIT, FCC_BYE:
		this.quit()
	case FCC_VERSION:
		this.version()
	case FCC_CLOSE, FCC_DISCONNECT:
		this.disconnect()
	case FCC_PWD:
		this.pwd()
	case FCC_QUESTION_MARK, FCC_HELP:
		if len(cmdParams) > 0 {
			this.cmdHelp(cmdParams)
		} else {
			this.help()
		}
	default:
		err = errors.New("?Invalid command.")
	}

	//重置命令
	this.ftpClientCmd.Name = ""
	this.ftpClientCmd.Params = nil
	return
}

//断开和ftp的连接
func (this *GoFtpClient) disconnect() {
	this.ftpClientCmd.disconnect()
}

//断开和ftp的连接，并且退出客户端程序
func (this *GoFtpClient) quit() {
	this.running = false
	this.disconnect()
}

//输出当前所在远程服务器的目录
func (this *GoFtpClient) pwd() {
	this.ftpClientCmd.pwd()
}
