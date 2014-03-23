package goftp

import (
	"fmt"
	"net"
	"strings"
)

const (
	FC_REQUEST_SUFFIX  string = "\r\n"
	FC_RESPONSE_BUFFER int    = 1024
)

//定义与ftp服务器进行交互的命令，前缀FC表示Ftp Command
const (
	FC_USER string = "USER" //USER login_name
	FC_PASS string = "PASS" //PASS login_pass
	FC_QUIT string = "QUIT" //QUIT
	FC_PWD  string = "PWD"  //PWD
	FC_CWD  string = "CWD"  //CWD remote_dir
)

type GoFtpClientCmd struct {
	Name      string
	Params    []string
	Connected bool

	FtpConn net.Conn

	GoFtpClientHelp
}

func (this *GoFtpClientCmd) welcome() {
	var data []byte = make([]byte, 1024)
	_, err := this.FtpConn.Read(data)
	if err == nil {
		fmt.Print(string(data))
		this.user()
		this.pass()
	} else {
		fmt.Println(err)
	}
}

func (this *GoFtpClientCmd) sendCmdRequest(ftpParams []string) {
	if this.Connected {
		var sendData = fmt.Sprint(strings.Join(ftpParams, " "), FC_REQUEST_SUFFIX)
		this.FtpConn.Write([]byte(sendData))
	} else {
		fmt.Println("Not connected.")
	}
}

func (this *GoFtpClientCmd) recvCmdResponse() {
	if this.Connected {
		var recvData []byte = make([]byte, FC_RESPONSE_BUFFER)
		_, err := this.FtpConn.Read(recvData)
		if err == nil {
			fmt.Print(string(recvData))
		} else {
			fmt.Println(err)
		}
	}
}

func (this *GoFtpClientCmd) user() {
	fmt.Print("Name:")
	var username string
	fmt.Scanln(&username)
	this.sendCmdRequest([]string{FC_USER, username})
	this.recvCmdResponse()
}

func (this *GoFtpClientCmd) pass() {
	fmt.Print("Password:")
	var password string
	fmt.Scanln(&password)
	this.sendCmdRequest([]string{FC_PASS, password})
	this.recvCmdResponse()
}

func (this *GoFtpClientCmd) pwd() {
	this.sendCmdRequest([]string{FC_PWD})
	this.recvCmdResponse()
}

func (this *GoFtpClientCmd) cwd() {
	var paramCount = len(this.Params)
	if paramCount == 0 {
		fmt.Println("(remote-directory)")
		var remoteDir string
		fmt.Scanln(&remoteDir)
		if remoteDir != "" {
			this.sendCmdRequest([]string{FC_CWD, remoteDir})
			this.recvCmdResponse()
		} else {
			this.cmdHelp(this.Name)
		}
	} else if paramCount > 1 {
		this.cmdHelp(this.Name)
	} else {
		this.sendCmdRequest([]string{FC_CWD, this.Params[0]})
		this.recvCmdResponse()
	}
}

func (this *GoFtpClientCmd) disconnect() {
	this.close()
}

func (this *GoFtpClientCmd) close() {
	if this.FtpConn != nil {
		this.sendCmdRequest([]string{FC_QUIT})
		this.recvCmdResponse()

		this.FtpConn = nil
		this.Name = ""
		this.Params = nil
		this.Connected = false
	}
}
