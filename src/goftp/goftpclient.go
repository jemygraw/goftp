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
	FCC_EXIT          string = "exit"
	FCC_VERSION       string = "version"
	FCC_DISCONNECT    string = "disconnect"
	FCC_CLOSE         string = "close"
	FCC_PWD           string = "pwd"
	FCC_LS            string = "ls"
	FCC_DIR           string = "dir"
	FCC_CD            string = "cd"
	FCC_LCD           string = "lcd"
	FCC_OPEN          string = "open"
	FCC_USER          string = "user"

	//这个命令是为了方便学习ftp客户端而加上的，并不是ftp客户端的标准命令
	FCC_USAGE string = "usage"
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

	//根据提供的主机名解析对应的ip地址，一个主机名可能有多个ip地址，
	//尝试依次进行连接，连接成功就不再尝试下一个ip地址
	ips, lookupErr := net.LookupIP(this.Host)
	if lookupErr != nil {
		fmt.Println("ftp: Can't lookup host `", this.Host, "'")
	} else {
		var port = strconv.Itoa(this.Port)
		for _, ip := range ips {
			//尝试连接ftp服务器，并设置连接的等待超时时间
			conn, connErr := net.DialTimeout("tcp", net.JoinHostPort(ip.String(), port),
				time.Duration(DIAL_FTP_SERVER_TIMEOUT_SECONDS)*time.Second)
			if connErr != nil {
				//连接出错了，悲剧，打印错误信息，然后尝试下一个ip地址
				fmt.Println("Trying ", ip, "...")
				fmt.Println("ftp:", connErr.Error())
			} else {
				fmt.Println("Connected to", ip, ".")
				//获取操作系统当前登录用户
				var sysUser, _ = user.Current()
				//设置ftp客户端命令结构体对象信息
				this.ftpClientCmd = GoFtpClientCmd{
					FtpConn:             conn,
					Connected:           true,
					Username:            sysUser.Username,
					DefaultLocalWorkDir: sysUser.HomeDir,
					LocalWorkDir:        sysUser.HomeDir,
				}
				//打印ftp服务器连接回复信息，并提示用户登录
				this.ftpClientCmd.welcome()
				//既然已经找到了能够连接的ip地址，下面的即使有也不去尝试了
				break
			}
		}
		//不管是否连接ftp服务器成功，我们都会进入命令交互模式
		this.EnterPromptMode()
	}
}

//进入命令交互模式
func (this *GoFtpClient) EnterPromptMode() {
	//设置ftp客户端运行状态
	this.running = true
	//在ftp客户端运行状态为true的时候，不断地检测用户输入的交互命令
	//然后解析输入的命令，并执行解析后的命令，执行完，再次等待用户
	//的交互命令
	for this.running {
		fmt.Print("ftp>")

		//这里使用bufio来读取用户的一行交互输入，这里之所以不使用
		//fmt包里面的scan那些函数，是因为用户的交互输入格式为命令
		//然后可能跟上一些参数，中间用空格分开。scan函数没有办法
		//一次读取这些数据，因为scan函数遇到空格就停止了，把剩下
		//的数据作为下一次scan读取的数据
		cmdReader := bufio.NewReader(os.Stdin)
		cmdStr, err := cmdReader.ReadString('\n')
		//这里把读取的数据后面的换行去掉，对于Mac是"\r"，Linux下面
		//是"\n"，Windows下面是"\r\n"，所以为了支持多平台，直接用
		//"\r\n"作为过滤字符
		cmdStr = strings.Trim(cmdStr, "\r\n")

		//如果输入为空，也就是用户直接按Enter键，那么直接等待下次
		//交互命令，否则去解析命令并执行
		if err == nil && cmdStr != "" {
			this.parseCommand(cmdStr)
			err := this.executeCommand()
			if err != nil {
				fmt.Println("ftp:", err.Error())
			}
			//重置cmdStr值
			cmdStr = ""
		}
	}
}

//解析交互命令
func (this *GoFtpClient) parseCommand(cmdStr string) {
	//使用strings包的Fields来分隔命令和参数，因为这个函数
	//是根据空白字符来分隔的
	var parts = strings.Fields(cmdStr)
	if len(parts) > 0 {
		this.ftpClientCmd.Name = parts[0]
		this.ftpClientCmd.Params = parts[1:]
	}
}

//执行交互命令
func (this *GoFtpClient) executeCommand() (err error) {
	//这里其实我们对交互命令的大小写是忽略的，比如你输入
	//LS和ls是表示的一个命令
	var cmdName = strings.ToLower(this.ftpClientCmd.Name)
	var cmdParams = this.ftpClientCmd.Params
	switch cmdName {
	case FCC_CD:
		this.cwd()
	case FCC_QUIT, FCC_BYE, FCC_EXIT:
		this.quit()
	case FCC_VERSION:
		this.version()
	case FCC_CLOSE, FCC_DISCONNECT:
		this.disconnect()
	case FCC_OPEN:
		this.open()
	case FCC_PWD:
		this.pwd()
	case FCC_LCD:
		this.lcd()
	case FCC_LS:
		this.ls()
	case FCC_USER:
		this.user()
	case FCC_USAGE:
		if len(cmdParams) > 0 {
			this.cmdUsage(cmdParams...)
		} else {
			this.cmdHelp(cmdName)
		}
	case FCC_QUESTION_MARK, FCC_HELP:
		if len(cmdParams) > 0 {
			this.cmdHelp(cmdParams...)
		} else {
			this.help()
		}
	default:
		err = errors.New("?Invalid command.")
	}

	//执行完成，重置命令
	this.ftpClientCmd.Name = ""
	this.ftpClientCmd.Params = nil
	return
}

/*
下面的这些函数其实使用了GoFtpClientCmd来代理执行
以后我们可以把这些函数的首字母改为大写的，就能够
导出了，这个包goftp就可以作为一个开源包来使用
*/

//断开和ftp的连接
func (this *GoFtpClient) disconnect() {
	this.ftpClientCmd.disconnect()
}

//断开和ftp的连接，并且退出客户端程序
func (this *GoFtpClient) quit() {
	this.running = false
	this.disconnect()
}

//建立到ftp服务器的连接
func (this *GoFtpClient) open() {
	this.ftpClientCmd.open()
}

//使用交互式的方式验证登录用户名和密码
func (this *GoFtpClient) user() {
	this.ftpClientCmd.user()
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
