package goftp

import (
	"bufio"
	"errors"
	"fmt"
	"net"
	"os"
	"os/user"
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
	FCC_LS            string = "ls"
	FCC_DIR           string = "dir"
	FCC_CD            string = "cd"
	FCC_LCD           string = "lcd"
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
		fmt.Println("goftp: Can't lookup host `", this.Host, "'")
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
				var sysUser, _ = user.Current()
				this.ftpClientCmd = GoFtpClientCmd{
					FtpConn:             conn,
					Connected:           true,
					Username:            sysUser.Username,
					DefaultLocalWorkDir: sysUser.HomeDir,
					LocalWorkDir:        sysUser.HomeDir,
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
		cmdStr = strings.Trim(cmdStr, "\r\n")
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
	case FCC_CD:
		this.cwd()
	case FCC_QUIT, FCC_BYE:
		this.quit()
	case FCC_VERSION:
		this.version()
	case FCC_CLOSE, FCC_DISCONNECT:
		this.disconnect()
	case FCC_PWD:
		this.pwd()
	case FCC_LCD:
		this.lcd()
	case FCC_LS:
		this.ls()
	case FCC_QUESTION_MARK, FCC_HELP:
		if len(cmdParams) > 0 {
			this.cmdHelp(cmdParams...)
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

//更改客户端所在远程服务器的目录
func (this *GoFtpClient) cwd() {
	this.ftpClientCmd.cwd()
}

//更改本地工作目录，默认为用户文件目录
func (this *GoFtpClient) lcd() {
	this.ftpClientCmd.lcd()
}

//获取指定目录(无参数时为当前目录)下的文件列表，并可以
//选择性地将结果输出到指定的文件中
func (this *GoFtpClient) ls() {
	this.ftpClientCmd.ls()
}
