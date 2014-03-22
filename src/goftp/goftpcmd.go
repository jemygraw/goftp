package goftp

import (
	"fmt"
	"net"
	"strings"
)

const (
	FC_SUFFIX = "\r\n"
)

const (
	FC_USER = "USER" //USER login_name
	FC_PASS = "PASS" //PASS login_pass
	FC_QUIT = "QUIT" //
)

type GoFtpClientCmd struct {
	Name   string
	Params []string

	FtpConn net.Conn
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
	var sendData = fmt.Sprint(strings.Join(ftpParams, " "), FC_SUFFIX)
	this.FtpConn.Write([]byte(sendData))
}

func (this *GoFtpClientCmd) recvCmdResponse() {
	var recvData []byte = make([]byte, 1024)
	_, err := this.FtpConn.Read(recvData)
	if err == nil {
		fmt.Print(string(recvData))
	} else {
		fmt.Println(err)
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
	}
}
